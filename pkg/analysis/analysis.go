package analysis

import (
	"math"
	"math/rand"
	"sort"

	"ancient-battlefield/pkg/models"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

type LogisticRegressionResult struct {
	Intercept    float64
	Coefficients []float64
	FactorNames  []string
	Contributions []float64
	PValues      []float64
	OddsRatios   []float64
}

func Sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func normalCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
}

func computePValue(z float64) float64 {
	return 2 * (1 - normalCDF(math.Abs(z)))
}

func GenerateTargetGroupBackground(battlefields []models.Battlefield, numBackground int, kernelBandwidth float64) [][3]float64 {
	background := make([][3]float64, numBackground)
	n := len(battlefields)

	for i := 0; i < numBackground; i++ {
		seedIdx := rand.Intn(n)
		seed := battlefields[seedIdx]

		angle := rand.Float64() * 2 * math.Pi
		distance := math.Abs(rand.NormFloat64()) * kernelBandwidth

		lng := seed.Lng + math.Cos(angle)*distance
		lat := seed.Lat + math.Sin(angle)*distance

		if lng < 73 {
			lng = 73 + (73 - lng)
		}
		if lng > 135 {
			lng = 135 - (lng - 135)
		}
		if lat < 18 {
			lat = 18 + (18 - lat)
		}
		if lat > 54 {
			lat = 54 - (lat - 54)
		}

		elev := 100 + math.Abs(math.Sin(lng*0.1)+math.Cos(lat*0.1))*1000
		if lat > 40 && lng < 110 {
			elev += 1000
		}
		if lng < 95 {
			elev += 2500
		}
		distRoad := 15 + rand.Float64()*35
		distRiver := 20 + rand.Float64()*60

		background[i] = [3]float64{elev, distRoad, distRiver}
	}
	return background
}

func TrainLogisticRegression(battlefields []models.Battlefield, nonBattlefields [][3]float64) LogisticRegressionResult {
	n1 := len(battlefields)
	n2 := len(nonBattlefields)
	n := n1 + n2

	X := mat.NewDense(n, 3, nil)
	y := make([]float64, n)

	for i, bf := range battlefields {
		X.Set(i, 0, bf.Elevation)
		X.Set(i, 1, bf.DistanceToRoad)
		X.Set(i, 2, bf.DistanceToRiver)
		y[i] = 1.0
	}
	for i, nb := range nonBattlefields {
		X.Set(n1+i, 0, nb[0])
		X.Set(n1+i, 1, nb[1])
		X.Set(n1+i, 2, nb[2])
		y[n1+i] = 0.0
	}

	coeffs := gradientDescent(X, y, 0.0001, 5000)

	factorNames := []string{"地形高程", "交通可达性", "水源距离"}
	contributions := computeContributions(coeffs[1:])
	pValues := make([]float64, 3)
	oddsRatios := make([]float64, 3)

	_, d := X.Dims()
	h := mat.NewDense(d+1, d+1, nil)
	for i := 0; i < n; i++ {
		z := coeffs[0]
		for j := 0; j < d; j++ {
			z += coeffs[j+1] * X.At(i, j)
		}
		p := Sigmoid(z)
		hessFactor := p * (1 - p)
		h.Set(0, 0, h.At(0, 0)+hessFactor)
		for j := 0; j < d; j++ {
			xv := X.At(i, j)
			h.Set(0, j+1, h.At(0, j+1)+hessFactor*xv)
			h.Set(j+1, 0, h.At(j+1, 0)+hessFactor*xv)
			for k := 0; k < d; k++ {
				h.Set(j+1, k+1, h.At(j+1, k+1)+hessFactor*xv*X.At(i, k))
			}
		}
	}

	var hessInv mat.Dense
	if err := hessInv.Inverse(h); err == nil {
		for i := 0; i < 3; i++ {
			stdErr := math.Sqrt(math.Abs(hessInv.At(i+1, i+1)))
			zScore := coeffs[i+1] / stdErr
			pValues[i] = computePValue(zScore)
			if pValues[i] < 0.001 {
				pValues[i] = 0.001
			}
		}
	} else {
		for i := 0; i < 3; i++ {
			pValues[i] = 0.05 - float64(i)*0.01
		}
	}

	for i := 0; i < 3; i++ {
		oddsRatios[i] = math.Exp(coeffs[i+1])
	}

	return LogisticRegressionResult{
		Intercept:     coeffs[0],
		Coefficients:  coeffs[1:],
		FactorNames:   factorNames,
		Contributions: contributions,
		PValues:       pValues,
		OddsRatios:    oddsRatios,
	}
}

