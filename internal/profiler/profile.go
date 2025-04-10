package profiler

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"
)

type DatasetProfile struct {
	Filename          string
	FileSize          int64
	Format            string
	RowCount          int
	ColumnCount       int
	MissingCells      int
	DuplicateRows     int
	Columns           map[string]*ColumnProfile
	QualityIssues     []QualityIssue
	QualityScore      int
	CorrelationMatrix *CorrelationMatrix
	Recommendations   []string
	ProcessingTime    time.Duration
	CreatedAt         time.Time
}

type ColumnProfile struct {
	Name             string
	DataType         string
	Count            int
	MissingCount     int
	UniqueCount      int
	Min              interface{}
	Max              interface{}
	Mean             float64
	Median           float64
	StdDev           float64
	HistogramBuckets []HistogramBucket
	TopValues        []ValueCount
	IsNumeric        bool
	IsCategorical    bool
	IsDateTime       bool
	IsUnique         bool
	QualityIssues    []QualityIssue
}

type HistogramBucket struct {
	LowerBound float64
	UpperBound float64
	Count      int
}

type ValueCount struct {
	Value string
	Count int
}

type QualityIssue struct {
	Type        string
	Description string
	Severity    int // 1-3 (low to high)
}

func ProfileDataset(filePath string) (*DatasetProfile, error) {
	extension := strings.ToLower(filepath.Ext(filePath))

	startTime := time.Now()

	var profile *DatasetProfile
	var err error

	switch extension {
	case ".csv":
		profile, err = ProfileCSV(filePath)
	case ".parquet":
		profile = &DatasetProfile{
			Filename:  filePath,
			Format:    "Parquet",
			CreatedAt: time.Now(),
			QualityIssues: []QualityIssue{
				{
					Type:        "unsupported_format",
					Description: "Parquet support is coming soon",
					Severity:    2,
				},
			},
		}
	case ".json":
		profile = &DatasetProfile{
			Filename:  filePath,
			Format:    "JSON",
			CreatedAt: time.Now(),
			QualityIssues: []QualityIssue{
				{
					Type:        "unsupported_format",
					Description: "JSON support is coming soon",
					Severity:    2,
				},
			},
		}
	default:
		profile, err = ProfileCSV(filePath)
	}

	if err != nil {
		return nil, err
	}

	// Calculate the quality score
	profile.QualityScore = CalculateQualityScore(profile)

	// Calculate correlations for numeric columns
	profile.CorrelationMatrix = CalculateCorrelationMatrix(profile)

	// Add correlation insights to recommendations
	if profile.CorrelationMatrix != nil && len(profile.CorrelationMatrix.TopPairs) > 0 {
		if profile.Recommendations == nil {
			profile.Recommendations = []string{}
		}

		for _, pair := range profile.CorrelationMatrix.TopPairs {
			if math.Abs(pair.Correlation) >= 0.7 {
				var recommendation string
				if pair.Correlation > 0 {
					recommendation = fmt.Sprintf("Strong positive correlation (%.2f) between '%s' and '%s' - consider if one could be derived from the other",
						pair.Correlation, pair.Column1, pair.Column2)
				} else {
					recommendation = fmt.Sprintf("Strong negative correlation (%.2f) between '%s' and '%s' - these features may provide complementary information",
						pair.Correlation, pair.Column1, pair.Column2)
				}

				profile.Recommendations = append(profile.Recommendations, recommendation)
			}
		}
	}

	profile.ProcessingTime = time.Since(startTime)

	return profile, nil
}

func CalculateQualityScore(profile *DatasetProfile) int {
	score := 100

	if profile.RowCount == 0 || profile.ColumnCount == 0 {
		return 0
	}

	totalCells := profile.RowCount * profile.ColumnCount

	// Deduct for missing values (up to 30 points)
	if totalCells > 0 {
		missingPercentage := float64(profile.MissingCells) / float64(totalCells) * 100
		if missingPercentage > 0 {
			penalty := int(missingPercentage * 3) // 3 points per percent missing
			if penalty > 30 {
				penalty = 30 // Cap at 30 points
			}
			score -= penalty
		}
	}

	// Deduct for quality issues (up to 40 points)
	issuePenalty := 0
	for _, issue := range profile.QualityIssues {
		issuePenalty += issue.Severity * 5 // 5-15 points per issue depending on severity
	}

	// Add column-level issues
	for _, col := range profile.Columns {
		for _, issue := range col.QualityIssues {
			issuePenalty += issue.Severity
		}
	}

	if issuePenalty > 40 {
		issuePenalty = 40 // Cap at 40 points
	}
	score -= issuePenalty

	// Deduct for duplicate rows (up to 15 points)
	if profile.RowCount > 0 {
		duplicatePercentage := float64(profile.DuplicateRows) / float64(profile.RowCount) * 100
		if duplicatePercentage > 0 {
			penalty := int(duplicatePercentage * 2) // 2 points per percent duplicates
			if penalty > 15 {
				penalty = 15 // Cap at 15 points
			}
			score -= penalty
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}
