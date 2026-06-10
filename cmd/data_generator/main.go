package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

const (
	ChinaMinLng = 73.0
	ChinaMaxLng = 135.0
	ChinaMinLat = 18.0
	ChinaMaxLat = 54.0
)

type Battlefield struct {
	ID              int       `json:"id"`
	BattleName      string    `json:"battle_name"`
	Dynasty         string    `json:"dynasty"`
	Era             string    `json:"era"`
	BattleYear      int       `json:"battle_year"`
	BelligerentA    string    `json:"belligerent_a"`
	BelligerentB    string    `json:"belligerent_b"`
	TroopA          int       `json:"troop_a"`
	TroopB          int       `json:"troop_b"`
	TotalTroops     int       `json:"total_troops"`
	TerrainType     string    `json:"terrain_type"`
	Result          string    `json:"result"`
	Lng             float64   `json:"lng"`
	Lat             float64   `json:"lat"`
	Elevation       float64   `json:"elevation"`
	DistanceToRiver float64   `json:"distance_to_river"`
	DistanceToRoad  float64   `json:"distance_to_road"`
}

type AncientRoad struct {
	ID         int         `json:"id"`
	RoadName   string      `json:"road_name"`
	RoadType   string      `json:"road_type"`
	Dynasty    string      `json:"dynasty"`
	Importance int         `json:"importance"`
	Coords     [][2]float64 `json:"coords"`
}

type AncientRiver struct {
	ID        int         `json:"id"`
	RiverName string      `json:"river_name"`
	RiverType string      `json:"river_type"`
	Coords    [][2]float64 `json:"coords"`
}

type DEMCell struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
	Elev float64 `json:"elev"`
}

type FullDataset struct {
	Battlefields []Battlefield  `json:"battlefields"`
	Roads        []AncientRoad  `json:"roads"`
	Rivers       []AncientRiver `json:"rivers"`
	DEMGrid      [][]DEMCell    `json:"dem_grid"`
}

var (
	eras = []struct {
		name     string
		dynasties []string
		yearRange [2]int
	}{
		{"春秋战国", []string{"春秋", "战国"}, [2]int{-770, -221}},
		{"秦汉", []string{"秦", "西汉", "东汉"}, [2]int{-221, 220}},
		{"三国两晋南北朝", []string{"三国", "西晋", "东晋", "南北朝"}, [2]int{220, 589}},
		{"隋唐五代", []string{"隋", "唐", "五代十国"}, [2]int{581, 960}},
		{"宋辽金元", []string{"北宋", "南宋", "辽", "金", "元"}, [2]int{960, 1368}},
		{"明清", []string{"明", "清"}, [2]int{1368, 1912}},
	}

	battleNamePrefixes = []string{"牧野", "长平", "巨鹿", "赤壁", "淝水", "官渡", "夷陵", "街亭", "巨鹿", "雁门", "函谷", "虎牢", "襄阳", "彭城", "垓下", "定军", "祁山", "潼关", "井陉", "马陵", "桂陵", "城濮", "鄢郢", "即墨", "肥下", "番吾", "阙与", "阏与", "蕞", "邯郸"}
	battleNameSuffixes = []string{"之战", "大捷", "保卫战", "攻坚战", "突围战", "伏击战", "遭遇战", "决战", "战役", "会战"}
	factions = []string{"秦", "楚", "齐", "燕", "赵", "魏", "韩", "晋", "吴", "越", "汉", "魏", "蜀", "吴", "隋", "唐", "宋", "辽", "金", "元", "明", "清", "匈奴", "突厥", "吐蕃", "契丹", "女真", "蒙古", "义军", "叛军"}
	terrainTypes = []string{"山地", "平原", "河谷", "关隘"}
	results = []string{"A方胜", "B方胜", "双方议和", "僵持不下"}
	roadTypes = []string{"驿道", "栈道", "漕运", "官道", "古道"}
	riverNames = []string{"黄河", "长江", "淮河", "珠江", "海河", "辽河", "松花江", "汉江", "湘江", "赣江", "岷江", "嘉陵江", "乌江", "渭河", "汾河", "洛河", "伊河", "漳河", "滹沱河", "大运河", "洞庭湖", "鄱阳湖", "太湖", "巢湖", "洪泽湖"}
)

func randRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func randInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func randChoice(arr []string) string {
	return arr[rand.Intn(len(arr))]
}

func generateElevation(lng, lat float64) float64 {
	base := 50.0
	if lat > 40 && lng < 110 {
		base = 1200
	}
	if lat > 30 && lat < 40 && lng < 105 {
		base = 2000
	}
	if lng < 95 {
		base = 3500
	}
	if lat < 25 {
		base = 200
	}
	noise := randRange(-200, 200)
	return math.Max(0, base+noise)
}