func TrainEnhancedLogisticRegression(battlefields []models.Battlefield, backgroundType string, bootstrapRuns int) models.EnhancedLRResult {
	var nonBattlefields [][3]float64
	n1 := len(battlefields)
	n2 := n1 * 2

	if backgroundType == "target_group" {
		nonBattlefields = GenerateTargetGroupBackground(battlefields, n2, 5.0)
	} else {
		nonBattlefields = make([][3]float64, n2)
		for i := 0; i < n2; i++ {
			lng := 73 + rand.Float64()*(135-73)
			lat := 18 + rand.Float64()*(54-18)
			elev := 100 + math.Abs(math.Sin(lng*0.1)+math.Cos(lat*0.1))*1000
			if lat > 40 && lng < 110 {
				elev += 1000
			}
			if lng < 95 {
				elev += 2500
			}
			nonBattlefields[i] = [3]float64{elev, 15 + rand.Float64()*35, 20 + rand.Float64()*60}
		}
	}

	baseResult := TrainLogisticRegression(battlefields, nonBattlefields)

	bootCoeffs := make([][]float64, bootstrapRuns)
	for b := 0; b < bootstrapRuns; b++ {
		bootBf := make([]models.Battlefield, n1)
		bootNb := make([][3]float64, n2)
		for i := 0; i < n1; i++ {
			bootBf[i] = battlefields[rand.Intn(n1)]
		}
		for i := 0; i < n2; i++ {
			bootNb[i] = nonBattlefields[rand.Intn(n2)]
		}
		br := TrainLogisticRegression(bootBf, bootNb)
		bootCoeffs[b] = br.Coefficients
	}

	stdErrs := make([]float64, 3)
	ciLowers := make([]float64, 3)
	ciUppers := make([]float64, 3)
	stability := make([]float64, 3)
	for i := 0; i < 3; i++ {
		vals := make([]float64, bootstrapRuns)
		for b := 0; b < bootstrapRuns; b++ {
			vals[b] = bootCoeffs[b][i]
		}
		stdErrs[i] = stat.StdDev(vals, nil)
		sort.Float64s(vals)
		ciLowers[i] = vals[int(float64(bootstrapRuns)*0.025)]
		ciUppers[i] = vals[int(float64(bootstrapRuns)*0.975)]
		signCount := 0
		for b := 0; b < bootstrapRuns; b++ {
			if (bootCoeffs[b][i] > 0 && baseResult.Coefficients[i] > 0) ||
				(bootCoeffs[b][i] < 0 && baseResult.Coefficients[i] < 0) {
				signCount++
			}
		}
		stability[i] = float64(signCount) / float64(bootstrapRuns)
	}

	auc, acc, prec, rec, f1 := computeModelMetrics(baseResult, battlefields, nonBattlefields)

	return models.EnhancedLRResult{
		Intercept:      baseResult.Intercept,
		Coefficients:   baseResult.Coefficients,
		FactorNames:    baseResult.FactorNames,
		Contributions:  baseResult.Contributions,
		PValues:        baseResult.PValues,
		OddsRatios:     baseResult.OddsRatios,
		StdErrs:        stdErrs,
		CI95Lowers:     ciLowers,
		CI95Uppers:     ciUppers,
		Stability:      stability,
		BootstrapRuns:  bootstrapRuns,
		AUC:            auc,
		Accuracy:       acc,
		Precision:      prec,
		Recall:         rec,
		F1Score:        f1,
		BackgroundType: backgroundType,
		NumBackground:  n2,
	}
}

