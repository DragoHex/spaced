package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"spaced/database"
)

var (
	exportFormat   string
	exportOutput   string
	exportTemplate bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all topics to Markdown or CSV.",
	Long: `Export all topics (including archived and completed) grouped by project.

Use --template to generate an exhaustive example file with mock data — useful
as a starting point for an import file or to understand the expected format.

Examples:
  spaced export --format markdown
  spaced export --format csv --output study-log.csv
  spaced export --format markdown --template
  spaced export --format csv --template --output template.csv`,
	Run: func(cmd *cobra.Command, args []string) {
		format := strings.ToLower(exportFormat)
		if format != "markdown" && format != "csv" {
			fmt.Println("Error: --format must be 'markdown' or 'csv'.")
			return
		}

		var content string

		if exportTemplate {
			switch format {
			case "markdown":
				content = database.RenderMarkdownTemplate()
			case "csv":
				content = database.RenderCSVTemplate()
			}
		} else {
			groups, err := database.GetTopicsGroupedByProject()
			if err != nil {
				fmt.Println("Error fetching topics:", err)
				return
			}
			switch format {
			case "markdown":
				content = database.RenderMarkdown(groups)
			case "csv":
				content = database.RenderCSV(groups)
			}
		}

		if exportOutput == "" {
			fmt.Print(content)
			return
		}

		if err := os.WriteFile(exportOutput, []byte(content), 0o644); err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
		if exportTemplate {
			fmt.Printf("Template written to %s\n", exportOutput)
		} else {
			fmt.Printf("Export written to %s\n", exportOutput)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVar(&exportFormat, "format", "markdown", "Output format: markdown or csv")
	exportCmd.Flags().StringVar(&exportOutput, "output", "", "Write to file instead of stdout")
	exportCmd.Flags().BoolVar(&exportTemplate, "template", false, "Generate a mock-data template instead of real data")
}
