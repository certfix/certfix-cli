package certfix

import (
	"encoding/json"
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
		clientId, _ := cmd.Flags().GetString("client-id")

		// Validate certificate type
		if certType != "server" && certType != "client" {
			cmd.SilenceUsage = true
			return fmt.Errorf("invalid certificate type: %s (must be 'server' or 'client')", certType)
		}

		// Validate client-id is required for client certificates
		if certType == "client" && clientId == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("--client-id is required for client certificates")
		}

		// Check authentication
		if !auth.IsAuthenticated() {
			cmd.SilenceUsage = true
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		log := logger.GetLogger()
		log.Infof("Creating %s certificate: %s", certType, commonName)

		client := api.NewClient()
		response, err := client.CreateCertificate(commonName, certType, description, days, keySize, san, clientId)
		if err != nil {
			cmd.SilenceUsage = true
			log.Debug("Failed to create certificate: ", err)
			return err
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
	Use:   "list [valid|revoked|expiring]",
	Short: "List certificates",
	Long: `List certificates by status:
  - valid: List all valid certificates
  - revoked: List all revoked certificates
  - expiring <days>: List certificates expiring in the specified number of days`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		listType := args[0]

		// Check authentication
		if !auth.IsAuthenticated() {
			cmd.SilenceUsage = true
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		log := logger.GetLogger()
		client := api.NewClient()
		var response []map[string]interface{}
		var err error

		switch listType {
		case "valid":
			log.Info("Listing valid certificates")
			response, err = client.ListValidCertificates()
		case "revoked":
			log.Info("Listing revoked certificates")
			response, err = client.ListRevokedCertificates()
		case "expiring":
			if len(args) < 2 {
				cmd.SilenceUsage = true
				return fmt.Errorf("missing days argument for 'expiring' command. Usage: cert list expiring <days>")
			}
			days := args[1]
			log.Infof("Listing certificates expiring in %s days", days)
			response, err = client.ListExpiringCertificates(days)
		default:
			cmd.SilenceUsage = true
			return fmt.Errorf("invalid list type: %s. Use 'valid', 'revoked', or 'expiring <days>'", listType)
		}

		if err != nil {
			cmd.SilenceUsage = true
			log.Debug("Failed to list certificates: ", err)
			return err
		}

		if len(response) == 0 {
			fmt.Println("[]")
			return nil
		}

		// Build simplified output with selected fields
		output := []map[string]interface{}{}
		for _, cert := range response {
			simplified := map[string]interface{}{
				"app_name":         cert["app_name"],
				"unique_id":        cert["unique_id"],
				"client_id":        cert["client_id"],
				"certificate_type": cert["certificate_type"],
				"expiration_date":  cert["expiration_date"],
				"status":           cert["status"],
				"revocation_date":  cert["revocation_date"],
			}
			output = append(output, simplified)
		}

		// Print as formatted JSON
		jsonOutput, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		fmt.Println(string(jsonOutput))

		return nil
	},
}

var certRevokeCmd = &cobra.Command{
	Use:   "revoke [unique-id|all]",
	Short: "Revoke a certificate or all certificates",
	Long: `Revoke a certificate by unique ID or revoke all certificates.
  - revoke <unique-id>: Revoke a specific certificate
  - revoke all: Revoke all certificates`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		cascade, _ := cmd.Flags().GetBool("cascade")
		reason, _ := cmd.Flags().GetString("reason")

		// Check authentication
		if !auth.IsAuthenticated() {
			cmd.SilenceUsage = true
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		log := logger.GetLogger()
		client := api.NewClient()
		var err error

		if target == "all" {
			log.Info("Revoking all certificates")
			err = client.RevokeAllCertificates(reason)
			if err != nil {
				cmd.SilenceUsage = true
				log.Debug("Failed to revoke all certificates: ", err)
				return err
			}
			fmt.Println("✓ All certificates revoked successfully")
		} else {
			log.Infof("Revoking certificate: %s", target)
			err = client.RevokeCertificate(target, cascade, reason)
			if err != nil {
				cmd.SilenceUsage = true
				log.Debug("Failed to revoke certificate: ", err)
				return err
			}
			fmt.Printf("✓ Certificate '%s' revoked successfully\n", target)
		}

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
	certCreateCmd.Flags().StringP("client-id", "c", "", "Client ID (required for client certificates)")
	certCreateCmd.Flags().StringP("description", "d", "", "Certificate description (optional)")
	certCreateCmd.Flags().IntP("days", "", 0, "Validity period in days (optional)")
	certCreateCmd.Flags().IntP("key-size", "k", 0, "RSA key size in bits (optional)")
	certCreateCmd.Flags().StringP("san", "s", "", "Subject Alternative Names, e.g., 'DNS:example.com,IP:192.168.1.1' (optional)")
	certCreateCmd.MarkFlagRequired("type")

	// Flags for cert revoke command
	certRevokeCmd.Flags().BoolP("cascade", "c", true, "Cascade revocation (default: true)")
	certRevokeCmd.Flags().StringP("reason", "r", "superseded", "Revocation reason (default: superseded)")
}
