package certfix

import (
	"fmt"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from Certfix services",
	Long:  `Logout from Certfix services and remove stored authentication token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		log.Info("Logging out...")

		if err := auth.Logout(); err != nil {
			log.WithError(err).Error("Logout failed")
			return fmt.Errorf("logout failed: %w", err)
		}

		log.Info("Successfully logged out")
		fmt.Println("Successfully logged out from Certfix")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