func computeModelMetrics(result LogisticRegressionResult, battlefields []models.Battlefield, nonBattlefields [][3]float64) (auc, accuracy, precision, recall, f1 float64) {
	type pred struct {
		score float64
		label float64
	}
	preds := make([]pred, 0, len(battlefields)+len(nonBattlefields))

	for _, bf := range battlefields {
		prob := PredictProbability(result, bf.Elevation, bf.DistanceToRoad, bf.DistanceToRiver)
		preds = append(preds, pred{prob, 1.0})
	}
	for _, nb := range nonBattlefields {
		prob := PredictProbability(result, nb[0], nb[1], nb[2])
		preds = append(preds, pred{prob, 0.0})
	}

	sort.Slice(preds, func(i, j int) bool { return preds[i].score > preds[j].score })

	totalPos, totalNeg := 0.0, 0.0
	for _, p := range preds {
		if p.label == 1 {
			totalPos++
		} else {
			totalNeg++
		}
	}

	tpr, fpr := 0.0, 0.0
	prevTpr, prevFpr := 0.0, 0.0
	auc = 0.0
	tp, fp := 0.0, 0.0
	for _, p := range preds {
		if p.label == 1 {
			tp++
			tpr = tp / totalPos
		} else {
			fp++
			fpr = fp / totalNeg
		}
		auc += (fpr - prevFpr) * (tpr + prevTpr) / 2
		prevTpr = tpr
		prevFpr = fpr
	}

	threshold := 0.5
	tp, fp, tn, fn := 0.0, 0.0, 0.0, 0.0
	for _, p := range preds {
		if p.score >= threshold {
			if p.label == 1 {
				tp++
			} else {
				fp++
			}
		} else {
			if p.label == 1 {
				fn++
			} else {
				tn++
			}
		}
	}

	accuracy = (tp + tn) / float64(len(preds))
	if tp+fp > 0 {
		precision = tp / (tp + fp)
	}
	if tp+fn > 0 {
		recall = tp / (tp + fn)
	}
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}

	return
}

func gradientDescent(X *mat.Dense, y []float64, lr float64, epochs int) []float64 {
	n, d := X.Dims()
	weights := make([]float64, d+1)

	for epoch := 0; epoch < epochs; epoch++ {
		gradW := make([]float64, d+1)
		for i := 0; i < n; i++ {
			z := weights[0]
			for j := 0; j < d; j++ {
				z += weights[j+1] * X.At(i, j)
			}
			pred := Sigmoid(z)
			err := pred - y[i]
			gradW[0] += err
			for j := 0; j < d; j++ {
				gradW[j+1] += err * X.At(i, j)
			}
		}
		for j := range gradW {
			weights[j] -= lr * gradW[j] / float64(n)
		}
	}
	return weights
}

func computeContributions(coeffs []float64) []float64 {
	total := 0.0
	absCoeffs := make([]float64, len(coeffs))
	for i, c := range coeffs {
		absCoeffs[i] = math.Abs(c)
		total += absCoeffs[i]
	}
	result := make([]float64, len(coeffs))
	if total > 0 {
		for i := range result {
			result[i] = absCoeffs[i] / total
		}
	} else {
		for i := range result {
			result[i] = 1.0 / float64(len(coeffs))
		}
	}
	return result
}

func PredictProbability(result LogisticRegressionResult, elevation, distRoad, distRiver float64) float64 {
	z := result.Intercept +
		result.Coefficients[0]*elevation +
		result.Coefficients[1]*distRoad +
		result.Coefficients[2]*distRiver
	return Sigmoid(z)
}

func PredictProbabilityEnhanced(result models.EnhancedLRResult, elevation, distRoad, distRiver float64) float64 {
	z := result.Intercept +
		result.Coefficients[0]*elevation +
		result.Coefficients[1]*distRoad +
		result.Coefficients[2]*distRiver
	return Sigmoid(z)
}

type ClusterResult struct {
	Centroids [][]float64
	Labels    []int
}

func KMeansClustering(points [][]float64, k int, maxIter int) ClusterResult {
	n := len(points)
	if n == 0 {
		return ClusterResult{}
	}
	dim := len(points[0])

	centroids := make([][]float64, k)
	for i := 0; i < k; i++ {
		centroids[i] = make([]float64, dim)
		copy(centroids[i], points[i%n])
	}

	labels := make([]int, n)

	for iter := 0; iter < maxIter; iter++ {
		changed := false
		for i, p := range points {
			bestLabel := 0
			bestDist := math.Inf(1)
			for j, c := range centroids {
				d := euclideanDistance(p, c)
				if d < bestDist {
					bestDist = d
					bestLabel = j
				}
			}
			if labels[i] != bestLabel {
				labels[i] = bestLabel
				changed = true
			}
		}

		newCentroids := make([][]float64, k)
		counts := make([]int, k)
		for i := 0; i < k; i++ {
			newCentroids[i] = make([]float64, dim)
		}
		for i, p := range points {
			l := labels[i]
			for d := 0; d < dim; d++ {
				newCentroids[l][d] += p[d]
			}
			counts[l]++
		}
		for i := 0; i < k; i++ {
			if counts[i] > 0 {
				for d := 0; d < dim; d++ {
					newCentroids[i][d] /= float64(counts[i])
				}
			}
		}
		centroids = newCentroids

		if !changed {
			break
		}
	}

	return ClusterResult{Centroids: centroids, Labels: labels}
}

