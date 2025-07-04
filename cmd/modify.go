package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"spaced/database"
)

var ( 
	modifyTopic string
	modifyReviewCycle int64
)

var modifyCmd = &cobra.Command{
	Use:   "modify [topic_id]",
	Short: "Modify an existing topic or its review cycle.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid topic ID.")
			return
		}

		// Check if either --topic or --review-cycle flag is provided
		if modifyTopic == "" && !cmd.Flags().Changed("review-cycle") {
			fmt.Println("Error: Either --topic or --review-cycle flag must be provided.")
			return
		}

		// Modify topic text if --topic flag is provided
		if modifyTopic != "" {
			if err := database.ModifyTopic(id, modifyTopic); err != nil {
				fmt.Println("Error modifying topic text:", err)
				return
			} else {
				fmt.Println("Topic text modified.")
			}
		}

		// Modify review cycle if --review-cycle flag is provided
		if cmd.Flags().Changed("review-cycle") {
			if err := database.UpdateTopicReviewCycle(id, modifyReviewCycle); err != nil {
				fmt.Println("Error modifying review cycle:", err)
				return
			} else {
				fmt.Println("Topic review cycle modified.")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(modifyCmd)
	modifyCmd.Flags().StringVar(&modifyTopic, "topic", "", "New topic text")
	modifyCmd.Flags().Int64Var(&modifyReviewCycle, "review-cycle", 0, "New review cycle (1, 2, 3, or 4)")
}