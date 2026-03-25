package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nucleus-cli/internal/api"
	"nucleus-cli/internal/display"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch live execution stream",
	Long:  `Connect to the Nucleus WebSocket and display executions in real-time.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		wsURL := client.WebSocketURL()

		fmt.Println(display.StyleHeader("Live Watch"))
		fmt.Println(display.StyleMuted("Connecting to " + wsURL + "..."))
		fmt.Println()

		header := http.Header{}
		header.Set("X-API-Key", client.APIKey())

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		if err != nil {
			return fmt.Errorf("websocket connection failed: %w", err)
		}
		defer conn.Close()

		fmt.Println(display.StyleSuccess("Connected. Watching for events... (Ctrl+C to stop)"))
		fmt.Println()

		// Handle graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		doneChan := make(chan struct{})

		go func() {
			defer close(doneChan)
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						fmt.Fprintln(os.Stderr, display.StyleError("WebSocket error: "+err.Error()))
					}
					return
				}

				var event api.WatchEvent
				if jsonErr := json.Unmarshal(message, &event); jsonErr != nil {
					fmt.Fprintln(os.Stderr, display.StyleWarning("Failed to parse event: "+jsonErr.Error()))
					continue
				}

				renderWatchEvent(event)
			}
		}()

		select {
		case <-sigChan:
			fmt.Println()
			fmt.Println(display.StyleMuted("Disconnecting..."))
			closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
			conn.WriteMessage(websocket.CloseMessage, closeMsg)
			return nil
		case <-doneChan:
			return nil
		}
	},
}

func renderWatchEvent(event api.WatchEvent) {
	e := event.Data
	timestamp := event.Timestamp.Format(time.Kitchen)

	fmt.Printf("%s %s %s\n",
		display.StyleMuted("["+timestamp+"]"),
		display.FormatRiskBadge(e.RiskLevel),
		display.StyleValue(e.Command),
	)

	fmt.Printf("  Exit: %s  Duration: %.2fs  Reversible: %s\n",
		display.FormatExitCode(e.ExitCode),
		e.Duration,
		display.FormatBool(e.Reversible),
	)

	if e.Output != "" {
		outputPreview := truncate(e.Output, 120)
		fmt.Printf("  %s\n", display.StyleMuted(outputPreview))
	}
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
