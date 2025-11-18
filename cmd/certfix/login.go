package certfix

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Certfix services",
	Long: `Login to Certfix services using your credentials.
This will store an authentication token for subsequent commands.

Run without flags for interactive mode, or provide credentials via flags.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		// Check if API endpoint is configured FIRST
		endpoint := config.GetDefaultEndpoint()
		if endpoint == "" || endpoint == "https://api.certfix.io" {
			cmd.SilenceUsage = true
			fmt.Println("⚠ No API endpoint configured.")
			fmt.Println("Please run 'certfix configure' first to set up your API endpoint.")
			return fmt.Errorf("API endpoint not configured")
		}

		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		// Interactive mode if no flags provided
		if !cmd.Flags().Changed("username") && !cmd.Flags().Changed("password") {
			var err error
			username, password, err = interactiveLogin()
			if err != nil {
				return err
			}
		}

		// Validate inputs
		if username == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("username is required")
		}
		if password == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("password is required")
		}

		log.Info("Attempting to login...")

		// Perform authentication (endpoint is now always from config)
		token, err := auth.Login(username, password, "")
		if err != nil {
			cmd.SilenceUsage = true
			log.Debug("Login failed: ", err)
			return fmt.Errorf("login failed: invalid credentials or connection error")
		}

		// Store the token
		if err := auth.StoreToken(token); err != nil {
			cmd.SilenceUsage = true
			log.WithError(err).Error("Failed to store authentication token")
			return fmt.Errorf("failed to store token: %w", err)
		}

		log.Info("Successfully logged in")
		fmt.Println("✓ Successfully logged in to Certfix")
		return nil
	},
}

// interactiveLogin prompts the user for credentials
func interactiveLogin() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Prompt for username
	fmt.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)

	// Prompt for password (hidden input)
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Add newline after password input

	password := strings.TrimSpace(string(passwordBytes))

	return username, password, nil
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringP("username", "u", "", "Username for authentication")
	loginCmd.Flags().StringP("password", "p", "", "Password for authentication")
}
