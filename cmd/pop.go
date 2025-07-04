package cmd

import (
	"fmt"
	"spaced/database"

	"github.com/spf13/cobra"
)

var popCmd = &cobra.Command{
	Use:   "pop",
	Short: "Show topics to be reviewed today.",
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

		// ANSI escape codes for green color and reset
		const (
			green = "\033[32m"
			reset = "\033[0m"
		)

		headers := []string{"ID", "Topic", "Review Cycle"}
		colWidths := make([]int, len(headers))

		// Initialize column widths with header lengths
		for i, header := range headers {
			colWidths[i] = len(header)
		}

		// Calculate maximum column widths based on data
		for _, topic := range topics {
			// Convert all fields to string for length calculation
			data := []string{
				fmt.Sprintf("%d", topic["id"]),
				fmt.Sprintf("%s", topic["topic"]),
				fmt.Sprintf("Day %d", database.GetReviewDay(topic["review_cycle"].(int64))),
			}
			for i, d := range data {
				if len(d) > colWidths[i] {
					colWidths[i] = len(d)
				}
			}
		}

		fmt.Println("Topics to review today:")
		// Print header
		for i, header := range headers {
			fmt.Printf("%-"+fmt.Sprintf("%d", colWidths[i])+"s  ", header)
		}
		fmt.Println()

		// Print separator
		for i := range headers {
			for j := 0; j < colWidths[i]; j++ {
				fmt.Print("-")
			}
			fmt.Print("  ")
		}
		fmt.Println()

		// Print data rows
		for _, topic := range topics {
			data := []string{
				fmt.Sprintf("%d", topic["id"]),
				fmt.Sprintf("%s", topic["topic"]),
				fmt.Sprintf("Day %d", database.GetReviewDay(topic["review_cycle"].(int64))),
			}

			for i, d := range data {
				formattedCell := fmt.Sprintf("%-"+fmt.Sprintf("%d", colWidths[i])+"s", d)
				// Pop command only shows incomplete topics, so no green highlighting needed here
				fmt.Print(formattedCell + "  ")
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(popCmd)
}
