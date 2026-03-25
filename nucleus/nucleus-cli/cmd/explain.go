package cmd

import (
	"encoding/json"
	"fmt"
	"nucleus-cli/internal/api"
	"nucleus-cli/internal/display"
	"strings"

	"github.com/spf13/cobra"
)

var explainCmd = &cobra.Command{
	Use:   "explain [command]",
	Short: "Explain what a command does without running it",
	Long: `Explain what a shell command does in plain English, including
what files it will affect and potential risks.

Examples:
  nuc explain "find . -name '*.log' -mtime +7 -delete"
  nuc explain "chmod -R 755 /var/www"
  nuc explain "docker system prune -a -f"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := strings.Join(args, " ")
		client := api.NewClient(apiURL, apiKey)

		// Use dry-run risk assessment
		result, err := client.Execute(command, true)
		if err != nil {
			return fmt.Errorf("failed to assess: %w", err)
		}

		if outputFormat == "json" {
			jsonOut, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(jsonOut))
			return nil
		}

		fmt.Println()
		fmt.Println(display.StyleHeader("  Command Explanation"))
		fmt.Println()
		fmt.Printf("  %s %s\n", display.StyleMuted("Command:"), display.StyleSuccess(command))
		fmt.Println()

		// Parse the binary for common explanations
		explanations := map[string]string{
			"find":    "Search for files matching criteria in directory tree",
			"rm":      "Remove files or directories",
			"chmod":   "Change file permissions",
			"chown":   "Change file ownership",
			"docker":  "Manage Docker containers, images, and resources",
			"git":     "Version control operation",
			"curl":    "Transfer data from/to a URL",
			"wget":    "Download files from the web",
			"npm":     "Node.js package manager operation",
			"pip":     "Python package manager operation",
			"cargo":   "Rust package manager and build tool",
			"make":    "Build automation using Makefile rules",
			"kubectl": "Kubernetes cluster management",
			"ssh":     "Secure shell connection to remote host",
			"scp":     "Secure copy files to/from remote host",
			"tar":     "Archive/extract files",
			"grep":    "Search file contents for patterns",
			"sed":     "Stream editor for text transformation",
			"awk":     "Pattern scanning and text processing",
			"kill":    "Send signal to process (terminate by default)",
		}

		binary := strings.Fields(command)[0]
		if desc, ok := explanations[binary]; ok {
			fmt.Printf("  %s %s\n", display.StyleMuted("Type:"), desc)
		}

		if result.RiskLevel != "" {
			fmt.Printf("  %s %s\n", display.StyleMuted("Risk:"), display.FormatRiskBadge(result.RiskLevel))
		}

		if len(result.Warnings) > 0 {
			fmt.Printf("  %s\n", display.StyleWarning("Warnings:"))
			for _, w := range result.Warnings {
				fmt.Printf("    • %s\n", w)
			}
		}

		if len(result.Suggestions) > 0 {
			fmt.Printf("  %s\n", display.StyleMuted("Suggestions:"))
			for _, s := range result.Suggestions {
				fmt.Printf("    → %s\n", s)
			}
		}

		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)
}
