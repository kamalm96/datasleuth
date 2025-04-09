package profiler

import (
	"os"
	"testing"
)

func TestProfileCSV(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	csvContent := `col_str,col_int,col_float,col_date,col_missing
value1,1,1.1,2023-01-01,
value2,2,2.2,2023-01-02,something
value3,3,3.3,2023-01-03,
value3,3,3.3,2023-01-03,value
,4,4.4,2023-01-04,
`
	if _, err := tempFile.Write([]byte(csvContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	profile, err := ProfileCSV(tempFile.Name())
	if err != nil {
		t.Fatalf("ProfileCSV failed: %v", err)
	}

	if profile.RowCount != 5 {
		t.Errorf("Expected 5 rows, got %d", profile.RowCount)
	}

	if profile.ColumnCount != 5 {
		t.Errorf("Expected 5 columns, got %d", profile.ColumnCount)
	}

	if profile.MissingCells != 4 {
		t.Errorf("Expected 4 missing cells, got %d", profile.MissingCells)
	}

	if profile.DuplicateRows != 0 {
		t.Errorf("Expected 0 duplicate rows, got %d", profile.DuplicateRows)
	}

	col, exists := profile.Columns["col_int"]
	if !exists {
		t.Fatal("Expected column 'col_int' to exist")
	}

	if col.DataType != "integer" {
		t.Errorf("Expected col_int to be 'integer', got '%s'", col.DataType)
	}

	col, exists = profile.Columns["col_float"]
	if !exists {
		t.Fatal("Expected column 'col_float' to exist")
	}

	if col.DataType != "float" {
		t.Errorf("Expected col_float to be 'float', got '%s'", col.DataType)
	}

	col, exists = profile.Columns["col_str"]
	if !exists {
		t.Fatal("Expected column 'col_str' to exist")
	}

	if col.DataType != "string" {
		t.Errorf("Expected col_str to be 'string', got '%s'", col.DataType)
	}

	col, exists = profile.Columns["col_date"]
	if !exists {
		t.Fatal("Expected column 'col_date' to exist")
	}

	if col.DataType != "datetime" {
		t.Errorf("Expected col_date to be 'datetime', got '%s'", col.DataType)
	}

	col, exists = profile.Columns["col_missing"]
	if !exists {
		t.Fatal("Expected column 'col_missing' to exist")
	}

	if col.MissingCount != 3 {
		t.Errorf("Expected col_missing to have 3 missing values, got %d", col.MissingCount)
	}
}

func TestInferDataType(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{
			name:     "empty_values",
			values:   []string{},
			expected: "unknown",
		},
		{
			name:     "integers",
			values:   []string{"1", "2", "3", "4", "5"},
			expected: "integer",
		},
		{
			name:     "floats",
			values:   []string{"1.1", "2.2", "3.3", "4", "5.5"},
			expected: "float",
		},
		{
			name:     "dates_iso",
			values:   []string{"2023-01-01", "2023-01-02", "2023-01-03"},
			expected: "datetime",
		},
		{
			name:     "dates_us",
			values:   []string{"01/01/2023", "01/02/2023", "01/03/2023"},
			expected: "datetime",
		},
		{
			name:     "strings",
			values:   []string{"abc", "def", "ghi", "jkl"},
			expected: "string",
		},
		{
			name:     "mixed",
			values:   []string{"abc", "123", "def", "456"},
			expected: "string",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dataType := inferDataType(tc.values)
			if dataType != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, dataType)
			}
		})
	}
}

func TestCalculateNumericStats(t *testing.T) {
	col := &ColumnProfile{
		Name:             "test_col",
		DataType:         "integer",
		IsNumeric:        true,
		HistogramBuckets: []HistogramBucket{},
		QualityIssues:    []QualityIssue{},
	}

	values := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	calculateNumericStats(col, values)

	if col.Min.(float64) != 1 {
		t.Errorf("Expected min to be 1, got %v", col.Min)
	}

	if col.Max.(float64) != 10 {
		t.Errorf("Expected max to be 10, got %v", col.Max)
	}

	if col.Mean != 5.5 {
		t.Errorf("Expected mean to be 5.5, got %v", col.Mean)
	}

	if col.Median != 5.5 {
		t.Errorf("Expected median to be 5.5, got %v", col.Median)
	}

	if len(col.HistogramBuckets) != 10 {
		t.Errorf("Expected 10 histogram buckets, got %d", len(col.HistogramBuckets))
	}
}

func TestGetTopValues(t *testing.T) {
	valueCounts := map[string]int{
		"a": 10,
		"b": 5,
		"c": 3,
		"d": 2,
		"e": 1,
	}

	limit := 3
	topValues := getTopValues(valueCounts, limit)

	if len(topValues) != limit {
		t.Errorf("Expected %d top values, got %d", limit, len(topValues))
	}

	if topValues[0].Value != "a" || topValues[0].Count != 10 {
		t.Errorf("Expected top value to be 'a' with count 10, got '%s' with count %d",
			topValues[0].Value, topValues[0].Count)
	}

	if topValues[1].Value != "b" || topValues[1].Count != 5 {
		t.Errorf("Expected second value to be 'b' with count 5, got '%s' with count %d",
			topValues[1].Value, topValues[1].Count)
	}

	if topValues[2].Value != "c" || topValues[2].Count != 3 {
		t.Errorf("Expected third value to be 'c' with count 3, got '%s' with count %d",
			topValues[2].Value, topValues[2].Count)
	}
}

func TestDetectQualityIssues(t *testing.T) {
	col := &ColumnProfile{
		Name:          "test_col",
		MissingCount:  50,
		QualityIssues: []QualityIssue{},
	}

	rowCount := 100
	detectQualityIssues(col, rowCount)

	if len(col.QualityIssues) != 1 {
		t.Errorf("Expected 1 quality issue, got %d", len(col.QualityIssues))
	}

	if col.QualityIssues[0].Type != "missing_values" {
		t.Errorf("Expected 'missing_values' issue, got '%s'", col.QualityIssues[0].Type)
	}

	if col.QualityIssues[0].Severity != 3 {
		t.Errorf("Expected severity 3, got %d", col.QualityIssues[0].Severity)
	}
}
