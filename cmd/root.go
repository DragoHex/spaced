package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"spaced/database"
)

var rootCmd = &cobra.Command{
	Use:   "spaced",
	Short: "A simple CLI tool for spaced repetition.",
	Long:  `A simple CLI tool to manage topics for spaced repetition.`,
}

func init() {
	database.InitDB()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}
