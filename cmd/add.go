package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"spaced/database"
)

var addCmd = &cobra.Command{
	Use:   "add [topic]",
	Short: "Add a new topic.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.AddTopic(args[0]); err != nil {
			fmt.Println("Error adding topic:", err)
		} else {
			fmt.Println("Topic added successfully.")
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
