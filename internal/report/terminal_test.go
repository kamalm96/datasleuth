package report

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kamalm96/datasleuth/internal/profiler"
)

func TestPrintTerminalReport(t *testing.T) {
	profile := createTestProfile()

	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintTerminalReport(profile, false)

	w.Close()
	os.Stdout = originalStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedStrings := []string{
		"Dataset Summary",
		"Rows: 1,000",
		"Columns: 3",
		"Missing cells: 50 (1.67%)",
		"Column Overview",
		"NAME",
		"TYPE",
		"MISSING",
		"UNIQUE",
		"STATS",
		"test_str",
		"test_int",
		"test_float",
		"Potential Data Quality Issues",
		"Column 'test_int': Missing values: 2.00%",
		"Recommendations",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

func TestCollectAllIssues(t *testing.T) {
	profile := createTestProfile()

	issues := collectAllIssues(profile)

	if len(issues) != 4 {
		t.Errorf("Expected 4 issues, got %d", len(issues))
	}

	expectedIssues := []string{
		"High overall missing value rate",
		"Column 'test_str': Missing values",
		"Column 'test_int': Missing values",
		"Column 'test_float': Missing values",
	}

	for _, expected := range expectedIssues {
		found := false
		for _, issue := range issues {
			if strings.Contains(issue, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected issue containing '%s'", expected)
		}
	}
}

func TestGenerateRecommendations(t *testing.T) {
	profile := createTestProfile()

	recommendations := generateRecommendations(profile)

	if len(recommendations) < 1 {
		t.Errorf("Expected at least 1 recommendation, got %d", len(recommendations))
		return
	}

	t.Logf("Recommendations generated: %v", recommendations)
}
func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "small_number",
			input:    123,
			expected: "123",
		},
		{
			name:     "thousands",
			input:    1234,
			expected: "1,234",
		},
		{
			name:     "millions",
			input:    1234567,
			expected: "1,234,567",
		},
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatNumber(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func createTestProfile() *profiler.DatasetProfile {
	return &profiler.DatasetProfile{
		Filename:      "test.csv",
		FileSize:      1024 * 1024,
		Format:        "CSV",
		RowCount:      1000,
		ColumnCount:   3,
		MissingCells:  50,
		DuplicateRows: 0,
		QualityIssues: []profiler.QualityIssue{
			{
				Type:        "high_missing_values",
				Description: "High overall missing value rate: 5.00%",
				Severity:    2,
			},
		},
		QualityScore:   85,
		ProcessingTime: 500 * time.Millisecond,
		CreatedAt:      time.Now(),
		Columns: map[string]*profiler.ColumnProfile{
			"test_str": {
				Name:         "test_str",
				DataType:     "string",
				Count:        980,
				MissingCount: 20,
				UniqueCount:  100,
				TopValues: []profiler.ValueCount{
					{Value: "value1", Count: 200},
					{Value: "value2", Count: 180},
					{Value: "value3", Count: 150},
				},
				IsNumeric:     false,
				IsCategorical: true,
				QualityIssues: []profiler.QualityIssue{
					{
						Type:        "missing_values",
						Description: "Missing values: 2.00%",
						Severity:    1,
					},
				},
			},
			"test_int": {
				Name:         "test_int",
				DataType:     "integer",
				Count:        980,
				MissingCount: 20,
				UniqueCount:  50,
				Min:          float64(1),
				Max:          float64(100),
				Mean:         float64(50),
				Median:       float64(50),
				StdDev:       float64(25),
				HistogramBuckets: []profiler.HistogramBucket{
					{LowerBound: 1, UpperBound: 20, Count: 200},
					{LowerBound: 21, UpperBound: 40, Count: 200},
					{LowerBound: 41, UpperBound: 60, Count: 200},
					{LowerBound: 61, UpperBound: 80, Count: 200},
					{LowerBound: 81, UpperBound: 100, Count: 180},
				},
				IsNumeric:     true,
				IsCategorical: false,
				QualityIssues: []profiler.QualityIssue{
					{
						Type:        "missing_values",
						Description: "Missing values: 2.00%",
						Severity:    1,
					},
				},
			},
			"test_float": {
				Name:         "test_float",
				DataType:     "float",
				Count:        990,
				MissingCount: 10,
				UniqueCount:  100,
				Min:          float64(0.1),
				Max:          float64(9.9),
				Mean:         float64(5.0),
				Median:       float64(5.0),
				StdDev:       float64(2.5),
				HistogramBuckets: []profiler.HistogramBucket{
					{LowerBound: 0.1, UpperBound: 2.0, Count: 200},
					{LowerBound: 2.1, UpperBound: 4.0, Count: 200},
					{LowerBound: 4.1, UpperBound: 6.0, Count: 200},
					{LowerBound: 6.1, UpperBound: 8.0, Count: 200},
					{LowerBound: 8.1, UpperBound: 9.9, Count: 190},
				},
				IsNumeric:     true,
				IsCategorical: false,
				QualityIssues: []profiler.QualityIssue{
					{
						Type:        "missing_values",
						Description: "Missing values: 1.00%",
						Severity:    1,
					},
				},
			},
		},
	}
}
