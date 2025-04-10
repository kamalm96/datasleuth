package report

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/kamalm96/datasleuth/internal/profiler"
)

var (
	titleStyle   = color.New(color.FgHiCyan, color.Bold)
	headerStyle  = color.New(color.FgBlue, color.Bold)
	successStyle = color.New(color.FgGreen, color.Bold)
	warnStyle    = color.New(color.FgYellow)
	errorStyle   = color.New(color.FgRed)
	infoStyle    = color.New(color.FgCyan)
	boldStyle    = color.New(color.Bold)
)

func PrintTerminalReport(profile *profiler.DatasetProfile, verbose bool) {
	qualityColor := successStyle
	if profile.QualityScore < 70 {
		qualityColor = errorStyle
	} else if profile.QualityScore < 90 {
		qualityColor = warnStyle
	}

	titleStyle.Printf("╔════════════════════════════════════════════════════════════╗\n")
	titleStyle.Printf("║ %-58s ║\n", fmt.Sprintf("DATASLEUTH REPORT: %s", profile.Filename))
	titleStyle.Printf("╚════════════════════════════════════════════════════════════╝\n\n")

	fmt.Print("Quality Score: ")
	qualityColor.Printf("%d/100", profile.QualityScore)
	fmt.Println()

	renderQualityBar(profile.QualityScore)
	fmt.Println()

	headerStyle.Println("📋 DATASET SUMMARY")
	fmt.Printf("   ┌ Rows:          %s\n", formatNumber(profile.RowCount))
	fmt.Printf("   ├ Columns:       %d\n", profile.ColumnCount)

	if profile.MissingCells > 0 {
		totalCells := profile.RowCount * profile.ColumnCount
		missingPct := float64(profile.MissingCells) / float64(totalCells) * 100
		fmt.Printf("   ├ Missing cells:  %s (%.2f%%)\n", formatNumber(profile.MissingCells), missingPct)
	} else {
		fmt.Printf("   ├ Missing cells:  0 (0.00%%)\n")
	}

	if profile.DuplicateRows > 0 {
		dupPct := float64(profile.DuplicateRows) / float64(profile.RowCount) * 100
		fmt.Printf("   ├ Duplicate rows: %s (%.2f%%)\n", formatNumber(profile.DuplicateRows), dupPct)
	} else {
		fmt.Printf("   ├ Duplicate rows: 0 (0.00%%)\n")
	}

	fmt.Printf("   └ Processing:    %.2f seconds\n", profile.ProcessingTime.Seconds())
	fmt.Println()

	headerStyle.Println("🔍 COLUMN OVERVIEW")
	fmt.Printf("   ┌─────────────────┬────────────┬──────────┬──────────┬─────────────────────┬───────────┐\n")
	boldStyle.Printf("   │ %-15s │ %-10s │ %-8s │ %-8s │ %-19s │ %-9s │\n", "COLUMN", "TYPE", "MISSING", "UNIQUE", "STATS", "ISSUES")
	fmt.Printf("   ├─────────────────┼────────────┼──────────┼──────────┼─────────────────────┼───────────┤\n")

	for name, col := range profile.Columns {
		colName := name
		if len(colName) > 15 {
			colName = colName[:12] + "..."
		}

		dataType := col.DataType

		var missingStr string
		if profile.RowCount > 0 {
			missingPct := float64(col.MissingCount) / float64(profile.RowCount) * 100
			missingStr = fmt.Sprintf("%.2f%%", missingPct)

			if missingPct > 20 {
				missingStr = errorStyle.Sprintf(missingStr)
			} else if missingPct > 5 {
				missingStr = warnStyle.Sprintf(missingStr)
			}
		} else {
			missingStr = "0.00%"
		}

		var uniqueStr string
		if col.Count > 0 {
			uniquePct := float64(col.UniqueCount) / float64(col.Count) * 100
			uniqueStr = fmt.Sprintf("%.2f%%", uniquePct)
		} else {
			uniqueStr = "0.00%"
		}

		var statsStr string
		if col.IsNumeric {
			statsStr = fmt.Sprintf("mean=%.1f, stddev=%.1f", col.Mean, col.StdDev)
		} else if col.IsDateTime {
			statsStr = "datetime"
		} else if col.IsCategorical && len(col.TopValues) > 0 {
			topValue := col.TopValues[0].Value
			if len(topValue) > 10 {
				topValue = topValue[:7] + "..."
			}
			topPct := float64(col.TopValues[0].Count) / float64(col.Count) * 100
			statsStr = fmt.Sprintf("top: %s (%.1f%%)", topValue, topPct)
		} else if col.IsUnique {
			statsStr = successStyle.Sprint("unique values")
		} else {
			statsStr = "-"
		}

		qualityMark := "✓"
		if len(col.QualityIssues) > 0 {
			severity := 0
			for _, issue := range col.QualityIssues {
				if issue.Severity > severity {
					severity = issue.Severity
				}
			}

			switch severity {
			case 1:
				qualityMark = warnStyle.Sprintf("⚠ Low")
			case 2:
				qualityMark = warnStyle.Sprintf("⚠ Medium")
			case 3:
				qualityMark = errorStyle.Sprintf("⚠ High")
			}
		} else {
			qualityMark = successStyle.Sprint("✓ Good")
		}

		fmt.Printf("   │ %-15s │ %-10s │ %-8s │ %-8s │ %-19s │ %-9s │\n",
			colName, dataType, missingStr, uniqueStr, statsStr, qualityMark)
	}

	fmt.Printf("   └─────────────────┴────────────┴──────────┴──────────┴─────────────────────┴───────────┘\n\n")

	allIssues := collectAllIssues(profile)
	if len(allIssues) > 0 {
		headerStyle.Println("⚠️ QUALITY ISSUES")
		for _, issue := range allIssues {
			fmt.Printf("   • %s\n", issue)
		}
		fmt.Println()
	}

	recommendations := generateRecommendations(profile)
	if len(recommendations) > 0 {
		headerStyle.Println("💡 RECOMMENDATIONS")
		for _, rec := range recommendations {
			fmt.Printf("   • %s\n", rec)
		}
		fmt.Println()
	}

	if verbose {
		headerStyle.Println("📊 COLUMN DETAILS")
		for name, col := range profile.Columns {
			fmt.Printf("\n   %s (%s)\n", boldStyle.Sprint(name), col.DataType)
			fmt.Printf("   ├── Missing: %d (%.2f%%)\n", col.MissingCount, float64(col.MissingCount)/float64(profile.RowCount)*100)
			fmt.Printf("   ├── Unique:  %d (%.2f%%)\n", col.UniqueCount, float64(col.UniqueCount)/float64(col.Count)*100)

			if col.IsNumeric {
				fmt.Printf("   ├── Min:     %v\n", col.Min)
				fmt.Printf("   ├── Max:     %v\n", col.Max)
				fmt.Printf("   ├── Mean:    %.4f\n", col.Mean)
				fmt.Printf("   ├── Median:  %.4f\n", col.Median)
				fmt.Printf("   ├── StdDev:  %.4f\n", col.StdDev)

				if len(col.HistogramBuckets) > 0 {
					fmt.Printf("   └── Histogram:\n\n")
					maxCount := 0
					for _, bucket := range col.HistogramBuckets {
						if bucket.Count > maxCount {
							maxCount = bucket.Count
						}
					}

					maxBarWidth := 40
					for i, bucket := range col.HistogramBuckets {
						barWidth := 0
						if maxCount > 0 {
							barWidth = int(float64(bucket.Count) / float64(maxCount) * float64(maxBarWidth))
						}

						label := fmt.Sprintf("        [%.2f to %.2f]", bucket.LowerBound, bucket.UpperBound)
						bar := strings.Repeat("█", barWidth)

						if i == len(col.HistogramBuckets)-1 {
							fmt.Printf("%s %s %d\n", label, bar, bucket.Count)
						} else {
							fmt.Printf("%s %s %d\n", label, bar, bucket.Count)
						}
					}
				} else {
					fmt.Printf("   └── No histogram available\n")
				}
			} else if col.IsCategorical && len(col.TopValues) > 0 {
				fmt.Printf("   └── Top values:\n")

				maxCount := 0
				for _, val := range col.TopValues {
					if val.Count > maxCount {
						maxCount = val.Count
					}
				}

				maxBarWidth := 30
				for i, val := range col.TopValues {
					barWidth := 0
					if maxCount > 0 {
						barWidth = int(float64(val.Count) / float64(maxCount) * float64(maxBarWidth))
					}

					valuePct := float64(val.Count) / float64(col.Count) * 100
					bar := strings.Repeat("█", barWidth)

					valueStr := val.Value
					if len(valueStr) > 20 {
						valueStr = valueStr[:17] + "..."
					}

					if i == len(col.TopValues)-1 {
						fmt.Printf("        %-20s %s %d (%.2f%%)\n", valueStr, bar, val.Count, valuePct)
					} else {
						fmt.Printf("        %-20s %s %d (%.2f%%)\n", valueStr, bar, val.Count, valuePct)
					}
				}
			} else {
				fmt.Printf("   └── No detailed statistics available\n")
			}

			if len(col.QualityIssues) > 0 {
				fmt.Println("\n   Quality Issues:")
				for _, issue := range col.QualityIssues {
					severityMarker := "⚠️ "
					if issue.Severity == 2 {
						severityMarker = warnStyle.Sprint("⚠️ ")
					} else if issue.Severity == 3 {
						severityMarker = errorStyle.Sprint("⚠️ ")
					}
					fmt.Printf("   %s %s\n", severityMarker, issue.Description)
				}
			}
		}
	}
}

