package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"spaced/database"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [topic_id]",
	Short: "Delete a topic.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid topic ID.")
			return
		}

		if err := database.DeleteTopic(id); err != nil {
			fmt.Println("Error deleting topic:", err)
		} else {
			fmt.Println("Topic deleted.")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
