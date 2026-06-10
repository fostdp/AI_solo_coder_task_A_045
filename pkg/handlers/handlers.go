package handlers

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"

	"ancient-battlefield/pkg/analysis"
	"ancient-battlefield/pkg/models"

	"github.com/gin-gonic/gin"
)

var (
	data       *Dataset
	dataOnce   sync.Once
	dataLoaded bool
)

type Dataset struct {
	Battlefields []models.Battlefield   `json:"battlefields"`
	Roads        []models.AncientRoad   `json:"roads"`
	Rivers       []models.AncientRiver  `json:"rivers"`
	DEMGrid      [][]DEMCell            `json:"dem_grid"`
}

type DEMCell struct {
	Lng  float64 `json:"lng"`
	Lat  float64 `json:"lat"`
	Elev float64 `json:"elev"`
}

func loadData() {
	dataOnce.Do(func() {
		path := "./web/data/data.json"
		if _, err := os.Stat(path); err == nil {
			if content, err := os.ReadFile(path); err == nil {
				var ds Dataset
				if err := json.Unmarshal(content, &ds); err == nil {
					data = &ds
					dataLoaded = true
				}
			}
		}
		if !dataLoaded {
			generateFallbackData()
		}
	})
}

func generateFallbackData() {
	data = &Dataset{
		Battlefields: make([]models.Battlefield, 800),
		Roads:        make([]models.AncientRoad, 60),
		Rivers:       make([]models.AncientRiver, 25),
		DEMGrid:      generateDEM(),
	}
	eras := []string{"春秋战国", "秦汉", "三国两晋南北朝", "隋唐五代", "宋辽金元", "明清"}
	dynasties := []string{"春秋", "战国", "秦", "西汉", "东汉", "三国", "隋", "唐", "北宋", "南宋", "元", "明", "清"}
	terrains := []string{"山地", "平原", "河谷", "关隘"}
	results := []string{"A方胜", "B方胜", "双方议和", "僵持不下"}

	for i := 0; i < 800; i++ {
		lng := 73.0 + rand.Float64()*(135.0-73.0)
		lat := 18.0 + rand.Float64()*(54.0-18.0)
		elev := 100.0 + math.Abs(math.Sin(lng*0.1)+math.Cos(lat*0.1))*1000
		ta := 1000 + rand.Intn(499000)
		tb := 1000 + rand.Intn(499000)
		data.Battlefields[i] = models.Battlefield{
			ID:              i + 1,
			BattleName:      "古战场" + strconv.Itoa(i+1) + "之战",
			Dynasty:         dynasties[rand.Intn(len(dynasties))],
			Era:             eras[rand.Intn(len(eras))],
			BattleYear:      -770 + rand.Intn(2682),
			BelligerentA:    "势力A" + strconv.Itoa(rand.Intn(20)),
			BelligerentB:    "势力B" + strconv.Itoa(rand.Intn(20)),
			TroopA:          ta,
			TroopB:          tb,
			TotalTroops:     ta + tb,
			TerrainType:     terrains[rand.Intn(len(terrains))],
			Result:          results[rand.Intn(len(results))],
			Lng:             lng,
			Lat:             lat,
			Elevation:       elev,
			DistanceToRiver: 1.0 + rand.Float64()*79.0,
			DistanceToRoad:  0.5 + rand.Float64()*49.5,
		}
	}
	for i := 0; i < 60; i++ {
		sl := 80.0 + rand.Float64()*50.0
		sla := 25.0 + rand.Float64()*25.0
		n := 5 + rand.Intn(10)
		coords := make([][2]float64, n)
		for j := 0; j < n; j++ {
			coords[j] = [2]float64{sl + float64(j)*1.5 + rand.Float64()*0.5, sla + rand.Float64()*2 - 1}
		}
		data.Roads[i] = models.AncientRoad{
			ID:         i + 1,
			RoadName:   "古道" + strconv.Itoa(i+1) + "号",
			RoadType:   []string{"驿道", "栈道", "漕运", "官道", "古道"}[rand.Intn(5)],
			Dynasty:    dynasties[rand.Intn(len(dynasties))],
			Importance: 1 + rand.Intn(5),
			Coords:     coords,
		}
	}
	for i := 0; i < 25; i++ {
		sl := 90.0 + rand.Float64()*40.0
		sla := 25.0 + rand.Float64()*20.0
		n := 5 + rand.Intn(15)
		coords := make([][2]float64, n)
		for j := 0; j < n; j++ {
			coords[j] = [2]float64{sl + float64(j)*1.8, sla + rand.Float64()*1.5 - 0.75}
		}
		rType := "河流"
		if i > 18 {
			rType = "湖泊"
		}
		if i == 19 {
			rType = "运河"
		}
		data.Rivers[i] = models.AncientRiver{
			ID:        i + 1,
			RiverName: "河流" + strconv.Itoa(i+1),
			RiverType: rType,
			Coords:    coords,
		}
	}
}

