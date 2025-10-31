package certfix

import (
	"fmt"

	"github.com/certfix/certfix-cli/internal/api"
	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Manage certificates",
	Long:  `Create, renew, revoke, and manage SSL/TLS certificates.`,
}

var certCreateCmd = &cobra.Command{
	Use:   "create [domain]",
	Short: "Create a new certificate",
	Long:  `Request a new SSL/TLS certificate for the specified domain.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]

		log := logger.GetLogger()
		log.Infof("Creating certificate for domain: %s", domain)

		// Check authentication
		if !auth.IsAuthenticated() {
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		client := api.NewClient()
		cert, err := client.CreateCertificate(domain)
		if err != nil {
			log.WithError(err).Error("Failed to create certificate")
			return fmt.Errorf("failed to create certificate: %w", err)
		}

		fmt.Printf("Certificate created for domain '%s'\n", domain)
		fmt.Printf("Certificate ID: %s\n", cert.ID)
		fmt.Printf("Valid until: %s\n", cert.ExpiresAt)
		return nil
	},
}

var certListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all certificates",
	Long:  `List all certificates in your account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		log.Info("Listing certificates")

		// Check authentication
		if !auth.IsAuthenticated() {
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		client := api.NewClient()
		certs, err := client.ListCertificates()
		if err != nil {
			log.WithError(err).Error("Failed to list certificates")
			return fmt.Errorf("failed to list certificates: %w", err)
		}

		if len(certs) == 0 {
			fmt.Println("No certificates found")
			return nil
		}

		fmt.Println("Certificates:")
		for _, cert := range certs {
			fmt.Printf("  - %s (ID: %s, Status: %s, Expires: %s)\n",
				cert.Domain, cert.ID, cert.Status, cert.ExpiresAt)
		}
		return nil
	},
}

var certRenewCmd = &cobra.Command{
	Use:   "renew [id]",
	Short: "Renew a certificate",
	Long:  `Renew an existing certificate by ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		log := logger.GetLogger()
		log.Infof("Renewing certificate: %s", id)

		// Check authentication
		if !auth.IsAuthenticated() {
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		client := api.NewClient()
		cert, err := client.RenewCertificate(id)
		if err != nil {
			log.WithError(err).Error("Failed to renew certificate")
			return fmt.Errorf("failed to renew certificate: %w", err)
		}

		fmt.Printf("Certificate '%s' renewed successfully\n", id)
		fmt.Printf("New expiration date: %s\n", cert.ExpiresAt)
		return nil
	},
}

var certRevokeCmd = &cobra.Command{
	Use:   "revoke [id]",
	Short: "Revoke a certificate",
	Long:  `Revoke an existing certificate by ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		log := logger.GetLogger()
		log.Infof("Revoking certificate: %s", id)

		// Check authentication
		if !auth.IsAuthenticated() {
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		client := api.NewClient()
		if err := client.RevokeCertificate(id); err != nil {
			log.WithError(err).Error("Failed to revoke certificate")
			return fmt.Errorf("failed to revoke certificate: %w", err)
		}

		fmt.Printf("Certificate '%s' revoked successfully\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(certCmd)
	certCmd.AddCommand(certCreateCmd)
	certCmd.AddCommand(certListCmd)
	certCmd.AddCommand(certRenewCmd)
	certCmd.AddCommand(certRevokeCmd)
}
