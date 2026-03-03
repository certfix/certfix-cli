package certfix

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/client"
	"github.com/spf13/cobra"
)

var certsCmd = &cobra.Command{
	Use:     "certs",
	Aliases: []string{"cert", "certificates", "certificate"},
	Short:   "Manage service certificates",
	Long:    `Manage certificates for services including listing, getting details, and revoking certificates.`,
}

var certsListCmd = &cobra.Command{
	Use:   "list <service-hash>",
	Short: "List all certificates for a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceHash := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth(fmt.Sprintf("/services/%s/certificates", serviceHash), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list certificates: %w", err)
		}

		var certs []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if cert, ok := item.(map[string]interface{}); ok {
						certs = append(certs, cert)
					}
				}
			}
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(certs, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(certs) == 0 {
			fmt.Println("No certificates found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "UNIQUE ID\tTYPE\tSTATUS\tSERIAL\tCOMMON NAME\tEXPIRES AT")
		fmt.Fprintln(w, "---------\t----\t------\t------\t-----------\t----------")

		for _, cert := range certs {
			uniqueID := fmt.Sprintf("%v", cert["unique_id"])
			certType := fmt.Sprintf("%v", cert["certificate_type"])
			status := fmt.Sprintf("%v", cert["status"])
			serial := fmt.Sprintf("%v", cert["serial_number"])
			cn := fmt.Sprintf("%v", cert["common_name"])
			if len(cn) > 30 {
				cn = cn[:27] + "..."
			}
			expiresAt := ""
			if cert["expires_at"] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", cert["expires_at"])); err == nil {
					expiresAt = t.Format("2006-01-02 15:04")
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", uniqueID, certType, status, serial, cn, expiresAt)
		}
		w.Flush()

		return nil
	},
}

var certsGetCmd = &cobra.Command{
	Use:   "get <unique-id>",
	Short: "Get details of a specific certificate",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		uniqueID := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth(fmt.Sprintf("/services/certificates/%s/details", uniqueID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get certificate: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Unique ID:    %v\n", response["unique_id"])
		fmt.Printf("Common Name:  %v\n", response["common_name"])
		fmt.Printf("Type:         %v\n", response["certificate_type"])
		fmt.Printf("Status:       %v\n", response["status"])
		fmt.Printf("Serial:       %v\n", response["serial_number"])
		if response["expires_at"] != nil {
			fmt.Printf("Expires At:   %v\n", response["expires_at"])
		}
		if response["revoked_at"] != nil {
			fmt.Printf("Revoked At:   %v\n", response["revoked_at"])
		}
		if response["revocation_reason"] != nil {
			fmt.Printf("Revoke Reason:%v\n", response["revocation_reason"])
		}
		if response["san"] != nil {
			fmt.Printf("SAN:          %v\n", response["san"])
		}
		if response["created_at"] != nil {
			fmt.Printf("Created At:   %v\n", response["created_at"])
		}

		return nil
	},
}

var certsRevokeCmd = &cobra.Command{
	Use:   "revoke <unique-id>",
	Short: "Revoke a certificate",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		uniqueID := args[0]
		reason, _ := cmd.Flags().GetString("reason")
		force, _ := cmd.Flags().GetBool("force")
		outputFormat, _ := cmd.Flags().GetString("output")

		if !force {
			fmt.Printf("Are you sure you want to revoke certificate %s? (y/N): ", uniqueID)
			var ans string
			fmt.Scanln(&ans)
			if strings.ToLower(ans) != "y" && strings.ToLower(ans) != "yes" {
				fmt.Println("Revocation cancelled.")
				return nil
			}
		}

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		payload := map[string]interface{}{}
		if reason != "" {
			payload["reason"] = reason
		}

		response, err := apiClient.PostWithAuth(fmt.Sprintf("/services/certificates/%s/revoke", uniqueID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to revoke certificate: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Certificate revoked successfully\n")
		fmt.Printf("Unique ID:    %s\n", uniqueID)
		if reason != "" {
			fmt.Printf("Reason:       %s\n", reason)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(certsCmd)
	certsCmd.AddCommand(certsListCmd)
	certsCmd.AddCommand(certsGetCmd)
	certsCmd.AddCommand(certsRevokeCmd)

	certsListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	certsGetCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	certsRevokeCmd.Flags().StringP("reason", "r", "", "Revocation reason (e.g. cessationOfOperation, superseded, keyCompromise)")
	certsRevokeCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	certsRevokeCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}
