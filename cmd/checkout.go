package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Create or retrieve a graph",
	Long:  `Create a new graph or retrieve an existing graph by title`,
	Run: func(cmd *cobra.Command, args []string) {
		graphTitle, _ := cmd.Flags().GetString("graph")
		if graphTitle == "" {
			fmt.Println("Error: graph title is required. Use -g flag to specify graph title.")
			os.Exit(1)
		}

		if err := checkoutGraph(graphTitle); err != nil {
			fmt.Printf("Error checking out graph: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	checkoutCmd.Flags().StringP("graph", "g", "", "Graph title to checkout")
	rootCmd.AddCommand(checkoutCmd)
}

func checkoutGraph(title string) error {
	configPath := filepath.Join(".tribal", "config.json")
	
	// Check if .tribal exists
	if _, err := os.Stat(".tribal"); os.IsNotExist(err) {
		return fmt.Errorf("not a tribal repository. Run 'tribal init' first")
	}

	// Read current config
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	graphs, ok := config["graphs"].(map[string]interface{})
	if !ok {
		graphs = make(map[string]interface{})
		config["graphs"] = graphs
	}

	// Create graph filename from title
	filename := strings.ReplaceAll(strings.ToLower(title), " ", "_") + ".json"
	graphPath := filepath.Join(".tribal", "graphs", filename)

	// Create graphs directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join(".tribal", "graphs"), 0755); err != nil {
		return fmt.Errorf("failed to create graphs directory: %w", err)
	}

	// Check if graph already exists
	if _, err := os.Stat(graphPath); err == nil {
		fmt.Printf("Checked out existing graph: %s\n", title)
		fmt.Printf("Graph file: %s\n", graphPath)
	} else {
		// Create new graph
		newGraph := map[string]interface{}{
			"title":    title,
			"nodes":    []interface{}{},
			"edges":    []interface{}{},
			"metadata": map[string]interface{}{
				"created": "now", // TODO: use proper timestamp
				"author":  "user", // TODO: get from git config
			},
		}

		graphData, err := json.MarshalIndent(newGraph, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to serialize graph: %w", err)
		}

		if err := ioutil.WriteFile(graphPath, graphData, 0644); err != nil {
			return fmt.Errorf("failed to create graph file: %w", err)
		}

		fmt.Printf("Created new graph: %s\n", title)
		fmt.Printf("Graph file: %s\n", graphPath)
	}

	// Update current graph in config
	config["current_graph"] = title
	config["current_graph_file"] = graphPath

	// Save updated config
	configData, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := ioutil.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return nil
}