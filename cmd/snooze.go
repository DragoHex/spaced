package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"spaced/database"
)

var snoozeDays int

var snoozeCmd = &cobra.Command{
	Use:   "snooze [topic_id]",
	Short: "Postpone a topic's review without advancing its cycle.",
	Args:  cobra.ExactArgs(1),
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
			fmt.Println("Topic is already completed; nothing to snooze.")
			return
		}

		if err := database.SnoozeTopic(id, snoozeDays); err != nil {
			fmt.Println("Error snoozing topic:", err)
			return
		}
		fmt.Printf("Topic snoozed by %d day(s).\n", snoozeDays)
	},
}

func init() {
	rootCmd.AddCommand(snoozeCmd)
	snoozeCmd.Flags().IntVar(&snoozeDays, "days", 1, "Number of days to postpone")
}
