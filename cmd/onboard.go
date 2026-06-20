package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"spaced/database"
)

var onboardCycle int64

var onboardCmd = &cobra.Command{
	Use:   "onboard <id>",
	Short: "Onboard a parked topic back onto the revision cycle.",
	Long: `Onboard a parked topic so it appears in review sessions again.
By default the topic resumes from the cycle stage it was at when parked.
Use --cycle to start from a specific stage (0–6).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Error: invalid topic id")
			return
		}

		var cycle *int64
		if cmd.Flags().Changed("cycle") {
			if onboardCycle < 0 || onboardCycle > 6 {
				fmt.Println("Error: --cycle must be 0–6")
				return
			}
			cycle = &onboardCycle
		}

		if err := database.OnboardTopic(id, cycle); err != nil {
			fmt.Println("Error onboarding topic:", err)
			return
		}
		fmt.Printf("Topic %d onboarded — it is now due for review.\n", id)
	},
}

func init() {
	rootCmd.AddCommand(onboardCmd)
	onboardCmd.Flags().Int64Var(&onboardCycle, "cycle", 0, "Revision cycle stage to start from (0–6)")
}
