package models

type Battlefield struct {
	ID             int      `json:"id"`
	BattleName     string   `json:"battle_name"`
	Dynasty        string   `json:"dynasty"`
	Era            string   `json:"era"`
	BattleYear     int      `json:"battle_year"`
	BelligerentA   string   `json:"belligerent_a"`
	BelligerentB   string   `json:"belligerent_b"`
	TroopA         int      `json:"troop_a"`
	TroopB         int      `json:"troop_b"`
	TotalTroops    int      `json:"total_troops"`
	TerrainType    string   `json:"terrain_type"`
	Result         string   `json:"result"`
	Lng            float64  `json:"lng"`
	Lat            float64  `json:"lat"`
	Elevation      float64  `json:"elevation"`
	DistanceToRiver float64 `json:"distance_to_river"`
	DistanceToRoad  float64 `json:"distance_to_road"`
}

type AncientRoad struct {
	ID         int       `json:"id"`
	RoadName   string    `json:"road_name"`
	RoadType   string    `json:"road_type"`
	Dynasty    string    `json:"dynasty"`
	Importance int       `json:"importance"`
	Coords     [][2]float64 `json:"coords"`
}

type AncientRiver struct {
	ID        int       `json:"id"`
	RiverName string    `json:"river_name"`
	RiverType string    `json:"river_type"`
	Coords    [][2]float64 `json:"coords"`
}

type MilitaryRegion struct {
	ID             int       `json:"id"`
	RegionName     string    `json:"region_name"`
	RegionCode     string    `json:"region_code"`
	BattleCount    int       `json:"battle_count"`
	AvgDensity     float64   `json:"avg_density"`
	DominantTerrain string   `json:"dominant_terrain"`
	Coords         [][][2]float64 `json:"coords"`
	AvgMembership  float64   `json:"avg_membership"`
	Uncertainty   float64   `json:"uncertainty"`
	PartitionCoef float64   `json:"partition_coef"`
	Entropy       float64   `json:"entropy"`
}

type FuzzyClusterResult struct {
	Centroids     [][]float64   `json:"centroids"`
	Membership    [][]float64   `json:"membership"`
	Labels        []int         `json:"labels"`
	PartitionCoef float64       `json:"partition_coef"`
	PartitionEnt  float64       `json:"partition_entropy"`
	AvgUncertainty float64      `json:"avg_uncertainty"`
	Fuzzifier     float64       `json:"fuzzifier"`
	PointUncertainty []float64  `json:"point_uncertainty"`
	PointMembership []float64   `json:"point_membership"`
}

type HighProbArea struct {
	ID            int        `json:"id"`
	Probability   float64    `json:"probability"`
	TerrainFactor float64    `json:"terrain_factor"`
	RoadFactor    float64    `json:"road_factor"`
	RiverFactor   float64    `json:"river_factor"`
	Coords        [][][2]float64 `json:"coords"`
}

type SiteSelectionFactor struct {
	ID             int     `json:"id"`
	FactorName     string  `json:"factor_name"`
	Contribution   float64 `json:"contribution"`
	PValue         float64 `json:"p_value"`
	OddsRatio      float64 `json:"odds_ratio"`
	Method         string  `json:"method"`
	StdErr         float64 `json:"std_err"`
	CI95Lower      float64 `json:"ci95_lower"`
	CI95Upper      float64 `json:"ci95_upper"`
	Significance   bool    `json:"significance"`
	StabilityScore float64 `json:"stability_score"`
}

type EnhancedLRResult struct {
	Intercept     float64   `json:"intercept"`
	Coefficients  []float64 `json:"coefficients"`
	FactorNames   []string  `json:"factor_names"`
	Contributions []float64 `json:"contributions"`
	PValues       []float64 `json:"p_values"`
	OddsRatios    []float64 `json:"odds_ratios"`
	StdErrs       []float64 `json:"std_errs"`
	CI95Lowers    []float64 `json:"ci95_lowers"`
	CI95Uppers    []float64 `json:"ci95_uppers"`
	Stability     []float64 `json:"stability"`
	BootstrapRuns int       `json:"bootstrap_runs"`
	AUC           float64   `json:"auc"`
	Accuracy      float64   `json:"accuracy"`
	Precision     float64   `json:"precision"`
	Recall        float64   `json:"recall"`
	F1Score       float64   `json:"f1_score"`
	BackgroundType string   `json:"background_type"`
	NumBackground int       `json:"num_background"`
}

type ProfilePoint struct {
	Distance  float64 `json:"distance"`
	Elevation float64 `json:"elevation"`
}

type TerrainProfile struct {
	StartLng  float64        `json:"start_lng"`
	StartLat  float64        `json:"start_lat"`
	EndLng    float64        `json:"end_lng"`
	EndLat    float64        `json:"end_lat"`
	MinElev   float64        `json:"min_elev"`
	MaxElev   float64        `json:"max_elev"`
	AvgElev   float64        `json:"avg_elev"`
	Points    []ProfilePoint `json:"points"`
}

type AccessibilityAnalysis struct {
	BattlefieldID   int      `json:"battlefield_id"`
	NearestRoadDist float64  `json:"nearest_road_dist"`
	NearestRoadName string   `json:"nearest_road_name"`
	NearestRiverDist float64 `json:"nearest_river_dist"`
	NearestRiverName string   `json:"nearest_river_name"`
	RoadCountIn10km int      `json:"road_count_in_10km"`
	RiverCountIn10km int     `json:"river_count_in_10km"`
	AccessibilityScore float64 `json:"accessibility_score"`
}

type StatsByEra struct {
	Era      string `json:"era"`
	Count    int    `json:"count"`
	AvgTroops float64 `json:"avg_troops"`
}

type StatsByTerrain struct {
	TerrainType string  `json:"terrain_type"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
}