func generateDEM() [][]DEMCell {
	cols, rows := 50, 40
	grid := make([][]DEMCell, rows)
	for r := 0; r < rows; r++ {
		grid[r] = make([]DEMCell, cols)
		for c := 0; c < cols; c++ {
			lng := 73.0 + float64(c)*(135.0-73.0)/float64(cols-1)
			lat := 54.0 - float64(r)*(54.0-18.0)/float64(rows-1)
			elev := 50.0 + math.Abs(math.Sin(lng*0.05)+math.Cos(lat*0.05))*800
			if lat > 40 && lng < 110 {
				elev += 1000
			}
			if lng < 95 {
				elev += 2500
			}
			grid[r][c] = DEMCell{lng, lat, elev}
		}
	}
	return grid
}

func GetBattlefields(c *gin.Context) {
	loadData()
	era := c.Query("era")
	terrain := c.Query("terrain")
	minTroops, _ := strconv.Atoi(c.DefaultQuery("min_troops", "0"))

	result := make([]models.Battlefield, 0)
	for _, bf := range data.Battlefields {
		if era != "" && bf.Era != era {
			continue
		}
		if terrain != "" && bf.TerrainType != terrain {
			continue
		}
		if bf.TotalTroops < minTroops {
			continue
		}
		result = append(result, bf)
	}
	c.JSON(http.StatusOK, result)
}

