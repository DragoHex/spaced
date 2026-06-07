package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"spaced/database"
)

var popCmd = &cobra.Command{
	Use:   "pop",
	Short: "Show topics due for review today.",
	Run: func(cmd *cobra.Command, args []string) {
		topics, err := database.GetTopicsToReview()
		if err != nil {
			fmt.Println("Error getting topics to review:", err)
			return
		}
		if len(topics) == 0 {
			fmt.Println("No topics to review today.")
			return
		}

		fmt.Printf("Topics to review today (%d):\n", len(topics))

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Topic", "Project", "Cycle"})
		table.SetBorder(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetColumnSeparator("  ")
		table.SetHeaderLine(true)

		for _, t := range topics {
			projectLabel := t.ProjectName
			if projectLabel == "" {
				projectLabel = "-"
			}
			table.Append([]string{
				fmt.Sprintf("%d", t.ID),
				t.Topic,
				projectLabel,
				fmt.Sprintf("Day %d", database.GetReviewDay(t.ReviewCycle)),
			})
		}
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(popCmd)
}
