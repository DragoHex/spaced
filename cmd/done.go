package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"spaced/database"
)

var doneQuality int

var doneCmd = &cobra.Command{
	Use:   "done [topic_id]",
	Short: "Mark a topic as reviewed and advance to the next cycle.",
	Long: `Mark a topic as reviewed.

Without --quality, uses the fixed schedule (Day 1 → 4 → 11 → 25 → 55 → 115).
With --quality (0–5), uses the SM-2 algorithm for adaptive intervals:
  0 = complete blackout
  1 = incorrect, remembered on seeing answer
  2 = incorrect, easy to remember
  3 = correct with significant difficulty
  4 = correct after hesitation
  5 = perfect recall`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid topic ID.")
			return
		}

		completed, _, err := database.GetTopicStatus(id)
		if err != nil {
			fmt.Println("Error getting topic status:", err)
			return
		}
		if completed {
			fmt.Println("Topic is already completed.")
			return
		}

		if cmd.Flags().Changed("quality") {
			// SM-2 path
			nextDate, err := database.MarkTopicDoneWithQuality(id, doneQuality)
			if err != nil {
				fmt.Println("Error marking topic as done:", err)
				return
			}
			_ = database.LogReview(id)
			topics, err := database.GetAllTopics()
			if err != nil {
				fmt.Printf("Done! Next review: %s\n", nextDate.Format("2006-01-02"))
				return
			}
			for _, t := range topics {
				if t.ID == id {
					fmt.Printf("Done! Next review: %s (interval: %d days, EF: %.2f)\n",
						nextDate.Format("2006-01-02"),
						t.IntervalDays,
						t.EasinessFactor,
					)
					return
				}
			}
			fmt.Printf("Done! Next review: %s\n", nextDate.Format("2006-01-02"))
			return
		}

		// Fixed-schedule path (default, backward-compatible).
		nextReview, err := database.MarkTopicDone(id)
		if err != nil {
			fmt.Println("Error marking topic as done:", err)
			return
		}
		_ = database.LogReview(id)

		topics, err := database.GetAllTopics()
		if err != nil {
			fmt.Println("Topic marked as done.")
			return
		}
		for _, t := range topics {
			if t.ID == id {
				if t.Completed {
					fmt.Println("Topic completed! All review cycles finished.")
				} else {
					fmt.Printf("Done! Next review: %s (Day %d)\n",
						nextReview.Format("2006-01-02"),
						database.GetReviewDay(t.ReviewCycle),
					)
				}
				return
			}
		}
		fmt.Println("Topic marked as done.")
	},
}

func init() {
	rootCmd.AddCommand(doneCmd)
	doneCmd.Flags().IntVar(&doneQuality, "quality", 4, "SM-2 quality rating 0–5 (activates adaptive scheduling)")
}
