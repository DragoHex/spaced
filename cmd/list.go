package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"spaced/database"
)

var (
	listProjectFilter  string
	listOverdue        bool
	listCompleted      bool
	listIncludeArchived bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List topics. Use flags to filter.",
	Run: func(cmd *cobra.Command, args []string) {
		filter := database.TopicFilter{
			Overdue:         listOverdue,
			Completed:       listCompleted,
			IncludeArchived: listIncludeArchived,
		}

		if listProjectFilter != "" {
			projectID, err := database.GetOrCreateProject(listProjectFilter)
			if err != nil {
				fmt.Println("Error resolving project:", err)
				return
			}
			filter.ProjectID = &projectID
		}

		topics, err := database.GetTopicsFiltered(filter)
		if err != nil {
			fmt.Println("Error getting topics:", err)
			return
		}
		if len(topics) == 0 {
			fmt.Println("No topics found.")
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Topic", "Project", "Created", "Next Review", "Cycle", "Done", "Archived"})
		table.SetBorder(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetColumnSeparator("  ")
		table.SetHeaderLine(true)

		now := time.Now()
		for _, t := range topics {
			projectLabel := t.ProjectName
			if projectLabel == "" {
				projectLabel = "-"
			}
			row := []string{
				fmt.Sprintf("%d", t.ID),
				t.Topic,
				projectLabel,
				t.CreatedAt.Format("2006-01-02"),
				t.NextReviewAt.Format("2006-01-02"),
				fmt.Sprintf("Day %d", database.GetReviewDay(t.ReviewCycle)),
				fmt.Sprintf("%t", t.Completed),
				fmt.Sprintf("%t", t.Archived),
			}

			colors := make([]tablewriter.Colors, len(row))
			if t.Completed {
				_ = color.New(color.FgGreen)
				for i := range colors {
					colors[i] = tablewriter.Colors{tablewriter.FgGreenColor}
				}
			} else if t.NextReviewAt.Before(now) && !t.Archived {
				_ = color.New(color.FgRed)
				for i := range colors {
					colors[i] = tablewriter.Colors{tablewriter.FgRedColor}
				}
			}
			table.Rich(row, colors)
		}

		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listProjectFilter, "project", "", "Filter by project name")
	listCmd.Flags().BoolVar(&listOverdue, "overdue", false, "Show only overdue topics")
	listCmd.Flags().BoolVar(&listCompleted, "completed", false, "Show only completed topics")
	listCmd.Flags().BoolVar(&listIncludeArchived, "archived", false, "Include archived topics")
}
