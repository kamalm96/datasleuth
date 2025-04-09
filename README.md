# DataSleuth

**Fast dataset profiling and validation from the command line.**

DataSleuth is a lightweight, zero-configuration tool that provides instant insights into datasets. Point it at a CSV file and immediately get a comprehensive profile of your data's structure, quality, and statistical properties - no setup or Python environment required.

![DataSleuth Terminal Output](https://via.placeholder.com/800x400?text=DataSleuth+Terminal+Output)

## Features

- **One-Command Data Profiling**: Get schema info, statistics, and quality checks with a single command
- **Zero Configuration**: No setup, no Python environment, just a single binary
- **Intelligent Data Analysis**: Automatically detects column types, identifies quality issues, and suggests improvements
- **Rich Visual Reports**: Generate HTML reports with histograms and insights
- **Statistical Analysis**: Calculate mean, median, standard deviation, and more for numeric fields
- **Data Quality Checks**: Automatically detect issues like missing values, outliers, and duplicates

## Installation

### Download Binary

Download the pre-built binary for your platform from the [releases page](https://github.com/yourusername/datasleuth/releases).

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/datasleuth.git
cd datasleuth

# Build the binary
go build -o datasleuth ./cmd/datasleuth

# Optional: Move to a location in your PATH
# On Linux/macOS:
sudo mv datasleuth /usr/local/bin/
# On Windows, add the directory to your PATH environment variable
```

## Quick Start

### Running from the Command Line

DataSleuth is a command-line tool, so you need to run it from a terminal:

1. Open Command Prompt (Windows) or Terminal (macOS/Linux)
2. Navigate to the directory containing your data file or the DataSleuth executable
3. Run DataSleuth commands as shown below

### Basic Profile

To generate a basic profile of a CSV file:

```bash
# Windows
datasleuth.exe profile your_data.csv

# macOS/Linux
./datasleuth profile your_data.csv
```

### Generate HTML Report

For a detailed HTML report with visualizations:

```bash
# Windows
datasleuth.exe profile your_data.csv --output html --output-file report.html

# macOS/Linux
./datasleuth profile your_data.csv --output html --output-file report.html
```

To view the HTML report, open the generated file (e.g., `report.html`) in any web browser:
- Double-click the file in your file explorer
- Right-click and select "Open with" your preferred browser
- From your browser, use File > Open to navigate to the file

## Usage

```
DataSleuth is a command-line tool for quickly profiling and validating datasets.
Point it at a CSV file to get instant insights about your data's structure, 
quality, and statistical properties.

Usage:
  datasleuth [command]

Available Commands:
  profile     Profile a dataset and generate statistics
  validate    Validate a dataset against expectations (coming soon)
  compare     Compare two datasets and identify differences (coming soon)
  help        Help about any command

Flags:
  -h, --help      help for datasleuth
  -v, --version   version for datasleuth
```

### Profile Command

```
Analyze a dataset to generate a comprehensive statistical profile.
This command automatically detects the file type or database connection
and produces statistics including schema info, data types, missing values,
and basic distribution information.

Usage:
  datasleuth profile [file] [flags]

Examples:
  datasleuth profile data.csv
  datasleuth profile data.csv --output html --output-file report.html

Flags:
  -h, --help                help for profile
  -o, --output string       Output format: terminal, json, html, markdown (default "terminal")
      --output-file string  Save the report to a file
  -s, --sample int          Use a sample of rows (0 = all rows)
  -v, --verbose             Show detailed information
```

## Understanding the Report

DataSleuth generates comprehensive insights about your data:

### Terminal Output

The terminal output includes:
- **Dataset Summary**: Row count, column count, missing cells, and duplicates
- **Column Overview**: Data types, missing value percentages, unique values, and basic statistics
- **Quality Issues**: Potential data problems like outliers or high missing value rates
- **Recommendations**: Actionable suggestions to improve data quality

### HTML Report

The HTML report provides all the above plus:
- **Quality Score**: An overall assessment of your dataset's quality
- **Interactive Visualizations**: Histograms for numeric columns
- **Detailed Column Stats**: Complete statistical breakdown of each column
- **Categorical Distributions**: Frequency analysis of categorical fields

## Understanding Quality Issues

DataSleuth identifies several types of quality issues:

- **Missing Values**: Fields with empty or null values
- **Outliers**: Values that deviate significantly from the column's distribution
- **Duplicate Rows**: Identical records in the dataset
- **Imbalanced Categories**: Categorical fields dominated by one value
- **ID Columns**: Fields that likely contain unique identifiers

Each issue includes a severity assessment to help prioritize data cleaning efforts.

## Roadmap

- [ ] Support for more file formats (Parquet, JSON)
- [ ] Database connections (PostgreSQL, MySQL)
- [ ] Dataset validation against rules
- [ ] Dataset comparison and drift detection
- [ ] Custom rule definitions

## Troubleshooting

### "Select an app to open" Dialog

If you see a "Select an app to open" dialog when trying to run DataSleuth:
- You're trying to open the executable as a file rather than running it
- Make sure to run DataSleuth from the command line (Command Prompt or Terminal)

### Error Messages

If you encounter errors:
1. Ensure you're using the correct syntax: `datasleuth profile your_file.csv`
2. Check that your CSV file exists and is accessible
3. Verify the file is properly formatted CSV

### Large Files

For very large files:
- Use the sampling option to analyze a subset: `--sample 10000`
- Expect longer processing times for complete analysis

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.