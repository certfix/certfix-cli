package certfix

import (
	"fmt"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Certfix services",
	Long: `Login to Certfix services using your credentials.
This will store an authentication token for subsequent commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		endpoint, _ := cmd.Flags().GetString("endpoint")

		log := logger.GetLogger()
		log.Info("Attempting to login...")

		// Perform authentication
		token, err := auth.Login(username, password, endpoint)
		if err != nil {
			log.WithError(err).Error("Login failed")
			return fmt.Errorf("login failed: %w", err)
		}

		// Store the token
		if err := auth.StoreToken(token); err != nil {
			log.WithError(err).Error("Failed to store authentication token")
			return fmt.Errorf("failed to store token: %w", err)
		}

		log.Info("Successfully logged in")
		fmt.Println("Successfully logged in to Certfix")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringP("username", "u", "", "Username for authentication (required)")
	loginCmd.Flags().StringP("password", "p", "", "Password for authentication (required)")
	loginCmd.Flags().StringP("endpoint", "e", "", "API endpoint URL (optional, uses default if not specified)")
	loginCmd.MarkFlagRequired("username")
	loginCmd.MarkFlagRequired("password")
}
