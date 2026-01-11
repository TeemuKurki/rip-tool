package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rip-tool",
	Short: "DVD / Bluray ripping tool",
}

func Execute() {
	//doc.GenMarkdownTree(rootCmd, "./docs")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
