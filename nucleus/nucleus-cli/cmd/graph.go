package cmd

import (
	"encoding/json"
	"fmt"

	"nucleus-cli/internal/display"

	"github.com/spf13/cobra"
)

var (
	graphFormat  string
	graphSession string
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Display the execution dependency graph",
	Long:  `Show the command execution graph in mermaid or JSON format.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		graph, err := client.GetGraph(graphSession, graphFormat)
		if err != nil {
			return fmt.Errorf("failed to fetch graph: %w", err)
		}

		if output == "json" || graphFormat == "json" {
			data, _ := json.MarshalIndent(graph, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println(display.StyleHeader("Execution Graph"))
		fmt.Println()

		if graph.Content != "" {
			fmt.Println(graph.Content)
			return nil
		}

		if len(graph.Nodes) == 0 {
			fmt.Println(display.StyleMuted("No graph data available."))
			return nil
		}

		// Render a text representation of the graph
		fmt.Println(display.StyleLabel("Nodes:"))
		for _, node := range graph.Nodes {
			fmt.Printf("  [%s] %s (%s)\n", node.ID, node.Label, display.StyleMuted(node.Type))
		}

		if len(graph.Edges) > 0 {
			fmt.Println()
			fmt.Println(display.StyleLabel("Edges:"))
			for _, edge := range graph.Edges {
				label := ""
				if edge.Label != "" {
					label = " (" + edge.Label + ")"
				}
				fmt.Printf("  %s → %s%s\n", edge.Source, edge.Target, display.StyleMuted(label))
			}
		}

		return nil
	},
}

func init() {
	graphCmd.Flags().StringVarP(&graphFormat, "format", "f", "mermaid", "Output format: mermaid or json")
	graphCmd.Flags().StringVar(&graphSession, "session", "", "Session ID to graph")
	rootCmd.AddCommand(graphCmd)
}
