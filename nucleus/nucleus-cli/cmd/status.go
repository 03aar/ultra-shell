package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current session status and recent activity",
	Long:  `Display the current session information, recent commands, and git context.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, ctxErr := client.GetContext()
		executions, execErr := client.GetExecutions(5, "", "")

		if output == "json" {
			result := map[string]interface{}{
				"context":    nil,
				"executions": nil,
			}
			if ctxErr == nil {
				result["context"] = ctx
			}
			if execErr == nil {
				result["executions"] = executions
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleHeader("Nucleus Status"))
		fmt.Println()

		if ctxErr != nil {
			fmt.Println(display.StyleError("Could not fetch context: " + ctxErr.Error()))
		} else {
			fmt.Println(display.FormatKeyValue("Session", ctx.SessionID))
			fmt.Println(display.FormatKeyValue("Directory", ctx.WorkingDir))
			fmt.Println(display.FormatKeyValue("User", ctx.User+"@"+ctx.Hostname))
			fmt.Println(display.FormatKeyValue("Shell", ctx.Shell))
			if ctx.GitRepo != "" {
				fmt.Println()
				fmt.Println(display.StyleLabel("Git:"))
				fmt.Println(display.FormatKeyValue("  Repo", ctx.GitRepo))
				fmt.Println(display.FormatKeyValue("  Branch", ctx.GitBranch))
				fmt.Println("  " + display.StyleLabel("Dirty: ") + display.FormatBool(ctx.GitDirty))
			}
		}

		fmt.Println()
		fmt.Println(display.StyleLabel("Recent Commands:"))

		if execErr != nil {
			fmt.Println(display.StyleError("Could not fetch history: " + execErr.Error()))
		} else if len(executions) == 0 {
			fmt.Println(display.StyleMuted("  No commands recorded yet."))
		} else {
			headers := []string{"TIME", "COMMAND", "EXIT", "RISK"}
			rows := make([][]string, len(executions))
			for i, e := range executions {
				rows[i] = []string{
					e.Timestamp.Format(time.Kitchen),
					truncate(e.Command, 50),
					fmt.Sprintf("%d", e.ExitCode),
					e.RiskLevel,
				}
			}
			fmt.Print(display.FormatTable(headers, rows))
		}

		return nil
	},
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
