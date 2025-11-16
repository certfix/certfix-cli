package certfix

import (
	"fmt"

	"github.com/certfix/certfix-cli/internal/api"
	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize certificates",
	Long:  `Synchronize certificates with the Certificate Authority.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		log.Info("Synchronizing certificates...")

		// Check authentication
		if !auth.IsAuthenticated() {
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		client := api.NewClient()
		response, err := client.SyncCertificates()
		if err != nil {
			log.WithError(err).Error("Failed to synchronize certificates")
			return fmt.Errorf("failed to synchronize certificates: %w", err)
		}

		// Display success and synced count
		if success, ok := response["success"].(bool); ok && success {
			fmt.Println("✓ Synchronization successful")
		} else {
			fmt.Println("✗ Synchronization failed")
		}

		if synced, ok := response["synced"].(float64); ok {
			fmt.Printf("Synced: %.0f certificates\n", synced)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
