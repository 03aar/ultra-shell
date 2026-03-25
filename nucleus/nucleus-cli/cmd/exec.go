package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var execDryRun bool

var execCmd = &cobra.Command{
	Use:   "exec <command>",
	Short: "Execute a command through Nucleus",
	Long:  `Execute a shell command with Nucleus tracking, risk analysis, and optional dry-run mode.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := strings.Join(args, " ")

		execution, err := client.Execute(command, execDryRun)
		if err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}

		if output == "json" {
			data, _ := json.MarshalIndent(execution, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if execDryRun {
			fmt.Println(display.StyleHeader("Dry Run Analysis"))
		} else {
			fmt.Println(display.StyleHeader("Execution Result"))
		}
		fmt.Println()

		fmt.Println(display.FormatKeyValue("Command", execution.Command))
		fmt.Println(display.FormatKeyValue("ID", execution.ID))
		fmt.Print(display.StyleLabel("Risk: "))
		fmt.Println(display.FormatRiskBadge(execution.RiskLevel))
		fmt.Print(display.StyleLabel("Exit Code: "))
		fmt.Println(display.FormatExitCode(execution.ExitCode))
		fmt.Println(display.FormatKeyValue("Duration", fmt.Sprintf("%.2fs", execution.Duration)))
		fmt.Println("" + display.StyleLabel("Reversible: ") + display.FormatBool(execution.Reversible))

		if execution.Explanation != "" {
			fmt.Println()
			fmt.Println(display.StyleLabel("Explanation:"))
			fmt.Println("  " + execution.Explanation)
		}

		if execution.Output != "" {
			fmt.Println()
			fmt.Println(display.StyleLabel("Output:"))
			fmt.Println(execution.Output)
		}

		return nil
	},
}

func init() {
	execCmd.Flags().BoolVar(&execDryRun, "dry-run", false, "Analyze command without executing")
	rootCmd.AddCommand(execCmd)
}
