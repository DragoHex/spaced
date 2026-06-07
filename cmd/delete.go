package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"spaced/database"
)

var deleteYes bool

var deleteCmd = &cobra.Command{
	Use:   "delete [topic_id]",
	Short: "Delete a topic permanently.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid topic ID.")
			return
		}

		if !deleteYes {
			fmt.Printf("Delete topic #%d? This cannot be undone. [y/N] ", id)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return
			}
		}

		if err := database.DeleteTopic(id); err != nil {
			fmt.Println("Error deleting topic:", err)
			return
		}
		fmt.Println("Topic deleted.")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "Skip confirmation prompt")
}
