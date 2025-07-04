package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"spaced/database"
)

var archiveCmd = &cobra.Command{
	Use:   "archive [topic_id]",
	Short: "Archive a completed topic.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid topic ID.")
			return
		}

		_, archived, err := database.GetTopicStatus(id)
		if err != nil {
			fmt.Println("Error getting topic status:", err)
			return
		}

		if archived {
			fmt.Println("Topic is already archived.")
			return
		}

		if err := database.ArchiveTopic(id); err != nil {
			fmt.Println("Error archiving topic:", err)
		} else {
			fmt.Println("Topic archived.")
		}
	},
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}
