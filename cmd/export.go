package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"spaced/database"
)

var (
	exportFormat string
	exportOutput string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all topics to Markdown or CSV.",
	Long: `Export all topics (including archived and completed) grouped by project.

Examples:
  spaced export --format markdown
  spaced export --format csv --output study-log.csv`,
	Run: func(cmd *cobra.Command, args []string) {
		format := strings.ToLower(exportFormat)
		if format != "markdown" && format != "csv" {
			fmt.Println("Error: --format must be 'markdown' or 'csv'.")
			return
		}

		groups, err := database.GetTopicsGroupedByProject()
		if err != nil {
			fmt.Println("Error fetching topics:", err)
			return
		}

		var content string
		switch format {
		case "markdown":
			content = database.RenderMarkdown(groups)
		case "csv":
			content = database.RenderCSV(groups)
		}

		if exportOutput == "" {
			fmt.Print(content)
			return
		}

		if err := os.WriteFile(exportOutput, []byte(content), 0o644); err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
		fmt.Printf("Exported %d groups to %s\n", len(groups), exportOutput)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVar(&exportFormat, "format", "markdown", "Output format: markdown or csv")
	exportCmd.Flags().StringVar(&exportOutput, "output", "", "Write to file instead of stdout")
}
