package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"spaced/database"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all topics.",
	Run: func(cmd *cobra.Command, args []string) {
		topics, err := database.GetAllTopics()
		if err != nil {
			fmt.Println("Error getting topics:", err)
			return
		}

		if len(topics) == 0 {
			fmt.Println("No topics found.")
			return
		}

		// ANSI escape codes for green color and reset
		const (
			green = "\033[32m"
			red   = "\033[31m"
			reset = "\033[0m"
		)

		headers := []string{"ID", "Topic", "Created", "Next Review", "Review Cycle", "Completed", "Archived"}
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
				topic["created_at"].(time.Time).Format("2006-01-02"),
				topic["next_review_at"].(time.Time).Format("2006-01-02"),
				fmt.Sprintf("Day %d", database.GetReviewDay(topic["review_cycle"].(int64))),
				fmt.Sprintf("%t", topic["completed"]),
				fmt.Sprintf("%t", topic["archived"]),
			}
			for i, d := range data {
				if len(d) > colWidths[i] {
					colWidths[i] = len(d)
				}
			}
		}

		// Print header
		for i, header := range headers {
			fmt.Printf("%-" + fmt.Sprintf("%d", colWidths[i]) + "s  ", header)
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
				topic["created_at"].(time.Time).Format("2006-01-02"),
				topic["next_review_at"].(time.Time).Format("2006-01-02"),
				fmt.Sprintf("Day %d", database.GetReviewDay(topic["review_cycle"].(int64))),
				fmt.Sprintf("%t", topic["completed"]),
				fmt.Sprintf("%t", topic["archived"]),
			}

			colorToApply := ""
			if topic["completed"].(bool) {
				colorToApply = green
			} else if topic["next_review_at"].(time.Time).Before(time.Now()) {
				colorToApply = red
			}

			for i, d := range data {
				formattedCell := fmt.Sprintf("%-" + fmt.Sprintf("%d", colWidths[i]) + "s", d)
				fmt.Print(colorToApply + formattedCell + reset + "  ")
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
