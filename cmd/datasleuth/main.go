package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kamalm96/datasleuth/internal/profiler"
	"github.com/kamalm96/datasleuth/internal/report"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "datasleuth",
	Short:   "Fast dataset profiling and validation from the command line",
	Version: version,
	Long: `DataSleuth is a command-line tool for quickly profiling and validating datasets.
Point it at a CSV, Parquet file, or database table to get instant insights
about your data's structure, quality, and statistical properties.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var profileCmd = &cobra.Command{
	Use:   "profile [file|connection_string]",
	Short: "Profile a dataset and generate statistics",
	Long: `Analyze a dataset to generate a comprehensive statistical profile.
This command automatically detects the file type or database connection
and produces statistics including schema info, data types, missing values,
and basic distribution information.`,
	Example: `  datasleuth profile data.csv
  datasleuth profile data.parquet --output-html report.html
  datasleuth profile "postgresql://user:pass@localhost:5432/dbname?table=users"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")
		outputFile, _ := cmd.Flags().GetString("output-file")
		// will be used in future versions
		// sampleSize, _ := cmd.Flags().GetInt("sample")
		verbose, _ := cmd.Flags().GetBool("verbose")

		fmt.Printf("DataSleuth v%s - Fast dataset profiling and validation\n", version)
		fmt.Println("────────────────────────────────────────────────────────────────────────────────")
		fmt.Printf("\n📊 Dataset: %s\n", source)

		startTime := time.Now()

		profile, err := profiler.ProfileDataset(source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error profiling dataset: %v\n", err)
			os.Exit(1)
		}

		elapsedTime := time.Since(startTime)
		fmt.Printf("   Size: %.2f MB\n", float64(profile.FileSize)/(1024*1024))
		fmt.Printf("   Format: %s\n\n", profile.Format)
		fmt.Printf("⏱️  Profile completed in %.2f seconds\n\n", elapsedTime.Seconds())

		switch outputFormat {
		case "terminal":
			report.PrintTerminalReport(profile, verbose)
		case "html":
			htmlFile := outputFile
			if htmlFile == "" {
				htmlFile = fmt.Sprintf("%s_profile.html", profile.Filename)
			}
			if err := report.GenerateHTMLReport(profile, htmlFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating HTML report: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Full HTML report saved to: %s\n", htmlFile)
		case "markdown":
			mdFile := outputFile
			if mdFile == "" {
				mdFile = fmt.Sprintf("%s_profile.md", profile.Filename)
			}
			if err := report.GenerateMarkdownReport(profile, mdFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating Markdown report: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Full Markdown report saved to: %s\n", mdFile)
		case "json":
			jsonFile := outputFile
			if jsonFile == "" {
				jsonFile = fmt.Sprintf("%s_profile.json", profile.Filename)
			}
			if err := report.GenerateJSONReport(profile, jsonFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating JSON report: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Full JSON report saved to: %s\n", jsonFile)
		default:
			fmt.Fprintf(os.Stderr, "Unsupported output format: %s\n", outputFormat)
			os.Exit(1)
		}
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate [file|connection_string]",
	Short: "Validate a dataset against expectations",
	Long: `Check if a dataset meets defined quality expectations.
This command runs validation checks and reports any issues found.
You can use a configuration file to define expectations or rely on
automatically generated expectations from a previous profile.`,
	Example: `  datasleuth validate data.csv
  datasleuth validate data.csv --config validation_rules.yaml
  datasleuth validate data.csv --against baseline.json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		// These will be used in future versions
		// configFile, _ := cmd.Flags().GetString("config")
		// baselineFile, _ := cmd.Flags().GetString("against")

		fmt.Printf("DataSleuth v%s - Fast dataset profiling and validation\n", version)
		fmt.Println("────────────────────────────────────────────────────────────────────────────────")
		fmt.Printf("\nValidating dataset: %s\n", source)

		// This feature will be implemented in a future version
		fmt.Println("\n⚠️ Validation feature is coming soon in a future release.")
	},
}

var compareCmd = &cobra.Command{
	Use:   "compare [file1] [file2]",
	Short: "Compare two datasets and identify differences",
	Long: `Compare two datasets and generate a report of differences.
This command analyzes schema changes, statistical differences,
and data distribution shifts between two versions of a dataset.`,
	Example: `  datasleuth compare old_data.csv new_data.csv
  datasleuth compare old_data.csv new_data.csv --output-html diff_report.html`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source1 := args[0]
		source2 := args[1]
		// Will be used in future versions
		// outputFile, _ := cmd.Flags().GetString("output-file")

		fmt.Printf("DataSleuth v%s - Fast dataset profiling and validation\n", version)
		fmt.Println("────────────────────────────────────────────────────────────────────────────────")
		fmt.Printf("\nComparing datasets:\n  1. %s\n  2. %s\n", source1, source2)

		fmt.Println("\n⚠️ Comparison feature is coming soon in a future release.")
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(compareCmd)

	profileCmd.Flags().StringP("output", "o", "terminal", "Output format: terminal, json, html, markdown")
	profileCmd.Flags().String("output-file", "", "Save the report to a file")
	profileCmd.Flags().IntP("sample", "s", 0, "Use a sample of rows (0 = all rows)")
	profileCmd.Flags().BoolP("verbose", "v", false, "Show detailed information")

	validateCmd.Flags().String("config", "", "Configuration file with validation rules")
	validateCmd.Flags().String("against", "", "Baseline profile to validate against")
	validateCmd.Flags().String("output-file", "", "Save the validation report to a file")

	compareCmd.Flags().String("output-file", "", "Save the comparison report to a file")
	compareCmd.Flags().Bool("schema-only", false, "Compare only schema, not data distributions")
}
