package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"spaced/database"
)

const unassignedProject = "UNASSIGNED"

var (
	addProjectName string
	addProjectID   int64
	addNotes       string
	addPark        bool
)

var addCmd = &cobra.Command{
	Use:   "add [topic]",
	Short: "Add a new topic.",
	Long: `Add a new topic to your spaced repetition list.

If neither --project nor --project-id is given, the topic is placed in the
"UNASSIGNED" project (created automatically if it doesn't exist).

Flags --project and --project-id are mutually exclusive.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		topic := args[0]

		projectFlagSet := addProjectName != ""
		projectIDFlagSet := cmd.Flags().Changed("project-id")

		if projectFlagSet && projectIDFlagSet {
			fmt.Println("Error: --project and --project-id are mutually exclusive.")
			return
		}

		var projectID int64
		var resolvedProjectName string

		switch {
		case projectIDFlagSet:
			// Validate the project exists.
			p, err := database.GetProjectByID(addProjectID)
			if err != nil {
				fmt.Printf("Error: project with ID %d not found.\n", addProjectID)
				return
			}
			projectID = p.ID
			resolvedProjectName = p.Name

		case projectFlagSet:
			var err error
			projectID, err = database.GetOrCreateProject(addProjectName)
			if err != nil {
				fmt.Println("Error resolving project:", err)
				return
			}
			resolvedProjectName = addProjectName

		default:
			// Default: assign to UNASSIGNED project.
			var err error
			projectID, err = database.GetOrCreateProject(unassignedProject)
			if err != nil {
				fmt.Println("Error resolving UNASSIGNED project:", err)
				return
			}
			resolvedProjectName = unassignedProject
		}

		var (
			newID int64
			err   error
		)
		if addNotes != "" {
			newID, err = database.AddTopicFull(topic, addNotes, projectID)
		} else {
			newID, err = database.AddTopicWithProject(topic, projectID)
		}

		if err != nil {
			fmt.Println("Error adding topic:", err)
			return
		}

		if addPark {
			if err := database.ParkTopic(newID); err != nil {
				fmt.Println("Error parking topic:", err)
				return
			}
			fmt.Printf("Topic added to project %q (parked — not in revision cycle).\n", resolvedProjectName)
			return
		}
		fmt.Printf("Topic added to project %q.\n", resolvedProjectName)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVar(&addProjectName, "project", "", "Assign to a project by name (created if needed)")
	addCmd.Flags().Int64Var(&addProjectID, "project-id", 0, "Assign to a project by its numeric ID")
	addCmd.Flags().StringVar(&addNotes, "notes", "", "Optional notes or context for the topic")
	addCmd.Flags().BoolVar(&addPark, "park", false, "Park the topic without adding it to the revision cycle")
}
