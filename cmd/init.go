package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/tribal/tribal-cli/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a tribal repository",
	Long:  `Initialize a tribal repository by creating TRIBAL.md file and updating CLAUDE.md`,
	Run: func(cmd *cobra.Command, args []string) {
		registryURL, _ := cmd.Flags().GetString("registry")
		if err := initRepository(registryURL); err != nil {
			fmt.Printf("Error initializing repository: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully initialized tribal repository")
		
		if registryURL != "" {
			fmt.Printf("Registry URL set to: %s\n", registryURL)
			fmt.Println("Use 'tribal login' to authenticate with the registry")
		}
	},
}

func init() {
	initCmd.Flags().StringP("registry", "r", "", "Registry URL (default: http://localhost:8080)")
	rootCmd.AddCommand(initCmd)
}

func initRepository(registryURL string) error {
	// Create TRIBAL.md file
	tribalContent := `# Tribal Framework

This repository uses the Tribal framework for graph-based development.

## Overview

Tribal is a framework for organizing and managing development work through graph structures.
Each graph represents a specific feature, component, or concept within the codebase.

## Usage

- Use 'tribal checkout -g"<graph title>"' to create or retrieve a graph
- Use 'tribal search --context <description>' to find related graphs
- Use 'tribal add -A' to stage your graph changes
- Use 'tribal commit -m"<message>"' to commit your graph
- Use 'tribal push' to push changes to the remote

## Graph Structure

Graphs consist of nodes and edges that represent relationships between code components,
features, or concepts. Each node can contain markup for documentation and context.
`

	if err := ioutil.WriteFile("TRIBAL.md", []byte(tribalContent), 0644); err != nil {
		return fmt.Errorf("failed to create TRIBAL.md: %w", err)
	}

	// Update or create CLAUDE.md
	claudeContent := `This repository is developed using the TRIBAL framework. Whenever planning or making changes to the codebase, be sure to first consult TRIBAL.md in the root repository for reference.

`

	claudePath := "CLAUDE.md"
	if _, err := os.Stat(claudePath); err == nil {
		// File exists, prepend to existing content
		existing, err := ioutil.ReadFile(claudePath)
		if err != nil {
			return fmt.Errorf("failed to read existing CLAUDE.md: %w", err)
		}
		claudeContent += string(existing)
	}

	if err := ioutil.WriteFile(claudePath, []byte(claudeContent), 0644); err != nil {
		return fmt.Errorf("failed to update CLAUDE.md: %w", err)
	}

	// Create .tribal directory for internal storage
	tribalDir := ".tribal"
	if err := os.MkdirAll(tribalDir, 0755); err != nil {
		return fmt.Errorf("failed to create .tribal directory: %w", err)
	}

	// Create initial config using the new config system
	cfg := config.CreateDefaultConfig()
	if registryURL != "" {
		cfg.SetRegistryURL(registryURL)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	return nil
}