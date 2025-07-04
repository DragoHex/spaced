package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"spaced/database"
)

var doneCmd = &cobra.Command{
	Use:   "done [topic_id]",
	Short: "Mark a topic as done.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid topic ID.")
			return
		}

		if err := database.MarkTopicDone(id); err != nil {
			fmt.Println("Error marking topic as done:", err)
		} else {
			fmt.Println("Topic marked as done.")
		}
	},
}

func init() {
	rootCmd.AddCommand(doneCmd)
}
