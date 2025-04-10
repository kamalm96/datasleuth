package profiler

import (
	"math"
	"sort"
	"strconv"
)

type CorrelationMatrix struct {
	Columns  []string
	Values   map[string]map[string]float64
	TopPairs []CorrelationPair
}

type CorrelationPair struct {
	Column1     string
	Column2     string
	Correlation float64
}

func CalculateCorrelationMatrix(profile *DatasetProfile) *CorrelationMatrix {
	numericColumns := []string{}
	numericData := make(map[string][]float64)

	for name, col := range profile.Columns {
		if col.IsNumeric && col.Count > 0 {
			numericColumns = append(numericColumns, name)

			values := reconstructNumericValues(col)
			numericData[name] = values
		}
	}

	if len(numericColumns) < 2 {
		return nil
	}

	sort.Strings(numericColumns)

	matrix := &CorrelationMatrix{
		Columns:  numericColumns,
		Values:   make(map[string]map[string]float64),
		TopPairs: []CorrelationPair{},
	}

	for _, col1 := range numericColumns {
		matrix.Values[col1] = make(map[string]float64)
		for _, col2 := range numericColumns {
			matrix.Values[col1][col2] = 0
		}
	}

	for i, col1 := range numericColumns {
		matrix.Values[col1][col1] = 1.0

		data1 := numericData[col1]

		for j, col2 := range numericColumns {
			if j <= i {
				continue
			}

			data2 := numericData[col2]

			corr := calculatePearsonCorrelation(data1, data2)

			matrix.Values[col1][col2] = corr
			matrix.Values[col2][col1] = corr
		}
	}

	allPairs := []CorrelationPair{}

	for i, col1 := range numericColumns {
		for j, col2 := range numericColumns {
			if j <= i {
				continue
			}

			allPairs = append(allPairs, CorrelationPair{
				Column1:     col1,
				Column2:     col2,
				Correlation: matrix.Values[col1][col2],
			})
		}
	}

	sort.Slice(allPairs, func(i, j int) bool {
		return math.Abs(allPairs[i].Correlation) > math.Abs(allPairs[j].Correlation)
	})

	topLimit := 10
	if len(allPairs) < topLimit {
		topLimit = len(allPairs)
	}

	for i := 0; i < topLimit; i++ {
		if math.Abs(allPairs[i].Correlation) > 0.1 {
			matrix.TopPairs = append(matrix.TopPairs, allPairs[i])
		}
	}

	return matrix
}

func reconstructNumericValues(col *ColumnProfile) []float64 {
	if !col.IsNumeric || len(col.HistogramBuckets) == 0 {
		return []float64{}
	}

	values := []float64{}

	for _, bucket := range col.HistogramBuckets {
		if bucket.Count <= 0 {
			continue
		}

		bucketMidpoint := (bucket.LowerBound + bucket.UpperBound) / 2

		for i := 0; i < bucket.Count; i++ {
			values = append(values, bucketMidpoint)
		}
	}

	if len(col.TopValues) > 0 {
		for _, topValue := range col.TopValues {
			val, err := strconv.ParseFloat(topValue.Value, 64)
			if err == nil {
				for i := 0; i < topValue.Count; i++ {
					values = append(values, val)
				}
			}
		}
	}

	return values
}

func calculatePearsonCorrelation(x, y []float64) float64 {
	n := len(x)
	if n != len(y) || n == 0 {
		return 0
	}

	maxSampleSize := 10000
	if n > maxSampleSize {
		sampled_x := []float64{}
		sampled_y := []float64{}

		step := n / maxSampleSize
		if step < 1 {
			step = 1
		}

		for i := 0; i < n; i += step {
			sampled_x = append(sampled_x, x[i])
			sampled_y = append(sampled_y, y[i])
		}

		x = sampled_x
		y = sampled_y
		n = len(x)
	}

	sumX, sumY, sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0, 0.0

	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	nFloat := float64(n)

	numerator := nFloat*sumXY - sumX*sumY
	denominator := math.Sqrt((nFloat*sumX2 - sumX*sumX) * (nFloat*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	correlation := numerator / denominator

	return correlation
}

func GetCorrelationStrength(correlation float64) string {
	absCorr := math.Abs(correlation)

	if absCorr >= 0.9 {
		return "Very Strong"
	} else if absCorr >= 0.7 {
		return "Strong"
	} else if absCorr >= 0.5 {
		return "Moderate"
	} else if absCorr >= 0.3 {
		return "Weak"
	} else if absCorr >= 0.1 {
		return "Very Weak"
	}

	return "Negligible"
}

func GetCorrelationDirection(correlation float64) string {
	if correlation > 0 {
		return "Positive"
	} else if correlation < 0 {
		return "Negative"
	}
	return "None"
}
