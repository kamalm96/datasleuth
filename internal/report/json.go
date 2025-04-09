package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kamalm96/datasleuth/internal/profiler"
)

type JSONReport struct {
	Filename        string                      `json:"filename"`
	FileSize        int64                       `json:"file_size_bytes"`
	Format          string                      `json:"format"`
	RowCount        int                         `json:"row_count"`
	ColumnCount     int                         `json:"column_count"`
	MissingCells    int                         `json:"missing_cells"`
	DuplicateRows   int                         `json:"duplicate_rows"`
	QualityScore    int                         `json:"quality_score"`
	QualityIssues   []string                    `json:"quality_issues"`
	Recommendations []string                    `json:"recommendations"`
	Columns         map[string]JSONColumnReport `json:"columns"`
	ProcessingTime  float64                     `json:"processing_time_seconds"`
	GeneratedAt     string                      `json:"generated_at"`
}

type JSONColumnReport struct {
	Name           string      `json:"name"`
	DataType       string      `json:"data_type"`
	Count          int         `json:"count"`
	MissingCount   int         `json:"missing_count"`
	MissingPercent float64     `json:"missing_percent"`
	UniqueCount    int         `json:"unique_count"`
	UniquePercent  float64     `json:"unique_percent"`
	Min            interface{} `json:"min,omitempty"`
	Max            interface{} `json:"max,omitempty"`
	Mean           float64     `json:"mean,omitempty"`
	Median         float64     `json:"median,omitempty"`
	StdDev         float64     `json:"std_dev,omitempty"`
	TopValues      []TopValue  `json:"top_values,omitempty"`
	Histogram      []Bucket    `json:"histogram,omitempty"`
	QualityIssues  []string    `json:"quality_issues"`
}

type TopValue struct {
	Value   string  `json:"value"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

type Bucket struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Count int     `json:"count"`
}

func GenerateJSONReport(profile *profiler.DatasetProfile, outputPath string) error {
	report := JSONReport{
		Filename:        profile.Filename,
		FileSize:        profile.FileSize,
		Format:          profile.Format,
		RowCount:        profile.RowCount,
		ColumnCount:     profile.ColumnCount,
		MissingCells:    profile.MissingCells,
		DuplicateRows:   profile.DuplicateRows,
		QualityScore:    profile.QualityScore,
		QualityIssues:   collectAllIssues(profile),
		Recommendations: generateRecommendations(profile),
		Columns:         make(map[string]JSONColumnReport),
		ProcessingTime:  profile.ProcessingTime.Seconds(),
		GeneratedAt:     time.Now().Format(time.RFC3339),
	}

	for name, col := range profile.Columns {
		jsonCol := JSONColumnReport{
			Name:          name,
			DataType:      col.DataType,
			Count:         col.Count,
			MissingCount:  col.MissingCount,
			UniqueCount:   col.UniqueCount,
			QualityIssues: make([]string, 0),
		}

		if profile.RowCount > 0 {
			jsonCol.MissingPercent = float64(col.MissingCount) / float64(profile.RowCount) * 100
		}

		if col.Count > 0 {
			jsonCol.UniquePercent = float64(col.UniqueCount) / float64(col.Count) * 100
		}

		if col.IsNumeric {
			jsonCol.Min = col.Min
			jsonCol.Max = col.Max
			jsonCol.Mean = col.Mean
			jsonCol.Median = col.Median
			jsonCol.StdDev = col.StdDev

			if len(col.HistogramBuckets) > 0 {
				jsonCol.Histogram = make([]Bucket, len(col.HistogramBuckets))
				for i, bucket := range col.HistogramBuckets {
					jsonCol.Histogram[i] = Bucket{
						Min:   bucket.LowerBound,
						Max:   bucket.UpperBound,
						Count: bucket.Count,
					}
				}
			}
		}

		if len(col.TopValues) > 0 {
			jsonCol.TopValues = make([]TopValue, len(col.TopValues))
			for i, val := range col.TopValues {
				percent := 0.0
				if col.Count > 0 {
					percent = float64(val.Count) / float64(col.Count) * 100
				}

				jsonCol.TopValues[i] = TopValue{
					Value:   val.Value,
					Count:   val.Count,
					Percent: percent,
				}
			}
		}

		for _, issue := range col.QualityIssues {
			jsonCol.QualityIssues = append(jsonCol.QualityIssues, issue.Description)
		}

		report.Columns[name] = jsonCol
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report to file: %w", err)
	}

	return nil
}
