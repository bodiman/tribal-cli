package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Stage graph changes",
	Long:  `Stage the current graph for commit`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		if !all {
			fmt.Println("Error: currently only supports -A flag to stage all changes")
			os.Exit(1)
		}

		if err := stageGraph(); err != nil {
			fmt.Printf("Error staging graph: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	addCmd.Flags().BoolP("all", "A", false, "Stage all graph changes")
	rootCmd.AddCommand(addCmd)
}

func stageGraph() error {
	// Check if .tribal exists
	if _, err := os.Stat(".tribal"); os.IsNotExist(err) {
		return fmt.Errorf("not a tribal repository. Run 'tribal init' first")
	}

	configPath := filepath.Join(".tribal", "config.json")
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	currentGraph, exists := config["current_graph"].(string)
	if !exists {
		return fmt.Errorf("no current graph checked out. Use 'tribal checkout -g\"<title>\"' first")
	}

	currentGraphFile, exists := config["current_graph_file"].(string)
	if !exists {
		return fmt.Errorf("no current graph file found in config")
	}

	// Check if graph file exists
	if _, err := os.Stat(currentGraphFile); os.IsNotExist(err) {
		return fmt.Errorf("current graph file does not exist: %s", currentGraphFile)
	}

	// Create staging area if it doesn't exist
	stagingDir := filepath.Join(".tribal", "staging")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	// Copy current graph to staging area
	graphData, err := ioutil.ReadFile(currentGraphFile)
	if err != nil {
		return fmt.Errorf("failed to read current graph: %w", err)
	}

	stagingFile := filepath.Join(stagingDir, filepath.Base(currentGraphFile))
	if err := ioutil.WriteFile(stagingFile, graphData, 0644); err != nil {
		return fmt.Errorf("failed to stage graph: %w", err)
	}

	// Update config with staged changes
	config["staged_graph"] = currentGraph
	config["staged_graph_file"] = stagingFile

	configData, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := ioutil.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	fmt.Printf("Staged graph: %s\n", currentGraph)
	fmt.Printf("Staged file: %s\n", stagingFile)

	return nil
}