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

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for graphs by semantic similarity",
	Long:  `Search for graphs by title and description using semantic similarity matching`,
	Run: func(cmd *cobra.Command, args []string) {
		context, _ := cmd.Flags().GetString("context")
		if context == "" {
			fmt.Println("Error: context description is required. Use --context flag.")
			os.Exit(1)
		}

		if err := searchGraphs(context); err != nil {
			fmt.Printf("Error searching graphs: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	searchCmd.Flags().String("context", "", "Context description for semantic search")
	rootCmd.AddCommand(searchCmd)
}

func searchGraphs(context string) error {
	// Check if .tribal exists
	if _, err := os.Stat(".tribal"); os.IsNotExist(err) {
		return fmt.Errorf("not a tribal repository. Run 'tribal init' first")
	}

	graphsDir := filepath.Join(".tribal", "graphs")
	if _, err := os.Stat(graphsDir); os.IsNotExist(err) {
		fmt.Println("No graphs found.")
		return nil
	}

	// Read all graph files
	files, err := ioutil.ReadDir(graphsDir)
	if err != nil {
		return fmt.Errorf("failed to read graphs directory: %w", err)
	}

	fmt.Printf("Searching for graphs matching: %s\n\n", context)

	contextLower := strings.ToLower(context)
	matches := []map[string]interface{}{}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		graphPath := filepath.Join(graphsDir, file.Name())
		graphData, err := ioutil.ReadFile(graphPath)
		if err != nil {
			continue
		}

		var graph map[string]interface{}
		if err := json.Unmarshal(graphData, &graph); err != nil {
			continue
		}

		title, _ := graph["title"].(string)
		titleLower := strings.ToLower(title)

		// Simple keyword matching (could be enhanced with proper semantic search)
		score := 0
		contextWords := strings.Fields(contextLower)
		titleWords := strings.Fields(titleLower)

		for _, contextWord := range contextWords {
			for _, titleWord := range titleWords {
				if strings.Contains(titleWord, contextWord) || strings.Contains(contextWord, titleWord) {
					score++
				}
			}
		}

		// Check metadata for additional context
		if metadata, ok := graph["metadata"].(map[string]interface{}); ok {
			if description, exists := metadata["description"].(string); exists {
				descLower := strings.ToLower(description)
				for _, contextWord := range contextWords {
					if strings.Contains(descLower, contextWord) {
						score += 2 // Weight description matches higher
					}
				}
			}
		}

		if score > 0 {
			graph["_score"] = score
			graph["_filename"] = file.Name()
			matches = append(matches, graph)
		}
	}

	if len(matches) == 0 {
		fmt.Println("No matching graphs found.")
		return nil
	}

	// Sort by score (simple bubble sort for now)
	for i := 0; i < len(matches)-1; i++ {
		for j := 0; j < len(matches)-i-1; j++ {
			scoreA, _ := matches[j]["_score"].(int)
			scoreB, _ := matches[j+1]["_score"].(int)
			if scoreA < scoreB {
				matches[j], matches[j+1] = matches[j+1], matches[j]
			}
		}
	}

	// Display results
	fmt.Printf("Found %d matching graphs:\n\n", len(matches))
	for i, match := range matches {
		title, _ := match["title"].(string)
		score, _ := match["_score"].(int)
		filename, _ := match["_filename"].(string)

		fmt.Printf("%d. %s (score: %d)\n", i+1, title, score)
		fmt.Printf("   File: %s\n", filename)

		if metadata, ok := match["metadata"].(map[string]interface{}); ok {
			if description, exists := metadata["description"].(string); exists {
				fmt.Printf("   Description: %s\n", description)
			}
		}
		fmt.Println()
	}

	return nil
}