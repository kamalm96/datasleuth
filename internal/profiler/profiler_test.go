package profiler

import (
	"os"
	"testing"
)

func TestCalculateQualityScore(t *testing.T) {
	tests := []struct {
		name          string
		profile       *DatasetProfile
		expectedScore int
	}{
		{
			name: "perfect_score",
			profile: &DatasetProfile{
				RowCount:      1000,
				ColumnCount:   10,
				MissingCells:  0,
				DuplicateRows: 0,
				QualityIssues: []QualityIssue{},
				Columns:       map[string]*ColumnProfile{},
			},
			expectedScore: 100,
		},
		{
			name: "missing_values_penalty",
			profile: &DatasetProfile{
				RowCount:      1000,
				ColumnCount:   10,
				MissingCells:  1000, // 10% missing
				DuplicateRows: 0,
				QualityIssues: []QualityIssue{},
				Columns:       map[string]*ColumnProfile{},
			},
			expectedScore: 70, // 30 point penalty for 10% missing
		},
		{
			name: "quality_issues_penalty",
			profile: &DatasetProfile{
				RowCount:      1000,
				ColumnCount:   10,
				MissingCells:  0,
				DuplicateRows: 0,
				QualityIssues: []QualityIssue{
					{Type: "test_issue", Description: "Test issue", Severity: 3},
					{Type: "test_issue", Description: "Test issue", Severity: 2},
				},
				Columns: map[string]*ColumnProfile{},
			},
			expectedScore: 75, // 25 point penalty for issues
		},
		{
			name: "duplicate_rows_penalty",
			profile: &DatasetProfile{
				RowCount:      1000,
				ColumnCount:   10,
				MissingCells:  0,
				DuplicateRows: 100, // 10% duplicates
				QualityIssues: []QualityIssue{},
				Columns:       map[string]*ColumnProfile{},
			},
			expectedScore: 85, // 15 point penalty for 10% duplicates
		},
		{
			name: "empty_dataset",
			profile: &DatasetProfile{
				RowCount:      0,
				ColumnCount:   0,
				MissingCells:  0,
				DuplicateRows: 0,
				QualityIssues: []QualityIssue{},
				Columns:       map[string]*ColumnProfile{},
			},
			expectedScore: 0, // Empty dataset scores 0
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := CalculateQualityScore(tc.profile)
			if score != tc.expectedScore {
				t.Errorf("Expected score %d, got %d", tc.expectedScore, score)
			}
		})
	}
}

func TestProfileDataset_NonExistentFile(t *testing.T) {
	_, err := ProfileDataset("non_existent_file.csv")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestProfileDataset_EmptyFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "empty_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	_, err = ProfileDataset(tempFile.Name())
	if err == nil {
		t.Error("Expected error for empty file, got nil")
	}
}