func GetBattlefieldByID(c *gin.Context) {
	loadData()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	for _, bf := range data.Battlefields {
		if bf.ID == id {
			c.JSON(http.StatusOK, bf)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

func GetRoads(c *gin.Context) {
	loadData()
	c.JSON(http.StatusOK, data.Roads)
}

func GetRivers(c *gin.Context) {
	loadData()
	c.JSON(http.StatusOK, data.Rivers)
}

func GetDEM(c *gin.Context) {
	loadData()
	flat := make([][3]float64, 0)
	for _, row := range data.DEMGrid {
		for _, cell := range row {
			flat = append(flat, [3]float64{cell.Lng, cell.Lat, cell.Elev})
		}
	}
	c.JSON(http.StatusOK, flat)
}

func GetTerrainProfile(c *gin.Context) {
	loadData()
	slng, _ := strconv.ParseFloat(c.Query("start_lng"), 64)
	slat, _ := strconv.ParseFloat(c.Query("start_lat"), 64)
	elng, _ := strconv.ParseFloat(c.Query("end_lng"), 64)
	elat, _ := strconv.ParseFloat(c.Query("end_lat"), 64)
	numPts, _ := strconv.Atoi(c.DefaultQuery("num_points", "50"))

	flat := make([][3]float64, 0)
	for _, row := range data.DEMGrid {
		for _, cell := range row {
			flat = append(flat, [3]float64{cell.Lng, cell.Lat, cell.Elev})
		}
	}
	profile := analysis.ComputeTerrainProfile(slng, slat, elng, elat, flat, numPts)
	c.JSON(http.StatusOK, profile)
}

func GetAccessibility(c *gin.Context) {
	loadData()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	for _, bf := range data.Battlefields {
		if bf.ID == id {
			acc := analysis.ComputeAccessibility(bf, data.Roads, data.Rivers)
			c.JSON(http.StatusOK, acc)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

func GetSiteSelectionFactors(c *gin.Context) {
	loadData()
	nonBattlefields := make([][3]float64, 400)
	for i := 0; i < 400; i++ {
		lng := 73.0 + rand.Float64()*(135.0-73.0)
		lat := 18.0 + rand.Float64()*(54.0-18.0)
		elev := 100.0 + math.Abs(math.Sin(lng*0.1)+math.Cos(lat*0.1))*1000
		nonBattlefields[i] = [3]float64{elev, 15.0 + rand.Float64()*35, 20.0 + rand.Float64()*60}
	}
	result := analysis.TrainLogisticRegression(data.Battlefields, nonBattlefields)

	factors := make([]models.SiteSelectionFactor, 3)
	for i, name := range result.FactorNames {
		factors[i] = models.SiteSelectionFactor{
			ID:           i + 1,
			FactorName:   name,
			Contribution: result.Contributions[i],
			PValue:       result.PValues[i],
			OddsRatio:    result.OddsRatios[i],
			Method:       "逻辑回归",
		}
	}
	c.JSON(http.StatusOK, factors)
}

func GetHighProbAreas(c *gin.Context) {
	loadData()
	nonBattlefields := make([][3]float64, 400)
	for i := 0; i < 400; i++ {
		lng := 73.0 + rand.Float64()*(135.0-73.0)
		lat := 18.0 + rand.Float64()*(54.0-18.0)
		elev := 100.0 + math.Abs(math.Sin(lng*0.1)+math.Cos(lat*0.1))*1000
		nonBattlefields[i] = [3]float64{elev, 15.0 + rand.Float64()*35, 20.0 + rand.Float64()*60}
	}
	lrResult := analysis.TrainLogisticRegression(data.Battlefields, nonBattlefields)

	flat := make([][3]float64, 0)
	for _, row := range data.DEMGrid {
		for _, cell := range row {
			flat = append(flat, [3]float64{cell.Lng, cell.Lat, cell.Elev})
		}
	}

	bbox := [4]float64{73.0, 18.0, 135.0, 54.0}
	cellSize, _ := strconv.ParseFloat(c.DefaultQuery("cell_size", "2.0"), 64)
	areas := analysis.GenerateHighProbAreas(lrResult, flat, bbox, cellSize)
	c.JSON(http.StatusOK, areas)
}

func GetMilitaryRegions(c *gin.Context) {
	loadData()
	numRegions, _ := strconv.Atoi(c.DefaultQuery("num_regions", "8"))
	regions := analysis.GenerateMilitaryRegions(data.Battlefields, numRegions)
	c.JSON(http.StatusOK, regions)
}

func GetStats(c *gin.Context) {
	loadData()
	eraCount := map[string]int{}
	eraTroops := map[string]float64{}
	terrainCount := map[string]int{}
	total := len(data.Battlefields)

	for _, bf := range data.Battlefields {
		eraCount[bf.Era]++
		eraTroops[bf.Era] += float64(bf.TotalTroops)
		terrainCount[bf.TerrainType]++
	}

	statsByEra := make([]models.StatsByEra, 0, len(eraCount))
	for era, cnt := range eraCount {
		statsByEra = append(statsByEra, models.StatsByEra{
			Era:       era,
			Count:     cnt,
			AvgTroops: eraTroops[era] / float64(cnt),
		})
	}

	statsByTerrain := make([]models.StatsByTerrain, 0, len(terrainCount))
	for t, cnt := range terrainCount {
		statsByTerrain = append(statsByTerrain, models.StatsByTerrain{
			TerrainType: t,
			Count:       cnt,
			Percentage:  float64(cnt) / float64(total) * 100,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":          total,
		"stats_by_era":    statsByEra,
		"stats_by_terrain": statsByTerrain,
	})
}
