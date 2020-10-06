package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dimaskiddo/go-cron/internal/cmd"
)

// Root Variable Structure
var r = &cobra.Command{
	Use:   "go-cron",
	Short: "Go-Cron an Alternatives Binaries for Cron",
	Long:  "Go-Cron an Alternatives Binaries for Cron",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Init Function
func init() {
	// Add Child for Root Command
	r.AddCommand(cmd.Version)
	r.AddCommand(cmd.Daemon)
}

// Main Function
func main() {
	err := r.Execute()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
