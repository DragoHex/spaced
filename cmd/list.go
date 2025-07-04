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

		for _, topic := range topics {
			fmt.Printf("ID: %d, Topic: %s, Created: %s, Next Review: %s, Review Cycle: Day %d, Completed: %t, Archived: %t\n",
				topic["id"], topic["topic"], topic["created_at"].(time.Time).Format("2006-01-02"),
				topic["next_review_at"].(time.Time).Format("2006-01-02"), database.GetReviewDay(topic["review_cycle"].(int64)),
				topic["completed"], topic["archived"],
			)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}