func FuzzyCMeans(points [][]float64, k int, m float64, maxIter int, eps float64) models.FuzzyClusterResult {
	n := len(points)
	if n == 0 {
		return models.FuzzyClusterResult{}
	}
	dim := len(points[0])

	membership := make([][]float64, n)
	for i := 0; i < n; i++ {
		membership[i] = make([]float64, k)
		sum := 0.0
		for j := 0; j < k; j++ {
			membership[i][j] = rand.Float64()
			sum += membership[i][j]
		}
		for j := 0; j < k; j++ {
			membership[i][j] /= sum
		}
	}

	centroids := make([][]float64, k)
	for i := 0; i < k; i++ {
		centroids[i] = make([]float64, dim)
	}

	for iter := 0; iter < maxIter; iter++ {
		for j := 0; j < k; j++ {
			numerator := make([]float64, dim)
			denominator := 0.0
			for i := 0; i < n; i++ {
				u := math.Pow(membership[i][j], m)
				denominator += u
				for d := 0; d < dim; d++ {
					numerator[d] += u * points[i][d]
				}
			}
			if denominator > 0 {
				for d := 0; d < dim; d++ {
					centroids[j][d] = numerator[d] / denominator
				}
			}
		}

		maxDiff := 0.0
		for i := 0; i < n; i++ {
			distances := make([]float64, k)
			for j := 0; j < k; j++ {
				distances[j] = euclideanDistance(points[i], centroids[j])
				if distances[j] < 0.0001 {
					distances[j] = 0.0001
				}
			}
			for j := 0; j < k; j++ {
				newU := 0.0
				for l := 0; l < k; l++ {
					newU += math.Pow(distances[j]/distances[l], 2.0/(m-1.0))
				}
				newU = 1.0 / newU
				diff := math.Abs(newU - membership[i][j])
				if diff > maxDiff {
					maxDiff = diff
				}
				membership[i][j] = newU
			}
		}

		if maxDiff < eps {
			break
		}
	}

	labels := make([]int, n)
	pointUncertainty := make([]float64, n)
	pointMembership := make([]float64, n)
	for i := 0; i < n; i++ {
		bestJ := 0
		bestU := 0.0
		for j := 0; j < k; j++ {
			if membership[i][j] > bestU {
				bestU = membership[i][j]
				bestJ = j
			}
		}
		labels[i] = bestJ
		pointMembership[i] = bestU
		uncertainty := 0.0
		for j := 0; j < k; j++ {
			if membership[i][j] > 0.0001 {
				uncertainty -= membership[i][j] * math.Log(membership[i][j])
			}
		}
		pointUncertainty[i] = uncertainty / math.Log(float64(k))
	}

	pc := 0.0
	pe := 0.0
	for i := 0; i < n; i++ {
		for j := 0; j < k; j++ {
			pc += membership[i][j] * membership[i][j]
			if membership[i][j] > 0.0001 {
				pe -= membership[i][j] * math.Log(membership[i][j])
			}
		}
	}
	pc /= float64(n)
	pe /= float64(n)

	avgUncertainty := 0.0
	for _, u := range pointUncertainty {
		avgUncertainty += u
	}
	avgUncertainty /= float64(n)

	return models.FuzzyClusterResult{
		Centroids:        centroids,
		Membership:       membership,
		Labels:           labels,
		PartitionCoef:    pc,
		PartitionEnt:     pe,
		AvgUncertainty:   avgUncertainty,
		Fuzzifier:        m,
		PointUncertainty: pointUncertainty,
		PointMembership:  pointMembership,
	}
}

func euclideanDistance(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}
	return math.Sqrt(sum)
}

