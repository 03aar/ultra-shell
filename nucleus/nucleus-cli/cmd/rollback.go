package cmd

import (
	"encoding/json"
	"fmt"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var rollbackLast bool

var rollbackCmd = &cobra.Command{
	Use:   "rollback [execution-id]",
	Short: "Rollback a command execution",
	Long:  `Undo a previously executed command by its execution ID, or use --last to rollback the most recent command.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var executionID string

		if rollbackLast {
			// Fetch the latest execution to get its ID
			executions, err := client.GetExecutions(1, "", "")
			if err != nil {
				return fmt.Errorf("failed to fetch latest execution: %w", err)
			}
			if len(executions) == 0 {
				return fmt.Errorf("no executions found to rollback")
			}
			executionID = executions[0].ID
			fmt.Println(display.StyleMuted("Rolling back last execution: " + truncate(executions[0].Command, 60)))
		} else if len(args) == 1 {
			executionID = args[0]
		} else {
			return fmt.Errorf("provide an execution ID or use --last")
		}

		result, err := client.Rollback(executionID)
		if err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		if output == "json" {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleHeader("Rollback Result"))
		fmt.Println()

		if result.Success {
			fmt.Println(display.StyleSuccess("Rollback completed successfully"))
		} else {
			fmt.Println(display.StyleError("Rollback failed"))
		}

		fmt.Println()
		fmt.Println(display.FormatKeyValue("Execution", result.ExecutionID))
		fmt.Println(display.FormatKeyValue("Original", result.Command))
		fmt.Println(display.FormatKeyValue("Undo Command", result.UndoCommand))
		if result.Message != "" {
			fmt.Println(display.FormatKeyValue("Message", result.Message))
		}

		return nil
	},
}

func init() {
	rollbackCmd.Flags().BoolVar(&rollbackLast, "last", false, "Rollback the most recent execution")
	rootCmd.AddCommand(rollbackCmd)
}
