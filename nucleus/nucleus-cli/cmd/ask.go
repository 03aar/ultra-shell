package cmd

import (
	"encoding/json"
	"fmt"
	"nucleus-cli/internal/api"
	"nucleus-cli/internal/display"
	"strings"

	"github.com/spf13/cobra"
)

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask Nucleus a question in natural language",
	Long: `Ask Nucleus a question and get a shell command translation.

Examples:
  nuc ask "how do I find all files larger than 100MB?"
  nuc ask "what changed recently in git?"
  nuc ask "who is using port 3000?"
  nuc ask "why did my build fail?"
  nuc ask "undo my last change"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		question := strings.Join(args, " ")
		client := api.NewClient(apiURL, apiKey)

		body := map[string]interface{}{
			"input": question,
		}
		jsonBody, _ := json.Marshal(body)

		resp, err := client.PostRaw("/natural/translate", jsonBody)
		if err != nil {
			return fmt.Errorf("failed to translate: %w", err)
		}

		var result struct {
			Command     string `json:"command"`
			Explanation string `json:"explanation"`
			RiskLevel   string `json:"risk_level"`
		}

		data, _ := json.Marshal(resp)
		json.Unmarshal(data, &result)

		if outputFormat == "json" {
			jsonOut, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(jsonOut))
			return nil
		}

		fmt.Println()
		fmt.Println(display.StyleHeader("  Nucleus Answer"))
		fmt.Println()
		fmt.Printf("  %s %s\n", display.StyleMuted("Question:"), question)
		fmt.Printf("  %s %s\n", display.StyleMuted("Command:"), display.StyleSuccess(result.Command))
		fmt.Printf("  %s %s\n", display.StyleMuted("Why:"), result.Explanation)
		fmt.Printf("  %s %s\n", display.StyleMuted("Risk:"), display.FormatRiskBadge(result.RiskLevel))
		fmt.Println()
		fmt.Printf("  %s\n", display.StyleMuted("Run with: nuc exec \""+result.Command+"\""))
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