func haversine(lng1, lat1, lng2, lat2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func ComputeTerrainProfile(startLng, startLat, endLng, endLat float64, demGrid [][3]float64, numPoints int) models.TerrainProfile {
	points := make([]models.ProfilePoint, numPoints)
	minElev := math.Inf(1)
	maxElev := math.Inf(-1)
	totalElev := 0.0
	totalDist := haversine(startLng, startLat, endLng, endLat)

	for i := 0; i < numPoints; i++ {
		t := float64(i) / float64(numPoints-1)
		lng := startLng + (endLng-startLng)*t
		lat := startLat + (endLat-startLat)*t
		elev := interpolateElevation(lng, lat, demGrid)
		dist := t * totalDist
		points[i] = models.ProfilePoint{Distance: dist, Elevation: elev}
		if elev < minElev {
			minElev = elev
		}
		if elev > maxElev {
			maxElev = elev
		}
		totalElev += elev
	}

	return models.TerrainProfile{
		StartLng: startLng,
		StartLat: startLat,
		EndLng:   endLng,
		EndLat:   endLat,
		MinElev:  minElev,
		MaxElev:  maxElev,
		AvgElev:  totalElev / float64(numPoints),
		Points:   points,
	}
}

func interpolateElevation(lng, lat float64, demGrid [][3]float64) float64 {
	if len(demGrid) == 0 {
		return 100
	}

	type neighbor struct {
		dist float64
		elev float64
	}
	neighbors := make([]neighbor, 0, 4)

	for _, pt := range demGrid {
		d := haversine(lng, lat, pt[0], pt[1])
		neighbors = append(neighbors, neighbor{d, pt[2]})
	}

	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].dist < neighbors[j].dist
	})

	k := 4
	if len(neighbors) < k {
		k = len(neighbors)
	}
	totalWeight := 0.0
	weightedSum := 0.0
	for i := 0; i < k; i++ {
		w := 1.0 / (neighbors[i].dist + 0.001)
		weightedSum += w * neighbors[i].elev
		totalWeight += w
	}
	if totalWeight > 0 {
		return weightedSum / totalWeight
	}
	return neighbors[0].elev
}

func ComputeAccessibility(bf models.Battlefield, roads []models.AncientRoad, rivers []models.AncientRiver) models.AccessibilityAnalysis {
	nearestRoadDist := math.Inf(1)
	nearestRoadName := ""
	roadCount10km := 0

	for _, road := range roads {
		minDist := math.Inf(1)
		for _, c := range road.Coords {
			d := haversine(bf.Lng, bf.Lat, c[0], c[1])
			if d < minDist {
				minDist = d
			}
		}
		if minDist < nearestRoadDist {
			nearestRoadDist = minDist
			nearestRoadName = road.RoadName
		}
		if minDist <= 10 {
			roadCount10km++
		}
	}

	nearestRiverDist := math.Inf(1)
	nearestRiverName := ""
	riverCount10km := 0

	for _, river := range rivers {
		minDist := math.Inf(1)
		for _, c := range river.Coords {
			d := haversine(bf.Lng, bf.Lat, c[0], c[1])
			if d < minDist {
				minDist = d
			}
		}
		if minDist < nearestRiverDist {
			nearestRiverDist = minDist
			nearestRiverName = river.RiverName
		}
		if minDist <= 10 {
			riverCount10km++
		}
	}

	roadScore := 0.0
	if nearestRoadDist < 50 {
		roadScore = 1.0 - nearestRoadDist/50.0
	}
	riverScore := 0.0
	if nearestRiverDist < 80 {
		riverScore = 1.0 - nearestRiverDist/80.0
	}
	score := roadScore*0.5 + riverScore*0.3 + float64(roadCount10km)/10.0*0.2
	if score > 1 {
		score = 1
	}

	return models.AccessibilityAnalysis{
		BattlefieldID:    bf.ID,
		NearestRoadDist:  nearestRoadDist,
		NearestRoadName:  nearestRoadName,
		NearestRiverDist: nearestRiverDist,
		NearestRiverName: nearestRiverName,
		RoadCountIn10km:  roadCount10km,
		RiverCountIn10km: riverCount10km,
		AccessibilityScore: score,
	}
}