func renderQualityBar(score int) {
	totalBars := 50
	filledBars := totalBars * score / 100

	fmt.Print("   [")

	for i := 0; i < totalBars; i++ {
		if i < filledBars {
			if score >= 90 {
				successStyle.Print("█")
			} else if score >= 70 {
				warnStyle.Print("█")
			} else {
				errorStyle.Print("█")
			}
		} else {
			fmt.Print("░")
		}
	}

	fmt.Print("]")
}

func collectAllIssues(profile *profiler.DatasetProfile) []string {
	issues := make([]string, 0)

	for _, issue := range profile.QualityIssues {
		issues = append(issues, issue.Description)
	}

	for colName, col := range profile.Columns {
		for _, issue := range col.QualityIssues {
			issues = append(issues, fmt.Sprintf("Column '%s': %s", colName, issue.Description))
		}
	}

	return issues
}

func generateRecommendations(profile *profiler.DatasetProfile) []string {
	recommendations := make([]string, 0)

	columnsWithMissing := make([]string, 0)
	for colName, col := range profile.Columns {
		if col.MissingCount > 0 && float64(col.MissingCount)/float64(profile.RowCount) > 0.05 {
			columnsWithMissing = append(columnsWithMissing, colName)
		}
	}

	if len(columnsWithMissing) > 0 {
		if len(columnsWithMissing) <= 3 {
			for _, colName := range columnsWithMissing {
				recommendations = append(recommendations,
					fmt.Sprintf("Consider imputing missing values in '%s' column", colName))
			}
		} else {
			recommendations = append(recommendations,
				"Several columns have high missing value rates and may need imputation")
		}
	}

	columnsWithOutliers := make([]string, 0)
	for colName, col := range profile.Columns {
		for _, issue := range col.QualityIssues {
			if issue.Type == "outliers" {
				columnsWithOutliers = append(columnsWithOutliers, colName)
				break
			}
		}
	}

	if len(columnsWithOutliers) > 0 {
		if len(columnsWithOutliers) <= 3 {
			for _, colName := range columnsWithOutliers {
				recommendations = append(recommendations,
					fmt.Sprintf("Check outliers in '%s' column", colName))
			}
		} else {
			recommendations = append(recommendations,
				"Multiple numeric columns contain outliers")
		}
	}

	for colName, col := range profile.Columns {
		if col.DataType == "string" && !col.IsCategorical && col.UniqueCount > 0 &&
			col.UniqueCount <= 100 && float64(col.UniqueCount)/float64(col.Count) <= 0.2 {
			recommendations = append(recommendations,
				fmt.Sprintf("Column '%s' might benefit from being treated as categorical", colName))
		}
	}

	if profile.DuplicateRows > 0 && float64(profile.DuplicateRows)/float64(profile.RowCount) > 0.01 {
		recommendations = append(recommendations,
			"Dataset contains duplicate rows - consider deduplication")
	}

	if len(recommendations) == 0 && profile.QualityScore < 90 {
		recommendations = append(recommendations,
			"Review columns with quality issues (marked with ⚠️) for potential improvements")
	}

	return recommendations
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	result := ""
	for i := 0; i < len(fmt.Sprintf("%d", n)); i++ {
		if i > 0 && i%3 == 0 {
			result = "," + result
		}
		result = string(fmt.Sprintf("%d", n)[len(fmt.Sprintf("%d", n))-i-1]) + result
	}

	return result
}
