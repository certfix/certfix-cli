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
	Long:  `Create, revoke, and manage SSL/TLS certificates.`,
}

var certCreateCmd = &cobra.Command{
	Use:   "create [common-name]",
	Short: "Create a new certificate",
	Long:  `Request a new SSL/TLS certificate (server or client) with the specified common name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		commonName := args[0]
		certType, _ := cmd.Flags().GetString("type")
		description, _ := cmd.Flags().GetString("description")
		days, _ := cmd.Flags().GetInt("days")
		keySize, _ := cmd.Flags().GetInt("key-size")
		san, _ := cmd.Flags().GetString("san")

		log := logger.GetLogger()
		log.Infof("Creating %s certificate: %s", certType, commonName)

		// Validate certificate type
		if certType != "server" && certType != "client" {
			return fmt.Errorf("invalid certificate type: %s (must be 'server' or 'client')", certType)
		}

		// Check authentication
		if !auth.IsAuthenticated() {
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		client := api.NewClient()
		response, err := client.CreateCertificate(commonName, certType, description, days, keySize, san)
		if err != nil {
			log.WithError(err).Error("Failed to create certificate")
			return fmt.Errorf("failed to create certificate: %w", err)
		}

		// Display certificate information
		fmt.Println("✓ Certificate created successfully")
		
		// Extract certificate data based on type
		var certData map[string]interface{}
		if certType == "server" {
			if serverCert, ok := response["server_certificate"].(map[string]interface{}); ok {
				certData = serverCert
			}
		} else {
			if clientCert, ok := response["client_certificate"].(map[string]interface{}); ok {
				certData = clientCert
			}
		}

		if certData != nil {
			if uniqueID, ok := certData["unique_id"].(string); ok {
				fmt.Printf("Unique ID:     %s\n", uniqueID)
			}
			if serialNumber, ok := certData["serial_number"].(string); ok {
				fmt.Printf("Serial Number: %s\n", serialNumber)
			}
			if appName, ok := certData["app_name"].(string); ok {
				fmt.Printf("App Name:      %s\n", appName)
			}
			// Show client_id only for client certificates
			if certType == "client" {
				if clientID, ok := certData["client_id"].(string); ok {
					fmt.Printf("Client ID:     %s\n", clientID)
				}
			}
		}

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
	certCmd.AddCommand(certRevokeCmd)

	// Flags for cert create command
	certCreateCmd.Flags().StringP("type", "t", "server", "Certificate type: 'server' or 'client' (required)")
	certCreateCmd.Flags().StringP("description", "d", "", "Certificate description (optional)")
	certCreateCmd.Flags().IntP("days", "", 0, "Validity period in days (optional)")
	certCreateCmd.Flags().IntP("key-size", "k", 0, "RSA key size in bits (optional)")
	certCreateCmd.Flags().StringP("san", "s", "", "Subject Alternative Names, e.g., 'DNS:example.com,IP:192.168.1.1' (optional)")
	certCreateCmd.MarkFlagRequired("type")
}
