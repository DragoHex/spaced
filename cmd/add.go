package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"spaced/database"
)

var (
	addProjectName string
	addNotes       string
)

var addCmd = &cobra.Command{
	Use:   "add [topic]",
	Short: "Add a new topic.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		topic := args[0]

		var projectID int64
		if addProjectName != "" {
			var err error
			projectID, err = database.GetOrCreateProject(addProjectName)
			if err != nil {
				fmt.Println("Error resolving project:", err)
				return
			}
		}

		var err error
		switch {
		case projectID != 0 && addNotes != "":
			err = database.AddTopicFull(topic, addNotes, projectID)
		case projectID != 0:
			err = database.AddTopicWithProject(topic, projectID)
		case addNotes != "":
			err = database.AddTopicWithNotes(topic, addNotes)
		default:
			err = database.AddTopic(topic)
		}

		if err != nil {
			fmt.Println("Error adding topic:", err)
			return
		}

		msg := "Topic added"
		if addProjectName != "" {
			msg += " to project \"" + addProjectName + "\""
		}
		fmt.Println(msg + ".")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVar(&addProjectName, "project", "", "Assign to a project (created if needed)")
	addCmd.Flags().StringVar(&addNotes, "notes", "", "Optional notes or context for the topic")
}
