package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Open the Nucleus dashboard in a browser",
	Long:  `Open the Nucleus web dashboard in your default browser.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dashboardURL := apiURL + "/dashboard"

		fmt.Println(display.StyleHeader("Nucleus Dashboard"))
		fmt.Println()
		fmt.Println(display.FormatKeyValue("URL", dashboardURL))

		if err := openBrowser(dashboardURL); err != nil {
			fmt.Println()
			fmt.Println(display.StyleWarning("Could not open browser automatically: " + err.Error()))
			fmt.Println(display.StyleMuted("Please open the URL above manually."))
			return nil
		}

		fmt.Println()
		fmt.Println(display.StyleSuccess("Dashboard opened in browser"))

		return nil
	},
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
