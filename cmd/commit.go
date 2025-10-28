package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit staged graph changes",
	Long:  `Commit the staged graph with a message and show diff`,
	Run: func(cmd *cobra.Command, args []string) {
		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			fmt.Println("Error: commit message is required. Use -m flag.")
			os.Exit(1)
		}

		if err := commitGraph(message); err != nil {
			fmt.Printf("Error committing graph: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	commitCmd.Flags().StringP("message", "m", "", "Commit message")
	rootCmd.AddCommand(commitCmd)
}

func commitGraph(message string) error {
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

	stagedGraph, exists := config["staged_graph"].(string)
	if !exists {
		return fmt.Errorf("no staged changes. Use 'tribal add -A' first")
	}

	stagedGraphFile, exists := config["staged_graph_file"].(string)
	if !exists {
		return fmt.Errorf("no staged graph file found")
	}

	// Read staged graph
	stagedData, err := ioutil.ReadFile(stagedGraphFile)
	if err != nil {
		return fmt.Errorf("failed to read staged graph: %w", err)
	}

	var stagedGraphData map[string]interface{}
	if err := json.Unmarshal(stagedData, &stagedGraphData); err != nil {
		return fmt.Errorf("failed to parse staged graph: %w", err)
	}

	// Create commit object
	commit := map[string]interface{}{
		"id":        generateCommitID(),
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
		"author":    "user", // TODO: get from git config
		"graph":     stagedGraphData,
	}

	// Create commits directory if it doesn't exist
	commitsDir := filepath.Join(".tribal", "commits")
	if err := os.MkdirAll(commitsDir, 0755); err != nil {
		return fmt.Errorf("failed to create commits directory: %w", err)
	}

	// Save commit
	commitID := commit["id"].(string)
	commitFile := filepath.Join(commitsDir, commitID+".json")
	commitData, err := json.MarshalIndent(commit, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize commit: %w", err)
	}

	if err := ioutil.WriteFile(commitFile, commitData, 0644); err != nil {
		return fmt.Errorf("failed to save commit: %w", err)
	}

	// Update config with latest commit
	config["latest_commit"] = commitID
	config["latest_commit_file"] = commitFile

	// Clear staging
	delete(config, "staged_graph")
	delete(config, "staged_graph_file")

	configData, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := ioutil.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Show commit summary
	fmt.Printf("Committed graph: %s\n", stagedGraph)
	fmt.Printf("Commit ID: %s\n", commitID)
	fmt.Printf("Message: %s\n", message)
	fmt.Printf("Timestamp: %s\n", commit["timestamp"])

	// Show simple diff (node/edge counts)
	nodes, _ := stagedGraphData["nodes"].([]interface{})
	edges, _ := stagedGraphData["edges"].([]interface{})
	fmt.Printf("\nGraph Summary:\n")
	fmt.Printf("  Nodes: %d\n", len(nodes))
	fmt.Printf("  Edges: %d\n", len(edges))

	fmt.Printf("\nCommit saved to: %s\n", commitFile)
	fmt.Println("\nReview this commit before pushing. Use 'tribal push' when ready.")

	return nil
}

func generateCommitID() string {
	// Simple timestamp-based ID for now
	return fmt.Sprintf("commit_%d", time.Now().Unix())
}