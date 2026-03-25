package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var skillParams []string

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage and run skills",
	Long:  `List available skills or run a specific skill with parameters.`,
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		skills, err := client.GetSkills()
		if err != nil {
			return fmt.Errorf("failed to fetch skills: %w", err)
		}

		if output == "json" {
			data, _ := json.MarshalIndent(skills, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleHeader("Available Skills"))
		fmt.Println()

		if len(skills) == 0 {
			fmt.Println(display.StyleMuted("No skills available."))
			return nil
		}

		headers := []string{"NAME", "CATEGORY", "DESCRIPTION", "PARAMETERS"}
		rows := make([][]string, len(skills))
		for i, s := range skills {
			paramNames := make([]string, 0, len(s.Parameters))
			for k := range s.Parameters {
				paramNames = append(paramNames, k)
			}
			rows[i] = []string{
				s.Name,
				s.Category,
				truncate(s.Description, 40),
				strings.Join(paramNames, ", "),
			}
		}
		fmt.Print(display.FormatTable(headers, rows))

		return nil
	},
}

var skillRunCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Run a skill",
	Long:  `Execute a skill by name with optional key=value parameters.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		params := make(map[string]string)
		for _, p := range skillParams {
			parts := strings.SplitN(p, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid parameter format %q, expected key=value", p)
			}
			params[parts[0]] = parts[1]
		}

		result, err := client.RunSkill(args[0], params)
		if err != nil {
			return fmt.Errorf("skill execution failed: %w", err)
		}

		if output == "json" {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if result.Success {
			fmt.Println(display.StyleSuccess("Skill completed: " + result.Skill))
		} else {
			fmt.Println(display.StyleError("Skill failed: " + result.Skill))
		}
		fmt.Println()

		if result.Message != "" {
			fmt.Println(display.FormatKeyValue("Message", result.Message))
		}

		if result.Output != "" {
			fmt.Println()
			fmt.Println(display.StyleLabel("Output:"))
			fmt.Println(result.Output)
		}

		return nil
	},
}

func init() {
	skillRunCmd.Flags().StringArrayVarP(&skillParams, "param", "p", nil, "Skill parameters as key=value (repeatable)")
	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillRunCmd)
	rootCmd.AddCommand(skillCmd)
}
