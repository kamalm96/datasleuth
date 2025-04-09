package report

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGenerateJSONReport(t *testing.T) {
	profile := createTestProfile()

	tempFile, err := os.CreateTemp("", "report_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	err = GenerateJSONReport(profile, tempFile.Name())
	if err != nil {
		t.Fatalf("GenerateJSONReport failed: %v", err)
	}

	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	var jsonReport JSONReport
	err = json.Unmarshal(content, &jsonReport)
	if err != nil {
		t.Fatalf("Failed to parse JSON report: %v", err)
	}

	if jsonReport.Filename != "test.csv" {
		t.Errorf("Expected filename 'test.csv', got '%s'", jsonReport.Filename)
	}

	if jsonReport.RowCount != 1000 {
		t.Errorf("Expected 1000 rows, got %d", jsonReport.RowCount)
	}

	if jsonReport.ColumnCount != 3 {
		t.Errorf("Expected 3 columns, got %d", jsonReport.ColumnCount)
	}

	if jsonReport.QualityScore != 85 {
		t.Errorf("Expected quality score 85, got %d", jsonReport.QualityScore)
	}

	if len(jsonReport.Columns) != 3 {
		t.Errorf("Expected 3 columns in report, got %d", len(jsonReport.Columns))
	}

	strCol, ok := jsonReport.Columns["test_str"]
	if !ok {
		t.Error("Expected test_str column to exist in report")
	} else {
		if strCol.DataType != "string" {
			t.Errorf("Expected test_str to be type 'string', got '%s'", strCol.DataType)
		}

		if strCol.MissingCount != 20 {
			t.Errorf("Expected test_str to have 20 missing values, got %d", strCol.MissingCount)
		}

		if len(strCol.TopValues) == 0 {
			t.Error("Expected test_str to have top values")
		}
	}

	intCol, ok := jsonReport.Columns["test_int"]
	if !ok {
		t.Error("Expected test_int column to exist in report")
	} else {
		if intCol.DataType != "integer" {
			t.Errorf("Expected test_int to be type 'integer', got '%s'", intCol.DataType)
		}

		if intCol.Min != float64(1) {
			t.Errorf("Expected test_int min to be 1, got %v", intCol.Min)
		}

		if intCol.Max != float64(100) {
			t.Errorf("Expected test_int max to be 100, got %v", intCol.Max)
		}

		if intCol.Mean != float64(50) {
			t.Errorf("Expected test_int mean to be 50, got %v", intCol.Mean)
		}

		if len(intCol.Histogram) == 0 {
			t.Error("Expected test_int to have histogram data")
		}
	}

	floatCol, ok := jsonReport.Columns["test_float"]
	if !ok {
		t.Error("Expected test_float column to exist in report")
	} else {
		if floatCol.DataType != "float" {
			t.Errorf("Expected test_float to be type 'float', got '%s'", floatCol.DataType)
		}

		if len(floatCol.QualityIssues) == 0 {
			t.Error("Expected test_float to have quality issues")
		}
	}
}
