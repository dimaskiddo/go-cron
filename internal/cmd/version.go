package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version Variable Structure
var Version = &cobra.Command{
	Use:   "version",
	Short: "Show current version",
	Long:  "Go-Cron Version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Go-Cron Version v0.0.1")
	},
}
