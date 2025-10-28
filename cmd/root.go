package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tribal",
	Short: "Tribal CLI - Graph-based development workflow",
	Long: `Tribal CLI provides commands for managing graph-based development workflows.
Use tribal to initialize repositories, manage graphs, and collaborate on graph structures.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}