package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tribal/tribal-cli/internal/client"
	"github.com/tribal/tribal-cli/internal/config"
	"golang.org/x/crypto/ssh/terminal"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to tribal registry",
	Long:  `Authenticate with the tribal registry server`,
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		registryURL, _ := cmd.Flags().GetString("registry")

		if err := loginToRegistry(username, password, registryURL); err != nil {
			fmt.Printf("Error logging in: %v\n", err)
			os.Exit(1)
		}
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from tribal registry",
	Long:  `Clear authentication credentials`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := logoutFromRegistry(); err != nil {
			fmt.Printf("Error logging out: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully logged out from tribal registry")
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new account",
	Long:  `Create a new account on the tribal registry`,
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		registryURL, _ := cmd.Flags().GetString("registry")

		if err := registerAccount(username, email, password, registryURL); err != nil {
			fmt.Printf("Error registering: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Login command flags
	loginCmd.Flags().StringP("username", "u", "", "Username")
	loginCmd.Flags().StringP("password", "p", "", "Password (will prompt if not provided)")
	loginCmd.Flags().StringP("registry", "r", "", "Registry URL (uses config default if not provided)")

	// Register command flags
	registerCmd.Flags().StringP("username", "u", "", "Username")
	registerCmd.Flags().StringP("email", "e", "", "Email address")
	registerCmd.Flags().StringP("password", "p", "", "Password (will prompt if not provided)")
	registerCmd.Flags().StringP("registry", "r", "", "Registry URL (uses config default if not provided)")

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(registerCmd)
}

func getCredentials(username, password string) (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Get username if not provided
	if username == "" {
		fmt.Print("Username: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", "", fmt.Errorf("failed to read username: %w", err)
		}
		username = strings.TrimSpace(input)
	}

	// Get password if not provided
	if password == "" {
		fmt.Print("Password: ")
		passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", "", fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // New line after password input
		password = string(passwordBytes)
	}

	return username, password, nil
}

func getRegistrationInfo(username, email, password string) (string, string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Get username if not provided
	if username == "" {
		fmt.Print("Username: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", "", "", fmt.Errorf("failed to read username: %w", err)
		}
		username = strings.TrimSpace(input)
	}

	// Get email if not provided
	if email == "" {
		fmt.Print("Email: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", "", "", fmt.Errorf("failed to read email: %w", err)
		}
		email = strings.TrimSpace(input)
	}

	// Get password if not provided
	if password == "" {
		fmt.Print("Password: ")
		passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // New line after password input
		password = string(passwordBytes)
	}

	return username, email, password, nil
}

func loginToRegistry(username, password, registryURL string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Use provided registry URL or config default
	if registryURL != "" {
		cfg.SetRegistryURL(registryURL)
	}

	// Create client
	c := client.NewClient(cfg.RegistryURL)

	// Validate registry URL and connectivity
	if err := c.ValidateURL(); err != nil {
		return fmt.Errorf("invalid registry URL: %w", err)
	}

	if err := c.HealthCheck(); err != nil {
		return fmt.Errorf("cannot connect to registry at %s: %w", cfg.RegistryURL, err)
	}

	// Get credentials
	username, password, err = getCredentials(username, password)
	if err != nil {
		return err
	}

	// Login
	authResp, err := c.Login(username, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Save auth info to config
	cfg.SetAuth(authResp.Token, authResp.User.Username, authResp.User.ID.String())
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save authentication: %w", err)
	}

	fmt.Printf("Successfully logged in as %s\n", authResp.User.Username)
	return nil
}

func logoutFromRegistry() error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Clear auth
	cfg.ClearAuth()
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func registerAccount(username, email, password, registryURL string) error {
	// Load config (create default if doesn't exist)
	var cfg *config.Config
	loadedCfg, err := config.Load()
	if err != nil {
		// If no config exists, create a default one
		cfg = config.CreateDefaultConfig()
	} else {
		cfg = loadedCfg
	}

	// Use provided registry URL or config default
	if registryURL != "" {
		cfg.SetRegistryURL(registryURL)
	}

	// Create client
	c := client.NewClient(cfg.RegistryURL)

	// Validate registry URL and connectivity
	if err := c.ValidateURL(); err != nil {
		return fmt.Errorf("invalid registry URL: %w", err)
	}

	if err := c.HealthCheck(); err != nil {
		return fmt.Errorf("cannot connect to registry at %s: %w", cfg.RegistryURL, err)
	}

	// Get registration info
	username, email, password, err = getRegistrationInfo(username, email, password)
	if err != nil {
		return err
	}

	// Register
	authResp, err := c.Register(username, email, password)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	// Save auth info to config
	cfg.SetAuth(authResp.Token, authResp.User.Username, authResp.User.ID.String())
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save authentication: %w", err)
	}

	fmt.Printf("Successfully registered and logged in as %s\n", authResp.User.Username)
	return nil
}