func generateBattlefield(id int) Battlefield {
	eraInfo := eras[rand.Intn(len(eras))]
	dynasty := eraInfo.dynasties[rand.Intn(len(eraInfo.dynasties))]
	year := randInt(eraInfo.yearRange[0], eraInfo.yearRange[1])

	lng := randRange(ChinaMinLng, ChinaMaxLng)
	lat := randRange(ChinaMinLat, ChinaMaxLat)

	elevation := generateElevation(lng, lat)
	terrainType := terrainTypes[rand.Intn(len(terrainTypes))]
	if elevation > 1500 {
		terrainType = "山地"
	} else if elevation < 100 && lng > 110 {
		terrainType = "平原"
	}

	troopA := randInt(1000, 500000)
	troopB := randInt(1000, 500000)

	return Battlefield{
		ID:              id,
		BattleName:      randChoice(battleNamePrefixes) + randChoice(battleNameSuffixes),
		Dynasty:         dynasty,
		Era:             eraInfo.name,
		BattleYear:      year,
		BelligerentA:    randChoice(factions),
		BelligerentB:    randChoice(factions),
		TroopA:          troopA,
		TroopB:          troopB,
		TotalTroops:     troopA + troopB,
		TerrainType:     terrainType,
		Result:          randChoice(results),
		Lng:             lng,
		Lat:             lat,
		Elevation:       elevation,
		DistanceToRiver: randRange(0.5, 80),
		DistanceToRoad:  randRange(0.1, 50),
	}
}

func generateRoad(id int) AncientRoad {
	startLng := randRange(ChinaMinLng, ChinaMaxLng)
	startLat := randRange(ChinaMinLat, ChinaMaxLat)
	endLng := startLng + randRange(-10, 10)
	endLat := startLat + randRange(-8, 8)

	numPoints := randInt(5, 15)
	coords := make([][2]float64, numPoints)
	for i := 0; i < numPoints; i++ {
		t := float64(i) / float64(numPoints-1)
		lng := startLng + (endLng-startLng)*t + randRange(-0.5, 0.5)
		lat := startLat + (endLat-startLat)*t + randRange(-0.3, 0.3)
		coords[i] = [2]float64{lng, lat}
	}

	eraInfo := eras[rand.Intn(len(eras))]
	return AncientRoad{
		ID:         id,
		RoadName:   fmt.Sprintf("%s古道%d号线", eraInfo.dynasties[0], id),
		RoadType:   randChoice(roadTypes),
		Dynasty:    eraInfo.dynasties[rand.Intn(len(eraInfo.dynasties))],
		Importance: randInt(1, 5),
		Coords:     coords,
	}
}

func generateRiver(id int) AncientRiver {
	rType := "河流"
	if id > 18 {
		rType = "湖泊"
	}
	if id == 19 {
		rType = "运河"
	}

	startLng := randRange(ChinaMinLng, ChinaMaxLng)
	startLat := randRange(ChinaMinLat, ChinaMaxLat)
	length := randRange(5, 20)
	coords := make([][2]float64, length)
	lng := startLng
	lat := startLat
	for i := 0; i < length; i++ {
		lng += randRange(0.5, 2.5)
		lat += randRange(-0.3, 0.3)
		coords[i] = [2]float64{lng, lat}
	}

	name := riverNames[id%len(riverNames)]
	return AncientRiver{
		ID:        id,
		RiverName: name,
		RiverType: rType,
		Coords:    coords,
	}
}

func generateDEMGrid() [][]DEMCell {
	cols := 50
	rows := 40
	grid := make([][]DEMCell, rows)
	for r := 0; r < rows; r++ {
		grid[r] = make([]DEMCell, cols)
		for c := 0; c < cols; c++ {
			lng := ChinaMinLng + float64(c)*(ChinaMaxLng-ChinaMinLng)/float64(cols-1)
			lat := ChinaMaxLat - float64(r)*(ChinaMaxLat-ChinaMinLat)/float64(rows-1)
			elev := generateElevation(lng, lat)
			grid[r][c] = DEMCell{lng, lat, elev}
		}
	}
	return grid
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("正在生成模拟数据...")

	battlefields := make([]Battlefield, 800)
	for i := 0; i < 800; i++ {
		battlefields[i] = generateBattlefield(i + 1)
	}

	roads := make([]AncientRoad, 60)
	for i := 0; i < 60; i++ {
		roads[i] = generateRoad(i + 1)
	}

	rivers := make([]AncientRiver, 25)
	for i := 0; i < 25; i++ {
		rivers[i] = generateRiver(i + 1)
	}

	demGrid := generateDEMGrid()

	dataset := FullDataset{
		Battlefields: battlefields,
		Roads:        roads,
		Rivers:       rivers,
		DEMGrid:      demGrid,
	}

	webDir := "./web/data"
	if err := os.MkdirAll(webDir, 0755); err != nil {
		fmt.Printf("创建数据目录失败: %v\n", err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil {
		fmt.Printf("JSON序列化失败: %v\n", err)
		os.Exit(1)
	}

	outputPath := webDir + "/data.json"
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("数据生成完成！\n")
	fmt.Printf("  - 战场遗址: %d 个\n", len(battlefields))
	fmt.Printf("  - 古代道路: %d 条\n", len(roads))
	fmt.Printf("  - 河流水系: %d 条\n", len(rivers))
	fmt.Printf("  - DEM栅格: %d × %d = %d 个格点\n", len(demGrid[0]), len(demGrid), len(demGrid)*len(demGrid[0]))
	fmt.Printf("输出文件: %s\n", outputPath)
}
