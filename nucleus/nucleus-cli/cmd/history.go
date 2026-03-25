package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var (
	historyLimit   int
	historySearch  string
	historySession string
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show execution history",
	Long:  `Display past command executions with filtering options.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		executions, err := client.GetExecutions(historyLimit, historySearch, historySession)
		if err != nil {
			return fmt.Errorf("failed to fetch history: %w", err)
		}

		if output == "json" {
			data, _ := json.MarshalIndent(executions, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleHeader("Execution History"))
		fmt.Println()

		if len(executions) == 0 {
			fmt.Println(display.StyleMuted("No executions found."))
			return nil
		}

		headers := []string{"ID", "TIME", "COMMAND", "EXIT", "RISK", "DURATION", "REVERSIBLE"}
		rows := make([][]string, len(executions))
		for i, e := range executions {
			rev := "no"
			if e.Reversible {
				rev = "yes"
			}
			rows[i] = []string{
				e.ID[:8],
				e.Timestamp.Format(time.RFC3339)[:19],
				truncate(e.Command, 40),
				fmt.Sprintf("%d", e.ExitCode),
				e.RiskLevel,
				fmt.Sprintf("%.2fs", e.Duration),
				rev,
			}
		}
		fmt.Print(display.FormatTable(headers, rows))

		fmt.Println()
		fmt.Println(display.StyleMuted(fmt.Sprintf("Showing %d executions", len(executions))))

		return nil
	},
}

func init() {
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 20, "Maximum number of entries to show")
	historyCmd.Flags().StringVarP(&historySearch, "search", "s", "", "Filter by command text")
	historyCmd.Flags().StringVar(&historySession, "session", "", "Filter by session ID")
	rootCmd.AddCommand(historyCmd)
}
