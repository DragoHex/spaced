package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"spaced/database"
)

var parkCmd = &cobra.Command{
	Use:   "park <id>",
	Short: "Park a topic (suspend from revision cycle, preserving progress).",
	Long: `Park a topic to remove it from the revision cycle without losing its progress.
Parked topics remain visible in lists but are excluded from review sessions.
Use 'onboard' to return a parked topic to the cycle.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Error: invalid topic id")
			return
		}
		if err := database.ParkTopic(id); err != nil {
			fmt.Println("Error parking topic:", err)
			return
		}
		fmt.Printf("Topic %d parked. Use 'spd onboard %d' to resume it.\n", id, id)
	},
}

func init() {
	rootCmd.AddCommand(parkCmd)
}
