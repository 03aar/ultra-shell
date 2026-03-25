package cmd

import (
	"fmt"
	"os"

	"nucleus-cli/internal/api"
	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var (
	apiURL   string
	apiKey   string
	output   string
	noColor  bool
	client   *api.NucleusClient
)

var rootCmd = &cobra.Command{
	Use:   "nuc",
	Short: "Nucleus CLI — intelligent shell command platform",
	Long: `nuc is the command-line interface for the Nucleus platform.
It provides execution tracking, rollback, session management,
context awareness, and AI-powered shell planning.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		display.NoColor = noColor
		client = api.NewClient(apiURL, apiKey)
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, display.StyleError(err.Error()))
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:8080", "Nucleus API base URL")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "dev-nucleus-key-local", "API key for authentication")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "Output format: json or text")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}
