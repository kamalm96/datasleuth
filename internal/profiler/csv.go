package profiler

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func ProfileCSV(filePath string) (*DatasetProfile, error) {
	startTime := time.Now()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	profile := &DatasetProfile{
		Filename:      filepath.Base(filePath),
		FileSize:      fileInfo.Size(),
		Format:        "CSV",
		ColumnCount:   len(header),
		Columns:       make(map[string]*ColumnProfile),
		CreatedAt:     time.Now(),
		QualityIssues: make([]QualityIssue, 0),
	}

	for _, colName := range header {
		profile.Columns[colName] = &ColumnProfile{
			Name:          colName,
			TopValues:     make([]ValueCount, 0),
			QualityIssues: make([]QualityIssue, 0),
		}
	}

	columnValues := make(map[string][]string)
	valueCounts := make(map[string]map[string]int)

	for colName := range profile.Columns {
		columnValues[colName] = make([]string, 0)
		valueCounts[colName] = make(map[string]int)
	}

	rowHashes := make(map[string]int)

	rowCount := 0
	missingCells := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV: %w", err)
		}

		rowCount++

		rowHash := strings.Join(record, "|")
		if _, exists := rowHashes[rowHash]; exists {
			rowHashes[rowHash]++
		} else {
			rowHashes[rowHash] = 1
		}

		for i, value := range record {
			if i >= len(header) {
				continue
			}

			colName := header[i]

			if value == "" {
				profile.Columns[colName].MissingCount++
				missingCells++
				continue
			}

			columnValues[colName] = append(columnValues[colName], value)

			valueCounts[colName][value]++
		}
	}

	duplicateRows := 0
	for _, count := range rowHashes {
		if count > 1 {
			duplicateRows += (count - 1)
		}
	}

	profile.RowCount = rowCount
	profile.MissingCells = missingCells
	profile.DuplicateRows = duplicateRows

	for colName, values := range columnValues {
		col := profile.Columns[colName]
		col.Count = len(values)

		col.DataType = inferDataType(values)
		col.IsNumeric = col.DataType == "integer" || col.DataType == "float"
		col.IsDateTime = col.DataType == "datetime"

		col.UniqueCount = len(valueCounts[colName])
		col.IsCategorical = col.UniqueCount <= profile.RowCount/10 && col.UniqueCount <= 100
		col.IsUnique = col.UniqueCount == col.Count

		col.TopValues = getTopValues(valueCounts[colName], 5)

		if col.IsNumeric {
			calculateNumericStats(col, values)
		}

		detectQualityIssues(col, profile.RowCount)
	}

	collectDatasetQualityIssues(profile)

	profile.QualityScore = CalculateQualityScore(profile)

	profile.ProcessingTime = time.Since(startTime)

	return profile, nil
}

func inferDataType(values []string) string {
	if len(values) == 0 {
		return "unknown"
	}

	sampleSize := int(math.Min(float64(len(values)), 100))

	intCount := 0
	floatCount := 0
	dateCount := 0

	for i := 0; i < sampleSize; i++ {
		if _, err := strconv.ParseInt(values[i], 10, 64); err == nil {
			intCount++
			continue
		}

		if _, err := strconv.ParseFloat(values[i], 64); err == nil {
			floatCount++
			continue
		}

		if _, err := time.Parse(time.RFC3339, values[i]); err == nil {
			dateCount++
			continue
		}

		if _, err := time.Parse("2006-01-02", values[i]); err == nil {
			dateCount++
			continue
		}

		if _, err := time.Parse("01/02/2006", values[i]); err == nil {
			dateCount++
			continue
		}
	}

	if float64(intCount) >= float64(sampleSize)*0.9 {
		return "integer"
	}

	if float64(intCount+floatCount) >= float64(sampleSize)*0.9 {
		return "float"
	}

	if float64(dateCount) >= float64(sampleSize)*0.9 {
		return "datetime"
	}

	return "string"
}

