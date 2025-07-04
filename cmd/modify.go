package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"spaced/database"
)

var modifyCmd = &cobra.Command{
	Use:   "modify [topic_id] [new_topic]",
	Short: "Modify an existing topic.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid topic ID.")
			return
		}

		if err := database.ModifyTopic(id, args[1]); err != nil {
			fmt.Println("Error modifying topic:", err)
		} else {
			fmt.Println("Topic modified.")
		}
	},
}

func init() {
	rootCmd.AddCommand(modifyCmd)
}
