package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"spaced/database"
)

var importFormat string

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import topics from a Markdown or CSV file.",
	Long: `Import topics from a file exported by "spaced export".

The file format is detected automatically from its extension (.md / .csv)
or can be forced with --format.

Only the Topic, Notes, Project name, and Project description columns are used;
all other fields (IDs, dates, cycles) are ignored and recalculated fresh.

Examples:
  spaced import study-log.md
  spaced import study-log.csv
  spaced import data.txt --format markdown`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		content := string(data)

		// Auto-detect format from extension unless --format is given.
		format := strings.ToLower(importFormat)
		if format == "" {
			lower := strings.ToLower(path)
			switch {
			case strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown"):
				format = "markdown"
			case strings.HasSuffix(lower, ".csv"):
				format = "csv"
			default:
				fmt.Println("Error: cannot detect format. Use --format markdown or --format csv.")
				return
			}
		}

		var groups []database.ImportGroup
		switch format {
		case "markdown":
			groups, err = database.ParseMarkdownImport(content)
		case "csv":
			groups, err = database.ParseCSVImport(content)
		default:
			fmt.Println("Error: --format must be 'markdown' or 'csv'.")
			return
		}
		if err != nil {
			fmt.Println("Error parsing file:", err)
			return
		}

		count, err := database.ImportGroups(groups)
		if err != nil {
			fmt.Printf("Import failed after %d topics: %v\n", count, err)
			return
		}

		fmt.Printf("Imported %d topic(s) from %d project(s).\n", count, len(groups))
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVar(&importFormat, "format", "", "Force format: markdown or csv (auto-detected from extension by default)")
}
