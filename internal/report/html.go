package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/kamalm96/datasleuth/internal/profiler"
)

type HTMLTemplateData struct {
	Profile         *profiler.DatasetProfile
	GeneratedAt     string
	Issues          []string
	Recommendations []string
	FileSizeMB      float64
}

func GenerateHTMLReport(profile *profiler.DatasetProfile, outputPath string) error {
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"formatNumber":  formatNumberHTML,
		"formatPercent": formatPercentHTML,
		"formatDate":    formatDateHTML,
		"toJSON":        toJSON,
		"div":           divideFloat,
		"mul":           multiplyInts,
		"percentage":    calculatePercentage,
	}).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	fileSizeMB := float64(profile.FileSize) / 1048576.0

	data := HTMLTemplateData{
		Profile:         profile,
		GeneratedAt:     time.Now().Format("January 2, 2006 15:04:05"),
		Issues:          collectAllIssues(profile),
		Recommendations: generateRecommendations(profile),
		FileSizeMB:      fileSizeMB,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write HTML report to file: %w", err)
	}

	return nil
}

func formatNumberHTML(n interface{}) string {
	switch v := n.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		if v == float64(int(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%.2f", v)
	default:
		return fmt.Sprintf("%v", n)
	}
}

func formatPercentHTML(n float64) string {
	return fmt.Sprintf("%.2f%%", n*100)
}

func formatDateHTML(t time.Time) string {
	return t.Format("2006-01-02")
}

func toJSON(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return template.JS("{}")
	}
	return template.JS(b)
}

func divideFloat(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}

func calculatePercentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func multiplyInts(a, b int) int {
	return a * b
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DataSleuth Profile: {{.Profile.Filename}}</title>
    <style>
        :root {
            --primary-color: #1a73e8;
            --secondary-color: #5f6368;
            --background-color: #f8f9fa;
            --card-color: #ffffff;
            --border-color: #dadce0;
            --text-color: #202124;
            --success-color: #0f9d58;
            --warning-color: #f4b400;
            --error-color: #d93025;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
            line-height: 1.6;
            color: var(--text-color);
            background-color: var(--background-color);
            margin: 0;
            padding: 20px;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        
        header {
            background-color: var(--primary-color);
            color: white;
            padding: 20px;
            border-radius: 8px 8px 0 0;
        }
        
        h1, h2, h3 {
            margin-top: 0;
        }
        
        .summary-cards {
            display: flex;
            flex-wrap: wrap;
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .card {
            background-color: var(--card-color);
            border-radius: 8px;
            box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
            padding: 20px;
            flex: 1;
            min-width: 250px;
        }
        
        .metric {
            font-size: 2em;
            font-weight: bold;
            color: var(--primary-color);
        }
        
        .column-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(500px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        
        .column-card {
            background-color: var(--card-color);
            border-radius: 8px;
            box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
            padding: 20px;
        }
        
        table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 20px;
        }
        
        th, td {
            padding: 12px 15px;
            text-align: left;
            border-bottom: 1px solid var(--border-color);
        }
        
        th {
            background-color: var(--background-color);
        }
        
        .histogram {
            height: 200px;
            background-color: #f5f5f5;
            border: 1px solid #ddd;
            border-radius: 4px;
            display: flex;
            align-items: flex-end;
            padding: 10px;
            margin-top: 15px;
        }
        
        .histogram-bar {
            background-color: var(--primary-color);
            margin-right: 2px;
            flex: 1;
        }
        
        .quality-score {
            font-size: 3em;
            font-weight: bold;
            text-align: center;
        }
        
        .score-good {
            color: var(--success-color);
        }
        
        .score-warning {
            color: var(--warning-color);
        }
        
        .score-bad {
            color: var(--error-color);
        }
        
        .issues-list {
            list-style-type: none;
            padding: 0;
        }
        
        .issues-list li {
            padding: 10px;
            margin-bottom: 5px;
            background-color: rgba(217, 48, 37, 0.1);
            border-left: 4px solid var(--error-color);
            border-radius: 4px;
        }
        
        .recommendations-list {
            list-style-type: none;
            padding: 0;
        }
        
        .recommendations-list li {
            padding: 10px;
            margin-bottom: 5px;
            background-color: rgba(15, 157, 88, 0.1);
            border-left: 4px solid var(--success-color);
            border-radius: 4px;
        }
        
        .footer {
            text-align: center;
            margin-top: 40px;
            color: var(--secondary-color);
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>DataSleuth Profile: {{.Profile.Filename}}</h1>
            <p>Generated: {{.GeneratedAt}} | Size: {{formatNumber .FileSizeMB}} MB | Rows: {{formatNumber .Profile.RowCount}} | Columns: {{formatNumber .Profile.ColumnCount}}</p>
        </header>
        
        <div class="summary-cards">
            <div class="card">
                <h2>Quality Score</h2>
                <div class="quality-score {{if ge .Profile.QualityScore 90}}score-good{{else if ge .Profile.QualityScore 70}}score-warning{{else}}score-bad{{end}}">
                    {{.Profile.QualityScore}}/100
                </div>
            </div>
            
            <div class="card">
                <h2>Dataset Summary</h2>
                <p><strong>Rows:</strong> {{formatNumber .Profile.RowCount}}</p>
                <p><strong>Columns:</strong> {{formatNumber .Profile.ColumnCount}}</p>
                <p><strong>Missing cells:</strong> {{formatNumber .Profile.MissingCells}} ({{formatPercent (div .Profile.MissingCells (mul .Profile.RowCount .Profile.ColumnCount))}})</p>
                <p><strong>Duplicate rows:</strong> {{formatNumber .Profile.DuplicateRows}} ({{formatPercent (div .Profile.DuplicateRows .Profile.RowCount)}})</p>
                <p><strong>Processing Time:</strong> {{.Profile.ProcessingTime.Seconds}} seconds</p>
            </div>
            
            <div class="card">
                <h2>Quality Issues</h2>
                {{if .Issues}}
                <ul class="issues-list">
                    {{range .Issues}}
                    <li>{{.}}</li>
                    {{end}}
                </ul>
                {{else}}
                <p>No significant quality issues detected.</p>
                {{end}}
            </div>
        </div>
        
        {{if .Recommendations}}
        <div class="card">
            <h2>Recommendations</h2>
            <ul class="recommendations-list">
                {{range .Recommendations}}
                <li>{{.}}</li>
                {{end}}
            </ul>
        </div>
        {{end}}
        
        <h2>Column Details</h2>
        <div class="column-grid">
            {{range $name, $col := .Profile.Columns}}
            <div class="column-card">
                <h3>{{$name}} <small>({{$col.DataType}})</small></h3>
                
                <table>
                    <tr>
                        <th>Metric</th>
                        <th>Value</th>
                    </tr>
                    <tr>
                        <td>Count</td>
                        <td>{{formatNumber $col.Count}}</td>
                    </tr>
                    <tr>
                        <td>Missing</td>
                        <td>{{formatNumber $col.MissingCount}} ({{formatPercent (div $col.MissingCount $.Profile.RowCount)}})</td>
                    </tr>
                    <tr>
                        <td>Unique</td>
                        <td>{{formatNumber $col.UniqueCount}} ({{formatPercent (div $col.UniqueCount $col.Count)}})</td>
                    </tr>
                    {{if $col.IsNumeric}}
                    <tr>
                        <td>Min</td>
                        <td>{{formatNumber $col.Min}}</td>
                    </tr>
                    <tr>
                        <td>Max</td>
                        <td>{{formatNumber $col.Max}}</td>
                    </tr>
                    <tr>
                        <td>Mean</td>
                        <td>{{formatNumber $col.Mean}}</td>
                    </tr>
                    <tr>
                        <td>Median</td>
                        <td>{{formatNumber $col.Median}}</td>
                    </tr>
                    <tr>
                        <td>Std Dev</td>
                        <td>{{formatNumber $col.StdDev}}</td>
                    </tr>
                    {{end}}
                </table>
                
                {{if $col.IsNumeric}}
                <div class="histogram">
                    {{range $bucket := $col.HistogramBuckets}}
                    <div class="histogram-bar" style="height: {{percentage $bucket.Count $col.Count}}%"></div>
                    {{end}}
                </div>
                {{else if $col.IsCategorical}}
                <h4>Top Values:</h4>
                <ul>
                    {{range $val := $col.TopValues}}
                    <li>{{$val.Value}}: {{formatNumber $val.Count}} ({{formatPercent (div $val.Count $col.Count)}})</li>
                    {{end}}
                </ul>
                {{end}}
                
                {{if $col.QualityIssues}}
                <h4>Quality Issues:</h4>
                <ul class="issues-list">
                    {{range $issue := $col.QualityIssues}}
                    <li>{{$issue.Description}}</li>
                    {{end}}
                </ul>
                {{end}}
            </div>
            {{end}}
        </div>
        
        <div class="footer">
            <p>Generated by DataSleuth v0.1.0 - Fast dataset profiling and validation from the command line</p>
        </div>
    </div>
</body>
</html>`
