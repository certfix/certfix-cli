package certfix

import (
	"fmt"

	"github.com/certfix/certfix-cli/internal/api"
	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of the Certificate Authority",
	Long:  `Create a complete backup of the CA including certificates, private keys, and configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check authentication
		if !auth.IsAuthenticated() {
			cmd.SilenceUsage = true
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		log := logger.GetLogger()
		log.Info("Creating CA backup...")

		client := api.NewClient()
		response, err := client.CreateBackup()
		if err != nil {
			cmd.SilenceUsage = true
			log.Debug("Failed to create backup: ", err)
			return fmt.Errorf("failed to create backup")
		}

		// Display only the status
		if status, ok := response["status"].(string); ok {
			fmt.Printf("Backup status: %s\n", status)
		} else {
			fmt.Println("Backup completed")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
