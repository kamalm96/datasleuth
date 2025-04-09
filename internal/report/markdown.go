package report

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kamalm96/datasleuth/internal/profiler"
)

func GenerateMarkdownReport(profile *profiler.DatasetProfile, outputPath string) error {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# DataSleuth Profile: %s\n\n", profile.Filename))
	content.WriteString(fmt.Sprintf("**Generated:** %s | **Size:** %.2f MB | **Rows:** %s | **Columns:** %d\n\n",
		time.Now().Format("January 2, 2006"),
		float64(profile.FileSize)/(1024*1024),
		formatNumber(profile.RowCount),
		profile.ColumnCount))

	content.WriteString(fmt.Sprintf("## Dataset Quality Score: %d/100\n\n", profile.QualityScore))

	content.WriteString("## Dataset Summary\n\n")
	content.WriteString("| Metric | Value |\n")
	content.WriteString("|--------|-------|\n")
	content.WriteString(fmt.Sprintf("| Rows | %s |\n", formatNumber(profile.RowCount)))
	content.WriteString(fmt.Sprintf("| Columns | %d |\n", profile.ColumnCount))

	if profile.MissingCells > 0 {
		totalCells := profile.RowCount * profile.ColumnCount
		missingPct := float64(profile.MissingCells) / float64(totalCells) * 100
		content.WriteString(fmt.Sprintf("| Missing cells | %s (%.2f%%) |\n",
			formatNumber(profile.MissingCells), missingPct))
	} else {
		content.WriteString("| Missing cells | 0 (0.00%) |\n")
	}

	if profile.DuplicateRows > 0 {
		dupPct := float64(profile.DuplicateRows) / float64(profile.RowCount) * 100
		content.WriteString(fmt.Sprintf("| Duplicate rows | %s (%.2f%%) |\n",
			formatNumber(profile.DuplicateRows), dupPct))
	} else {
		content.WriteString("| Duplicate rows | 0 (0.00%) |\n")
	}

	content.WriteString(fmt.Sprintf("| Processing Time | %.2f seconds |\n\n", profile.ProcessingTime.Seconds()))

	issues := collectAllIssues(profile)
	if len(issues) > 0 {
		content.WriteString("## Quality Issues\n\n")
		for _, issue := range issues {
			content.WriteString(fmt.Sprintf("- %s\n", issue))
		}
		content.WriteString("\n")
	}

	recommendations := generateRecommendations(profile)
	if len(recommendations) > 0 {
		content.WriteString("## Recommendations\n\n")
		for _, rec := range recommendations {
			content.WriteString(fmt.Sprintf("- %s\n", rec))
		}
		content.WriteString("\n")
	}

	content.WriteString("## Column Details\n\n")

	for name, col := range profile.Columns {
		content.WriteString(fmt.Sprintf("### %s\n\n", name))
		content.WriteString(fmt.Sprintf("- **Type:** %s\n", col.DataType))

		if profile.RowCount > 0 {
			missingPct := float64(col.MissingCount) / float64(profile.RowCount) * 100
			content.WriteString(fmt.Sprintf("- **Missing:** %.2f%%\n", missingPct))
		}

		if col.Count > 0 {
			uniquePct := float64(col.UniqueCount) / float64(col.Count) * 100
			content.WriteString(fmt.Sprintf("- **Unique:** %.2f%%\n", uniquePct))
		}

		if col.IsNumeric {
			content.WriteString(fmt.Sprintf("- **Range:** %v - %v\n", col.Min, col.Max))
			content.WriteString(fmt.Sprintf("- **Mean:** %.2f\n", col.Mean))
			content.WriteString(fmt.Sprintf("- **Median:** %.2f\n", col.Median))
			content.WriteString(fmt.Sprintf("- **Std Dev:** %.2f\n", col.StdDev))
		}

		content.WriteString("\n")

		if col.IsCategorical && len(col.TopValues) > 0 {
			content.WriteString("**Top Values:**\n\n")
			for _, val := range col.TopValues {
				if col.Count > 0 {
					valPct := float64(val.Count) / float64(col.Count) * 100
					content.WriteString(fmt.Sprintf("- %s: %d (%.2f%%)\n", val.Value, val.Count, valPct))
				} else {
					content.WriteString(fmt.Sprintf("- %s: %d\n", val.Value, val.Count))
				}
			}
			content.WriteString("\n")
		}

		if len(col.QualityIssues) > 0 {
			content.WriteString("**Quality Issues:**\n\n")
			for _, issue := range col.QualityIssues {
				content.WriteString(fmt.Sprintf("- %s\n", issue.Description))
			}
			content.WriteString("\n")
		}
	}

	content.WriteString("---\n")
	content.WriteString("Generated by DataSleuth v0.1.0 - Fast dataset profiling and validation from the command line\n")

	if err := os.WriteFile(outputPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write Markdown report to file: %w", err)
	}

	return nil
}
