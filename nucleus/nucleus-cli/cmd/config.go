package cmd

import (
	"encoding/json"
	"fmt"
	"nucleus-cli/internal/display"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type NucleusConfig struct {
	API struct {
		URL  string `json:"url"`
		Key  string `json:"key"`
		Port int    `json:"port"`
	} `json:"api"`
	Dashboard struct {
		URL  string `json:"url"`
		Port int    `json:"port"`
	} `json:"dashboard"`
	LLM struct {
		Provider    string  `json:"provider"`
		Model       string  `json:"model"`
		Temperature float64 `json:"temperature"`
	} `json:"llm"`
	Shell struct {
		Binary          string `json:"binary"`
		AnnotationLevel string `json:"annotation_level"`
		NaturalLanguage bool   `json:"natural_language"`
	} `json:"shell"`
	Safety struct {
		AutoApproveRisk       string `json:"auto_approve_risk"`
		BlockCritical         bool   `json:"block_critical"`
		RequireConfirmAbove   string `json:"require_confirm_above"`
	} `json:"safety"`
}

func defaultConfig() NucleusConfig {
	cfg := NucleusConfig{}
	cfg.API.URL = "http://localhost:8080"
	cfg.API.Key = "dev-nucleus-key-local"
	cfg.API.Port = 8080
	cfg.Dashboard.URL = "http://localhost:3000"
	cfg.Dashboard.Port = 3000
	cfg.LLM.Provider = "claude"
	cfg.LLM.Model = ""
	cfg.LLM.Temperature = 0.1
	cfg.Shell.Binary = ""
	cfg.Shell.AnnotationLevel = "normal"
	cfg.Shell.NaturalLanguage = true
	cfg.Safety.AutoApproveRisk = "low"
	cfg.Safety.BlockCritical = true
	cfg.Safety.RequireConfirmAbove = "medium"
	return cfg
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "nucleus", "config.json")
}

func loadConfig() NucleusConfig {
	cfg := defaultConfig()
	data, err := os.ReadFile(configPath())
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	return cfg
}

func saveConfig(cfg NucleusConfig) error {
	path := configPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or edit Nucleus configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadConfig()

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(cfg, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println()
		fmt.Println(display.StyleHeader("  Nucleus Configuration"))
		fmt.Printf("  %s %s\n\n", display.StyleMuted("File:"), configPath())

		fmt.Println(display.StyleMuted("  API:"))
		fmt.Printf("    url:  %s\n", cfg.API.URL)
		fmt.Printf("    key:  %s\n", maskKey(cfg.API.Key))
		fmt.Printf("    port: %d\n", cfg.API.Port)

		fmt.Println(display.StyleMuted("  LLM:"))
		fmt.Printf("    provider:    %s\n", cfg.LLM.Provider)
		fmt.Printf("    model:       %s\n", orDefault(cfg.LLM.Model, "(auto)"))
		fmt.Printf("    temperature: %.1f\n", cfg.LLM.Temperature)

		fmt.Println(display.StyleMuted("  Shell:"))
		fmt.Printf("    binary:      %s\n", orDefault(cfg.Shell.Binary, "(auto-detect)"))
		fmt.Printf("    annotations: %s\n", cfg.Shell.AnnotationLevel)
		fmt.Printf("    natural_lang:%v\n", cfg.Shell.NaturalLanguage)

		fmt.Println(display.StyleMuted("  Safety:"))
		fmt.Printf("    auto_approve: %s\n", cfg.Safety.AutoApproveRisk)
		fmt.Printf("    block_critical: %v\n", cfg.Safety.BlockCritical)
		fmt.Println()

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long: `Set a Nucleus configuration value.

Examples:
  nuc config set api.port 8080
  nuc config set llm.provider claude
  nuc config set llm.model claude-sonnet-4-5
  nuc config set safety.auto_approve_risk medium
  nuc config set shell.natural_language true`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		cfg := loadConfig()

		switch key {
		case "api.url":
			cfg.API.URL = value
		case "api.key":
			cfg.API.Key = value
		case "llm.provider":
			cfg.LLM.Provider = value
		case "llm.model":
			cfg.LLM.Model = value
		case "shell.binary":
			cfg.Shell.Binary = value
		case "shell.annotation_level":
			cfg.Shell.AnnotationLevel = value
		case "shell.natural_language":
			cfg.Shell.NaturalLanguage = strings.ToLower(value) == "true"
		case "safety.auto_approve_risk":
			cfg.Safety.AutoApproveRisk = value
		case "safety.block_critical":
			cfg.Safety.BlockCritical = strings.ToLower(value) == "true"
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}

		if err := saveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("%s Set %s = %s\n", display.StyleSuccess("✓"), key, value)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
