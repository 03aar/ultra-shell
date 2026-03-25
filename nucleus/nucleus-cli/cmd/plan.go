package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan <goal>",
	Short: "Generate an execution plan for a goal",
	Long:  `Use AI to generate a step-by-step shell execution plan for achieving a stated goal.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		goal := strings.Join(args, " ")

		plan, err := client.Plan(goal)
		if err != nil {
			return fmt.Errorf("planning failed: %w", err)
		}

		if output == "json" {
			data, _ := json.MarshalIndent(plan, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleHeader("Execution Plan"))
		fmt.Println()
		fmt.Println(display.FormatKeyValue("Goal", plan.Goal))
		fmt.Print(display.StyleLabel("Overall Risk: "))
		fmt.Println(display.FormatRiskBadge(plan.Risk))
		fmt.Println()

		if len(plan.Steps) == 0 {
			fmt.Println(display.StyleMuted("No steps generated."))
			return nil
		}

		headers := []string{"#", "COMMAND", "DESCRIPTION", "RISK", "UNDO?"}
		rows := make([][]string, len(plan.Steps))
		for i, step := range plan.Steps {
			rev := "no"
			if step.Reversible {
				rev = "yes"
			}
			rows[i] = []string{
				fmt.Sprintf("%d", step.Order),
				truncate(step.Command, 35),
				truncate(step.Description, 30),
				step.RiskLevel,
				rev,
			}
		}
		fmt.Print(display.FormatTable(headers, rows))

		if plan.Notes != "" {
			fmt.Println()
			fmt.Println(display.StyleLabel("Notes:"))
			fmt.Println("  " + plan.Notes)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(planCmd)
}