func calculateNumericStats(col *ColumnProfile, values []string) {
	numValues := make([]float64, 0, len(values))

	for _, v := range values {
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			numValues = append(numValues, f)
		}
	}

	if len(numValues) == 0 {
		return
	}

	var sum, sumSquares float64
	var min, max float64 = numValues[0], numValues[0]

	for i, v := range numValues {
		sum += v
		sumSquares += v * v

		if i == 0 || v < min {
			min = v
		}

		if i == 0 || v > max {
			max = v
		}
	}

	n := float64(len(numValues))
	mean := sum / n
	variance := (sumSquares / n) - (mean * mean)
	stdDev := math.Sqrt(variance)

	for i := 1; i < len(numValues); i++ {
		key := numValues[i]
		j := i - 1

		for j >= 0 && numValues[j] > key {
			numValues[j+1] = numValues[j]
			j--
		}

		numValues[j+1] = key
	}

	var median float64
	mid := len(numValues) / 2

	if len(numValues)%2 == 0 {
		median = (numValues[mid-1] + numValues[mid]) / 2
	} else {
		median = numValues[mid]
	}

	bucketCount := 10
	bucketSize := (max - min) / float64(bucketCount)
	buckets := make([]HistogramBucket, bucketCount)

	for i := 0; i < bucketCount; i++ {
		lower := min + float64(i)*bucketSize
		upper := min + float64(i+1)*bucketSize

		if i == bucketCount-1 {
			upper = max
		}

		buckets[i] = HistogramBucket{
			LowerBound: lower,
			UpperBound: upper,
			Count:      0,
		}
	}

	for _, v := range numValues {
		bucketIndex := int((v - min) / bucketSize)
		if bucketIndex >= bucketCount {
			bucketIndex = bucketCount - 1
		}
		buckets[bucketIndex].Count++
	}

	outlierCount := 0
	if stdDev > 0 {
		for _, v := range numValues {
			zScore := math.Abs(v-mean) / stdDev
			if zScore > 3 {
				outlierCount++
			}
		}
	}

	col.Min = min
	col.Max = max
	col.Mean = mean
	col.Median = median
	col.StdDev = stdDev
	col.HistogramBuckets = buckets

	if outlierCount > 0 {
		outlierPct := float64(outlierCount) / float64(len(numValues)) * 100
		severity := 1
		if outlierPct > 5 {
			severity = 2
		}
		if outlierPct > 10 {
			severity = 3
		}

		col.QualityIssues = append(col.QualityIssues, QualityIssue{
			Type:        "outliers",
			Description: fmt.Sprintf("%d outliers detected (%.2f%%)", outlierCount, outlierPct),
			Severity:    severity,
		})
	}
}

func getTopValues(valueCounts map[string]int, limit int) []ValueCount {
	topValues := make([]ValueCount, 0, len(valueCounts))

	for value, count := range valueCounts {
		topValues = append(topValues, ValueCount{Value: value, Count: count})
	}

	for i := 1; i < len(topValues); i++ {
		key := topValues[i]
		j := i - 1

		for j >= 0 && topValues[j].Count < key.Count {
			topValues[j+1] = topValues[j]
			j--
		}

		topValues[j+1] = key
	}

	if len(topValues) > limit {
		topValues = topValues[:limit]
	}

	return topValues
}

func detectQualityIssues(col *ColumnProfile, rowCount int) {
	if col.MissingCount > 0 {
		missingPercentage := float64(col.MissingCount) / float64(rowCount) * 100
		severity := 1

		if missingPercentage > 5 {
			severity = 2
		}
		if missingPercentage > 20 {
			severity = 3
		}

		col.QualityIssues = append(col.QualityIssues, QualityIssue{
			Type:        "missing_values",
			Description: fmt.Sprintf("Missing values: %.2f%%", missingPercentage),
			Severity:    severity,
		})
	}

	if col.UniqueCount == col.Count && strings.Contains(strings.ToLower(col.Name), "id") {
		col.QualityIssues = append(col.QualityIssues, QualityIssue{
			Type:        "likely_id",
			Description: "Likely ID column",
			Severity:    1,
		})
	}

	if col.IsCategorical && len(col.TopValues) > 0 {
		topValuePercentage := float64(col.TopValues[0].Count) / float64(col.Count) * 100
		if topValuePercentage > 90 {
			col.QualityIssues = append(col.QualityIssues, QualityIssue{
				Type:        "imbalanced",
				Description: fmt.Sprintf("Imbalanced: top value appears in %.1f%% of records", topValuePercentage),
				Severity:    2,
			})
		}
	}
}

func collectDatasetQualityIssues(profile *DatasetProfile) {
	if profile.RowCount > 0 && profile.ColumnCount > 0 {
		totalCells := profile.RowCount * profile.ColumnCount
		missingPercentage := float64(profile.MissingCells) / float64(totalCells) * 100

		if missingPercentage > 5 {
			severity := 2
			if missingPercentage > 20 {
				severity = 3
			}

			profile.QualityIssues = append(profile.QualityIssues, QualityIssue{
				Type:        "high_missing_values",
				Description: fmt.Sprintf("High overall missing value rate: %.2f%%", missingPercentage),
				Severity:    severity,
			})
		}
	}

	if profile.RowCount > 0 && profile.DuplicateRows > 0 {
		duplicatePercentage := float64(profile.DuplicateRows) / float64(profile.RowCount) * 100

		severity := 1
		if duplicatePercentage > 5 {

			severity = 2
		}
		if duplicatePercentage > 20 {
			severity = 3
		}

		profile.QualityIssues = append(profile.QualityIssues, QualityIssue{
			Type:        "duplicate_rows",
			Description: fmt.Sprintf("Duplicate rows detected: %.2f%%", duplicatePercentage),
			Severity:    severity,
		})
	}
}
