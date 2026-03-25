package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
	Long:  `List, create, and replay shell sessions.`,
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := client.GetSessions()
		if err != nil {
			return fmt.Errorf("failed to fetch sessions: %w", err)
		}

		if output == "json" {
			data, _ := json.MarshalIndent(sessions, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleHeader("Sessions"))
		fmt.Println()

		if len(sessions) == 0 {
			fmt.Println(display.StyleMuted("No sessions found."))
			return nil
		}

		headers := []string{"ID", "NAME", "COMMANDS", "CREATED", "UPDATED", "ACTIVE"}
		rows := make([][]string, len(sessions))
		for i, s := range sessions {
			active := "no"
			if s.Active {
				active = "yes"
			}
			rows[i] = []string{
				s.ID[:8],
				s.Name,
				fmt.Sprintf("%d", s.Commands),
				s.CreatedAt.Format(time.RFC3339)[:19],
				s.UpdatedAt.Format(time.RFC3339)[:19],
				active,
			}
		}
		fmt.Print(display.FormatTable(headers, rows))

		return nil
	},
}

var sessionNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := client.CreateSession(args[0])
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		if output == "json" {
			data, _ := json.MarshalIndent(session, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleSuccess("Session created"))
		fmt.Println()
		fmt.Println(display.FormatKeyValue("ID", session.ID))
		fmt.Println(display.FormatKeyValue("Name", session.Name))
		fmt.Println(display.FormatKeyValue("Created", session.CreatedAt.Format(time.RFC3339)))

		return nil
	},
}

var sessionReplayCmd = &cobra.Command{
	Use:   "replay <session-id>",
	Short: "Replay a session's execution history",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		executions, err := client.GetSessionReplay(args[0])
		if err != nil {
			return fmt.Errorf("failed to fetch session replay: %w", err)
		}

		if output == "json" {
			data, _ := json.MarshalIndent(executions, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleHeader("Session Replay: " + args[0][:8]))
		fmt.Println()

		if len(executions) == 0 {
			fmt.Println(display.StyleMuted("No executions in this session."))
			return nil
		}

		for i, e := range executions {
			fmt.Printf("%s %s\n",
				display.StyleLabel(fmt.Sprintf("[%d]", i+1)),
				display.StyleValue(e.Command),
			)
			fmt.Printf("    %s  Exit: %s  Risk: %s  Duration: %.2fs\n",
				display.StyleMuted(e.Timestamp.Format(time.Kitchen)),
				display.FormatExitCode(e.ExitCode),
				display.FormatRiskBadge(e.RiskLevel),
				e.Duration,
			)
			if e.Output != "" {
				lines := truncate(e.Output, 200)
				fmt.Printf("    %s\n", display.StyleMuted(lines))
			}
			fmt.Println()
		}

		fmt.Println(display.StyleMuted(fmt.Sprintf("Total: %d commands", len(executions))))

		return nil
	},
}

func init() {
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionNewCmd)
	sessionCmd.AddCommand(sessionReplayCmd)
	rootCmd.AddCommand(sessionCmd)
}
