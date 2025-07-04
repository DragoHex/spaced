package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"spaced/database"
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

		fmt.Println("Topics to review today:")
		for _, topic := range topics {
			fmt.Printf("ID: %d, Topic: %s, Review Cycle: Day %d\n", topic["id"], topic["topic"], database.GetReviewDay(topic["review_cycle"].(int64)))
		}
	},
}



func init() {
	rootCmd.AddCommand(popCmd)
}
