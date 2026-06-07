package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"spaced/database"
)

var (
	modifyTopic       string
	modifyNotes       string
	modifyReviewCycle int64
	modifyProject     string
)

var modifyCmd = &cobra.Command{
	Use:   "modify [topic_id]",
	Short: "Modify a topic's text, review cycle, or project.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid topic ID.")
			return
		}

		topicChanged := modifyTopic != ""
		notesChanged := cmd.Flags().Changed("notes")
		cycleChanged := cmd.Flags().Changed("review-cycle")
		projectChanged := modifyProject != ""

		if !topicChanged && !notesChanged && !cycleChanged && !projectChanged {
			fmt.Println("Error: provide at least one of --topic, --notes, --review-cycle, or --project.")
			return
		}

		if topicChanged {
			if err := database.ModifyTopic(id, modifyTopic); err != nil {
				fmt.Println("Error modifying topic text:", err)
				return
			}
			fmt.Println("Topic text updated.")
		}

		if notesChanged {
			if err := database.UpdateNotes(id, modifyNotes); err != nil {
				fmt.Println("Error updating notes:", err)
				return
			}
			fmt.Println("Notes updated.")
		}

		if cycleChanged {
			if err := database.UpdateTopicReviewCycle(id, modifyReviewCycle); err != nil {
				fmt.Println("Error modifying review cycle:", err)
				return
			}
			fmt.Println("Review cycle updated.")
		}

		if projectChanged {
			projectID, err := database.GetOrCreateProject(modifyProject)
			if err != nil {
				fmt.Println("Error resolving project:", err)
				return
			}
			if err := database.AssignTopicToProject(id, projectID); err != nil {
				fmt.Println("Error assigning project:", err)
				return
			}
			fmt.Printf("Topic assigned to project %q.\n", modifyProject)
		}
	},
}

func init() {
	rootCmd.AddCommand(modifyCmd)
	modifyCmd.Flags().StringVar(&modifyTopic, "topic", "", "New topic text")
	modifyCmd.Flags().StringVar(&modifyNotes, "notes", "", "New notes or context")
	modifyCmd.Flags().Int64Var(&modifyReviewCycle, "review-cycle", 0, "New review cycle (0–4)")
	modifyCmd.Flags().StringVar(&modifyProject, "project", "", "Assign to a project (created if needed)")
}
