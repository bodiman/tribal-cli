package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push committed graph changes",
	Long:  `Push the latest commit to the remote repository`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := pushGraph(); err != nil {
			fmt.Printf("Error pushing graph: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func pushGraph() error {
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

	latestCommit, exists := config["latest_commit"].(string)
	if !exists {
		return fmt.Errorf("no commits to push. Use 'tribal commit -m\"<message>\"' first")
	}

	latestCommitFile, exists := config["latest_commit_file"].(string)
	if !exists {
		return fmt.Errorf("no commit file found")
	}

	// Check if commit file exists
	if _, err := os.Stat(latestCommitFile); os.IsNotExist(err) {
		return fmt.Errorf("commit file does not exist: %s", latestCommitFile)
	}

	// Read commit data
	commitData, err := ioutil.ReadFile(latestCommitFile)
	if err != nil {
		return fmt.Errorf("failed to read commit: %w", err)
	}

	var commit map[string]interface{}
	if err := json.Unmarshal(commitData, &commit); err != nil {
		return fmt.Errorf("failed to parse commit: %w", err)
	}

	// For now, simulate push by creating a "pushed" directory
	// In a real implementation, this would send to a remote server
	pushedDir := filepath.Join(".tribal", "pushed")
	if err := os.MkdirAll(pushedDir, 0755); err != nil {
		return fmt.Errorf("failed to create pushed directory: %w", err)
	}

	pushedFile := filepath.Join(pushedDir, filepath.Base(latestCommitFile))
	if err := ioutil.WriteFile(pushedFile, commitData, 0644); err != nil {
		return fmt.Errorf("failed to push commit: %w", err)
	}

	// Update config to mark as pushed
	config["last_pushed_commit"] = latestCommit
	config["last_pushed_file"] = pushedFile

	configData, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := ioutil.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Show push summary
	fmt.Printf("Pushed commit: %s\n", latestCommit)
	fmt.Printf("Message: %s\n", commit["message"])
	fmt.Printf("Timestamp: %s\n", commit["timestamp"])
	fmt.Printf("Pushed to: %s\n", pushedFile)

	// Show graph summary
	if graph, ok := commit["graph"].(map[string]interface{}); ok {
		if title, exists := graph["title"].(string); exists {
			fmt.Printf("Graph: %s\n", title)
		}
		if nodes, ok := graph["nodes"].([]interface{}); ok {
			fmt.Printf("Nodes: %d\n", len(nodes))
		}
		if edges, ok := graph["edges"].([]interface{}); ok {
			fmt.Printf("Edges: %d\n", len(edges))
		}
	}

	fmt.Println("\nGraph successfully pushed!")
	fmt.Println("Note: This is a local simulation. In production, this would sync with a remote tribal server.")

	return nil
}