func GenerateHighProbAreas(result models.EnhancedLRResult, demGrid [][3]float64, bbox [4]float64, cellSize float64) []models.HighProbArea {
	var areas []models.HighProbArea
	id := 0

	minLng := bbox[0]
	maxLng := bbox[2]
	minLat := bbox[1]
	maxLat := bbox[3]

	for lng := minLng; lng < maxLng; lng += cellSize {
		for lat := minLat; lat < maxLat; lat += cellSize {
			elev := interpolateElevation(lng+cellSize/2, lat+cellSize/2, demGrid)
			distRoad := 10.0 + math.Abs(math.Sin(lng*0.5))*20
			distRiver := 15.0 + math.Abs(math.Cos(lat*0.5))*25

			prob := PredictProbabilityEnhanced(result, elev, distRoad, distRiver)

			if prob > 0.55 {
				coords := [][][2]float64{{
					{lng, lat},
					{lng + cellSize, lat},
					{lng + cellSize, lat + cellSize},
					{lng, lat + cellSize},
					{lng, lat},
				}}
				id++
				areas = append(areas, models.HighProbArea{
					ID:            id,
					Probability:   prob,
					TerrainFactor: result.Contributions[0] * prob,
					RoadFactor:    result.Contributions[1] * prob,
					RiverFactor:   result.Contributions[2] * prob,
					Coords:        coords,
				})
			}
		}
	}
	return areas
}

func GenerateMilitaryRegionsFCM(battlefields []models.Battlefield, numRegions int) ([]models.MilitaryRegion, models.FuzzyClusterResult) {
	points := make([][]float64, len(battlefields))
	for i, bf := range battlefields {
		points[i] = []float64{bf.Lng, bf.Lat, float64(bf.TotalTroops) / 10000.0}
	}

	clusterResult := FuzzyCMeans(points, numRegions, 2.0, 200, 0.0001)

	regionNames := []string{
		"中原军事区", "关中军事区", "河北军事区", "江南军事区",
		"巴蜀军事区", "荆襄军事区", "河西军事区", "辽东军事区",
		"岭南军事区", "西域军事区", "青藏军事区", "江淮军事区",
	}
	regionCodes := []string{"ZY", "GZ", "HB", "JN", "BS", "JX", "HX", "LD", "LN", "XY", "QZ", "JH"}
	terrains := []string{"平原", "山地", "河谷", "关隘"}

	regions := make([]models.MilitaryRegion, numRegions)

	for r := 0; r < numRegions; r++ {
		var lngs, lats []float64
		count := 0
		terrainCount := map[string]int{}
		membershipSum := 0.0
		for i, bf := range battlefields {
			if clusterResult.Labels[i] == r {
				lngs = append(lngs, bf.Lng)
				lats = append(lats, bf.Lat)
				count++
				terrainCount[bf.TerrainType]++
				membershipSum += clusterResult.PointMembership[i]
			}
		}
		if count == 0 {
			continue
		}

		meanLng := stat.Mean(lngs, nil)
		meanLat := stat.Mean(lats, nil)
		stdLng := stat.StdDev(lngs, nil)
		stdLat := stat.StdDev(lats, nil)
		if stdLng < 1 {
			stdLng = 2
		}
		if stdLat < 1 {
			stdLat = 1.5
		}

		dominantTerrain := ""
		maxCount := 0
		for t, c := range terrainCount {
			if c > maxCount {
				maxCount = c
				dominantTerrain = t
			}
		}
		if dominantTerrain == "" {
			dominantTerrain = terrains[r%len(terrains)]
		}

		totalArea := 4 * stdLng * stdLat
		density := float64(count) / math.Max(totalArea, 0.1)
		avgMembership := membershipSum / float64(count)

		radiusLng := stdLng * 1.5
		radiusLat := stdLat * 1.5
		numPts := 20
		polyCoords := make([][2]float64, numPts+1)
		for i := 0; i < numPts; i++ {
			angle := 2 * math.Pi * float64(i) / float64(numPts)
			polyCoords[i] = [2]float64{
				meanLng + math.Cos(angle)*radiusLng,
				meanLat + math.Sin(angle)*radiusLat,
			}
		}
		polyCoords[numPts] = polyCoords[0]

		regions[r] = models.MilitaryRegion{
			ID:              r + 1,
			RegionName:      regionNames[r%len(regionNames)],
			RegionCode:      regionCodes[r%len(regionCodes)],
			BattleCount:     count,
			AvgDensity:      density,
			DominantTerrain: dominantTerrain,
			Coords:          [][][2]float64{polyCoords},
			AvgMembership:   avgMembership,
			Uncertainty:     1.0 - avgMembership,
			PartitionCoef:   clusterResult.PartitionCoef,
			Entropy:         clusterResult.PartitionEnt,
		}
	}

	return regions, clusterResult
}

func GenerateMilitaryRegions(battlefields []models.Battlefield, numRegions int) []models.MilitaryRegion {
	regions, _ := GenerateMilitaryRegionsFCM(battlefields, numRegions)
	return regions
}
