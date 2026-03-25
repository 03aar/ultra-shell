package cmd

import (
	"encoding/json"
	"fmt"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var contextJSON bool

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Show current shell context",
	Long:  `Display the current working directory, git status, environment, and session information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := client.GetContext()
		if err != nil {
			return fmt.Errorf("failed to fetch context: %w", err)
		}

		if output == "json" || contextJSON {
			data, _ := json.MarshalIndent(ctx, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleHeader("Shell Context"))
		fmt.Println()
		fmt.Println(display.FormatKeyValue("Working Dir", ctx.WorkingDir))
		fmt.Println(display.FormatKeyValue("User", ctx.User))
		fmt.Println(display.FormatKeyValue("Hostname", ctx.Hostname))
		fmt.Println(display.FormatKeyValue("Shell", ctx.Shell))
		fmt.Println(display.FormatKeyValue("Session", ctx.SessionID))

		if ctx.GitRepo != "" {
			fmt.Println()
			fmt.Println(display.StyleLabel("Git:"))
			fmt.Println(display.FormatKeyValue("  Repository", ctx.GitRepo))
			fmt.Println(display.FormatKeyValue("  Branch", ctx.GitBranch))
			fmt.Println("  " + display.StyleLabel("Dirty: ") + display.FormatBool(ctx.GitDirty))
		}

		if len(ctx.Environment) > 0 {
			fmt.Println()
			fmt.Println(display.StyleLabel("Environment:"))
			for k, v := range ctx.Environment {
				fmt.Println(display.FormatKeyValue("  "+k, v))
			}
		}

		return nil
	},
}

func init() {
	contextCmd.Flags().BoolVar(&contextJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(contextCmd)
}
