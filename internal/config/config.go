package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	Version     string `json:"version"`
	Remote      string `json:"remote"`
	Graphs      map[string]interface{} `json:"graphs"`
	CurrentGraph string `json:"current_graph,omitempty"`
	CurrentGraphFile string `json:"current_graph_file,omitempty"`
	LatestCommit string `json:"latest_commit,omitempty"`
	LatestCommitFile string `json:"latest_commit_file,omitempty"`
	LastPushedCommit string `json:"last_pushed_commit,omitempty"`
	LastPushedFile string `json:"last_pushed_file,omitempty"`
	// Network configuration
	RegistryURL string `json:"registry_url,omitempty"`
	Token       string `json:"token,omitempty"`
	Username    string `json:"username,omitempty"`
	UserID      string `json:"user_id,omitempty"`
}

const (
	ConfigDir  = ".tribal"
	ConfigFile = "config.json"
	DefaultRegistryURL = "http://localhost:8080"
)

func GetConfigPath() string {
	return filepath.Join(ConfigDir, ConfigFile)
}

func Load() (*Config, error) {
	configPath := GetConfigPath()
	
	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a tribal repository. Run 'tribal init' first")
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set default registry URL if not set
	if config.RegistryURL == "" {
		config.RegistryURL = DefaultRegistryURL
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath := GetConfigPath()
	
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func (c *Config) IsAuthenticated() bool {
	return c.Token != "" && c.Username != ""
}

func (c *Config) SetAuth(token, username, userID string) {
	c.Token = token
	c.Username = username
	c.UserID = userID
}

func (c *Config) ClearAuth() {
	c.Token = ""
	c.Username = ""
	c.UserID = ""
}

func (c *Config) SetRegistryURL(url string) {
	c.RegistryURL = url
}

func CreateDefaultConfig() *Config {
	return &Config{
		Version:     "1.0.0",
		Remote:      "",
		Graphs:      make(map[string]interface{}),
		RegistryURL: DefaultRegistryURL,
	}
}