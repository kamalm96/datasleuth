package report

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestGenerateHTMLReport(t *testing.T) {
	profile := createTestProfile()

	tempFile, err := os.CreateTemp("", "report_*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	err = GenerateHTMLReport(profile, tempFile.Name())
	if err != nil {
		t.Fatalf("GenerateHTMLReport failed: %v", err)
	}

	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	htmlContent := string(content)

	expectedStrings := []string{
		"<!DOCTYPE html>",
		"<title>DataSleuth Profile: test.csv</title>",
		"Dataset Summary",
		"Quality Score",
		"85/100",
		"test_str",
		"test_int",
		"test_float",
		"Missing cells:",
		"Duplicate rows:",
		"Quality Issues",
		"Recommendations",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(htmlContent, expected) {
			t.Errorf("Expected HTML to contain '%s'", expected)
		}
	}

	if !strings.Contains(htmlContent, "<div class=\"histogram-bar\"") {
		t.Error("Expected HTML to contain histogram bars")
	}
}

func TestFormatNumberHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "integer",
			input:    123,
			expected: "123",
		},
		{
			name:     "int64",
			input:    int64(123456),
			expected: "123456",
		},
		{
			name:     "float_whole",
			input:    123.0,
			expected: "123",
		},
		{
			name:     "float_decimal",
			input:    123.45,
			expected: "123.45",
		},
		{
			name:     "string",
			input:    "test",
			expected: "test",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatNumberHTML(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestFormatPercentHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "zero",
			input:    0,
			expected: "0.00%",
		},
		{
			name:     "small",
			input:    0.0123,
			expected: "1.23%",
		},
		{
			name:     "large",
			input:    0.9999,
			expected: "99.99%",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatPercentHTML(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestFormatDateHTML(t *testing.T) {
	date := time.Date(2023, 4, 15, 0, 0, 0, 0, time.UTC)
	expected := "2023-04-15"

	result := formatDateHTML(date)
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestToJSON(t *testing.T) {
	data := map[string]interface{}{
		"name":  "test",
		"value": 123,
	}

	result := string(toJSON(data))

	if !strings.Contains(result, "\"name\"") || !strings.Contains(result, "\"test\"") {
		t.Errorf("JSON output doesn't contain expected fields: %s", result)
	}

	if !strings.Contains(result, "\"value\"") || !strings.Contains(result, "123") {
		t.Errorf("JSON output doesn't contain expected fields: %s", result)
	}
}

func TestDivideFloat(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected float64
	}{
		{
			name:     "normal_division",
			a:        10,
			b:        2,
			expected: 5.0,
		},
		{
			name:     "fractional_result",
			a:        10,
			b:        3,
			expected: 3.3333333333333335,
		},
		{
			name:     "divide_by_zero",
			a:        10,
			b:        0,
			expected: 0.0,
		},
		{
			name:     "zero_numerator",
			a:        0,
			b:        10,
			expected: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := divideFloat(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Expected %f, got %f", tc.expected, result)
			}
		})
	}
}

func TestMultiplyInts(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "zero_multiply",
			a:        0,
			b:        5,
			expected: 0,
		},
		{
			name:     "positive_multiply",
			a:        3,
			b:        7,
			expected: 21,
		},
		{
			name:     "negative_multiply",
			a:        -3,
			b:        7,
			expected: -21,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := multiplyInts(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Expected %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestCalculatePercentage(t *testing.T) {
	tests := []struct {
		name     string
		part     int
		total    int
		expected float64
	}{
		{
			name:     "zero_total",
			part:     5,
			total:    0,
			expected: 0,
		},
		{
			name:     "zero_part",
			part:     0,
			total:    100,
			expected: 0,
		},
		{
			name:     "half",
			part:     50,
			total:    100,
			expected: 50,
		},
		{
			name:     "full",
			part:     100,
			total:    100,
			expected: 100,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := calculatePercentage(tc.part, tc.total)
			if result != tc.expected {
				t.Errorf("Expected %f, got %f", tc.expected, result)
			}
		})
	}
}
