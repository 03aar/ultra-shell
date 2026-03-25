package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var mcpClient string

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP server integration",
	Long:  `Install and manage Model Context Protocol server configuration for AI clients.`,
}

var mcpInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install MCP server configuration",
	Long:  `Install the Nucleus MCP server configuration for a supported AI client.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if mcpClient != "claude-desktop" && mcpClient != "cursor" {
			return fmt.Errorf("unsupported client %q, supported: claude-desktop, cursor", mcpClient)
		}

		configPath, err := getMCPConfigPath(mcpClient)
		if err != nil {
			return err
		}

		mcpConfig := buildMCPConfig()

		// Read existing config or create new
		existing := make(map[string]interface{})
		if data, err := os.ReadFile(configPath); err == nil {
			if jsonErr := json.Unmarshal(data, &existing); jsonErr != nil {
				return fmt.Errorf("existing config is invalid JSON: %w", jsonErr)
			}
		}

		// Merge nucleus server into mcpServers
		servers, ok := existing["mcpServers"].(map[string]interface{})
		if !ok {
			servers = make(map[string]interface{})
		}
		servers["nucleus"] = mcpConfig
		existing["mcpServers"] = servers

		data, err := json.MarshalIndent(existing, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		if output == "json" {
			result := map[string]interface{}{
				"client": mcpClient,
				"path":   configPath,
				"config": mcpConfig,
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(out))
			return nil
		}

		fmt.Println(display.StyleSuccess("MCP server installed for " + mcpClient))
		fmt.Println()
		fmt.Println(display.FormatKeyValue("Config Path", configPath))
		fmt.Println(display.FormatKeyValue("Server", "nucleus"))
		fmt.Println(display.FormatKeyValue("API URL", apiURL))
		fmt.Println()
		fmt.Println(display.StyleMuted("Restart " + mcpClient + " to activate the MCP server."))

		return nil
	},
}

var mcpStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check MCP installation status",
	RunE: func(cmd *cobra.Command, args []string) error {
		clients := []string{"claude-desktop", "cursor"}
		results := make([]map[string]string, 0)

		for _, c := range clients {
			configPath, err := getMCPConfigPath(c)
			if err != nil {
				continue
			}

			status := "not installed"
			data, err := os.ReadFile(configPath)
			if err == nil {
				var config map[string]interface{}
				if json.Unmarshal(data, &config) == nil {
					if servers, ok := config["mcpServers"].(map[string]interface{}); ok {
						if _, exists := servers["nucleus"]; exists {
							status = "installed"
						}
					}
				}
			}

			results = append(results, map[string]string{
				"client": c,
				"path":   configPath,
				"status": status,
			})
		}

		if output == "json" {
			data, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleHeader("MCP Status"))
		fmt.Println()

		headers := []string{"CLIENT", "STATUS", "CONFIG PATH"}
		rows := make([][]string, len(results))
		for i, r := range results {
			statusText := r["status"]
			if statusText == "installed" {
				statusText = display.StyleSuccess("installed")
			} else {
				statusText = display.StyleMuted("not installed")
			}
			rows[i] = []string{r["client"], statusText, r["path"]}
		}
		fmt.Print(display.FormatTable(headers, rows))

		return nil
	},
}

func getMCPConfigPath(clientName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	switch clientName {
	case "claude-desktop":
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
		case "linux":
			return filepath.Join(home, ".config", "claude", "claude_desktop_config.json"), nil
		case "windows":
			return filepath.Join(home, "AppData", "Roaming", "Claude", "claude_desktop_config.json"), nil
		}
	case "cursor":
		switch runtime.GOOS {
		case "darwin":
			return filepath.Join(home, ".cursor", "mcp.json"), nil
		case "linux":
			return filepath.Join(home, ".cursor", "mcp.json"), nil
		case "windows":
			return filepath.Join(home, ".cursor", "mcp.json"), nil
		}
	}

	return "", fmt.Errorf("unsupported client/OS combination: %s/%s", clientName, runtime.GOOS)
}

func buildMCPConfig() map[string]interface{} {
	return map[string]interface{}{
		"command": "nuc",
		"args":    []string{"mcp-serve", "--api-url", apiURL},
		"env": map[string]string{
			"NUCLEUS_API_KEY": apiKey,
		},
	}
}

func init() {
	mcpInstallCmd.Flags().StringVar(&mcpClient, "client", "claude-desktop", "Target client: claude-desktop or cursor")
	mcpCmd.AddCommand(mcpInstallCmd)
	mcpCmd.AddCommand(mcpStatusCmd)
	rootCmd.AddCommand(mcpCmd)
}
