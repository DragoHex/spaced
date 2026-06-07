package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"spaced/database"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show a summary of your study progress.",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := database.GetStats()
		if err != nil {
			fmt.Println("Error getting stats:", err)
			return
		}

		completionPct := 0
		if s.Total > 0 {
			completionPct = (s.Completed * 100) / s.Total
		}

		fmt.Println("─────────────────────────────")
		fmt.Printf("  Total topics:   %d\n", s.Total)
		fmt.Printf("  Completed:      %d  (%d%%)\n", s.Completed, completionPct)
		fmt.Printf("  In progress:    %d\n", s.InProgress)
		fmt.Printf("  Due today:      %d\n", s.DueToday)
		fmt.Printf("  Overdue:        %d\n", s.Overdue)
		fmt.Printf("  Archived:       %d\n", s.Archived)
		fmt.Printf("  Projects:       %d\n", s.Projects)
		fmt.Println("─────────────────────────────")
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
