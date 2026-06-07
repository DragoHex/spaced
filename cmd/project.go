package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"spaced/database"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects.",
}

var projectAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Create a new project.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.AddProject(args[0]); err != nil {
			fmt.Println("Error creating project:", err)
			return
		}
		fmt.Printf("Project %q created.\n", args[0])
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects.",
	Run: func(cmd *cobra.Command, args []string) {
		projects, err := database.GetAllProjects()
		if err != nil {
			fmt.Println("Error listing projects:", err)
			return
		}
		if len(projects) == 0 {
			fmt.Println("No projects found.")
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name"})
		table.SetBorder(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		for _, p := range projects {
			table.Append([]string{fmt.Sprintf("%d", p.ID), p.Name})
		}
		table.Render()
	},
}

var projectRenameCmd = &cobra.Command{
	Use:   "rename [project_id] [new_name]",
	Short: "Rename an existing project.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid project ID.")
			return
		}
		if err := database.RenameProject(id, args[1]); err != nil {
			fmt.Println("Error renaming project:", err)
			return
		}
		fmt.Printf("Project renamed to %q.\n", args[1])
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete [project_id]",
	Short: "Delete a project (topics are unassigned, not deleted).",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid project ID.")
			return
		}
		if err := database.DeleteProject(id); err != nil {
			fmt.Println("Error deleting project:", err)
			return
		}
		fmt.Println("Project deleted. Its topics have been unassigned.")
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectAddCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectRenameCmd)
	projectCmd.AddCommand(projectDeleteCmd)
}
