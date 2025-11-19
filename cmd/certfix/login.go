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
	Long: `Login to Certfix services using your email and personal access token.
This will store an authentication token for subsequent commands.

Run without flags for interactive mode, or provide credentials via flags.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		// Check if API endpoint is configured FIRST
		endpoint := config.GetDefaultEndpoint()
		if endpoint == "" || endpoint == "https://certfix.io" {
			cmd.SilenceUsage = true
			fmt.Println("⚠ No API endpoint configured.")
			fmt.Println("Please run 'certfix configure' first to set up your API endpoint.")
			fmt.Println("Example: certfix configure set endpoint http://localhost:3001")
			return fmt.Errorf("API endpoint not configured")
		}

		email, _ := cmd.Flags().GetString("email")
		personalToken, _ := cmd.Flags().GetString("token")

		// Interactive mode if no flags provided
		if !cmd.Flags().Changed("email") && !cmd.Flags().Changed("token") {
			var err error
			email, personalToken, err = interactiveLogin()
			if err != nil {
				return err
			}
		}

		// Validate inputs
		if email == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("email is required")
		}
		if personalToken == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("personal access token is required")
		}

		log.Info("Attempting to login with personal access token...")

		// Perform authentication
		token, err := auth.Login(email, personalToken, "")
		if err != nil {
			cmd.SilenceUsage = true
			log.Debug("Login failed: ", err)
			return fmt.Errorf("login failed: invalid token or connection error")
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

	// Prompt for email
	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)

	// Prompt for personal access token (hidden input)
	fmt.Print("Personal Access Token: ")
	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", fmt.Errorf("failed to read token: %w", err)
	}
	fmt.Println() // Add newline after token input

	token := strings.TrimSpace(string(tokenBytes))

	return email, token, nil
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringP("email", "e", "", "Email for authentication")
	loginCmd.Flags().StringP("token", "t", "", "Personal access token for authentication")
}
