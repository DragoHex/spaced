package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"spaced/database"
)

var unarchiveCmd = &cobra.Command{
	Use:   "unarchive [topic_id]",
	Short: "Unarchive a topic.",
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

		if !archived {
			fmt.Println("Topic is not archived.")
			return
		}

		if err := database.UnarchiveTopic(id); err != nil {
			fmt.Println("Error unarchiving topic:", err)
		} else {
			fmt.Println("Topic unarchived.")
		}
	},
}

func init() {
	rootCmd.AddCommand(unarchiveCmd)
}
