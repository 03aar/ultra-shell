package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

type checkResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Latency string `json:"latency,omitempty"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and dependencies",
	Long:  `Run diagnostic checks on all Nucleus services and dependencies.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		checks := runAllChecks()

		if output == "json" {
			data, _ := json.MarshalIndent(checks, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleHeader("Nucleus Doctor"))
		fmt.Println()

		allOk := true
		for _, c := range checks {
			var statusIcon string
			switch c.Status {
			case "ok":
				statusIcon = display.StyleSuccess(c.Name)
			case "warn":
				statusIcon = display.StyleWarning(c.Name)
				allOk = false
			case "fail":
				statusIcon = display.StyleError(c.Name)
				allOk = false
			}

			line := statusIcon
			if c.Latency != "" {
				line += "  " + display.StyleMuted("("+c.Latency+")")
			}
			if c.Message != "" {
				line += "  " + c.Message
			}
			fmt.Println(line)
		}

		fmt.Println()
		if allOk {
			fmt.Println(display.StyleSuccess("All checks passed"))
		} else {
			fmt.Println(display.StyleWarning("Some checks need attention"))
		}

		return nil
	},
}

func runAllChecks() []checkResult {
	checks := []checkResult{}

	// Check Nucleus API
	checks = append(checks, checkAPI())

	// Check WebSocket
	checks = append(checks, checkWebSocket())

	// Check git
	checks = append(checks, checkBinary("git"))

	// Check docker
	checks = append(checks, checkBinary("docker"))

	// Check shell tools
	checks = append(checks, checkBinary("bash"))
	checks = append(checks, checkBinary("curl"))

	return checks
}

func checkAPI() checkResult {
	start := time.Now()
	resp, err := http.Get(apiURL + "/api/health")
	latency := time.Since(start)

	if err != nil {
		return checkResult{
			Name:    "Nucleus API",
			Status:  "fail",
			Message: "Cannot connect to " + apiURL + ": " + err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return checkResult{
			Name:    "Nucleus API",
			Status:  "warn",
			Message: fmt.Sprintf("API returned HTTP %d", resp.StatusCode),
			Latency: latency.Round(time.Millisecond).String(),
		}
	}

	return checkResult{
		Name:    "Nucleus API",
		Status:  "ok",
		Message: apiURL,
		Latency: latency.Round(time.Millisecond).String(),
	}
}

func checkWebSocket() checkResult {
	wsURL := apiURL
	if len(wsURL) > 4 && wsURL[:4] == "http" {
		wsURL = "ws" + wsURL[4:]
	}
	wsURL += "/api/ws/watch"

	// Just verify the URL is reachable at HTTP level
	start := time.Now()
	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Get(apiURL + "/api/ws/watch")
	latency := time.Since(start)

	if err != nil {
		return checkResult{
			Name:    "WebSocket",
			Status:  "warn",
			Message: "WebSocket endpoint not reachable",
		}
	}
	defer resp.Body.Close()

	// A 400 or 426 is expected when connecting via HTTP instead of WS
	return checkResult{
		Name:    "WebSocket",
		Status:  "ok",
		Message: wsURL,
		Latency: latency.Round(time.Millisecond).String(),
	}
}

func checkBinary(name string) checkResult {
	path, err := exec.LookPath(name)
	if err != nil {
		return checkResult{
			Name:    name,
			Status:  "fail",
			Message: "not found in PATH",
		}
	}

	return checkResult{
		Name:    name,
		Status:  "ok",
		Message: path,
	}
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
