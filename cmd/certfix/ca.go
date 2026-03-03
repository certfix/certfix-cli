package certfix

import (
	"encoding/json"
	"fmt"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/client"
	"github.com/spf13/cobra"
)

var caCmd = &cobra.Command{
	Use:   "ca",
	Short: "Inspect the Certificate Authority",
	Long:  `View information about the Certificate Authority (CA), its details, and the Certificate Revocation List (CRL).`,
}

var caInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show CA serial number and validity dates",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth("/ca/info", token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get CA info: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Serial Number:  %v\n", response["serial_number"])
		fmt.Printf("Not Before:     %v\n", response["not_before"])
		fmt.Printf("Not After:      %v\n", response["not_after"])
		fmt.Printf("Subject:        %v\n", response["subject"])

		return nil
	},
}

var caDetailsCmd = &cobra.Command{
	Use:   "details",
	Short: "Show full CA certificate content",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth("/ca/details", token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get CA details: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if cert, ok := response["certificate"].(string); ok {
			fmt.Println(cert)
		} else {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
		}

		return nil
	},
}

var caCRLInfoCmd = &cobra.Command{
	Use:   "crl-info",
	Short: "Show the SHA-256 hash of the current CRL",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth("/ca/crl/info", token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get CRL info: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("CRL Hash:   %v\n", response["hash"])
		fmt.Printf("Updated At: %v\n", response["updated_at"])

		return nil
	},
}

var caCRLContentCmd = &cobra.Command{
	Use:   "crl-content",
	Short: "Show the base64-encoded CRL content",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth("/ca/crl/content", token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get CRL content: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Hash:    %v\n", response["hash"])
		fmt.Printf("Content: %v\n", response["content"])

		return nil
	},
}

func init() {
	rootCmd.AddCommand(caCmd)
	caCmd.AddCommand(caInfoCmd)
	caCmd.AddCommand(caDetailsCmd)
	caCmd.AddCommand(caCRLInfoCmd)
	caCmd.AddCommand(caCRLContentCmd)

	caInfoCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	caDetailsCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	caCRLInfoCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	caCRLContentCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}
