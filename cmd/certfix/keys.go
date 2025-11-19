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
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:     "keys",
	Aliases: []string{"key"},
	Short:   "Manage service API keys",
	Long:    `Manage service API keys including listing, creating, enabling/disabling, and deleting keys.`,
}

var keysListCmd = &cobra.Command{
	Use:     "list <service-hash>",
	Aliases: []string{"ls"},
	Short:   "List all API keys for a service",
	Long:    `List all API keys for a specific service.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		serviceHash := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		// Create API client
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		apiEndpoint := fmt.Sprintf("/services/%s/keys/list", serviceHash)
		log.Debugf("GET %s%s", endpoint, apiEndpoint)

		// Make request
		response, err := apiClient.GetWithAuth(apiEndpoint, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list service keys: %w", err)
		}

		// Parse response
		var keys []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if key, ok := item.(map[string]interface{}); ok {
						keys = append(keys, key)
					}
				}
			}
		}

		if len(keys) == 0 {
			fmt.Println("No API keys found.")
			return nil
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(keys, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "KEY ID\tKEY NAME\tAPI KEY\tSTATUS\tEXPIRATION\tCREATED AT")
		fmt.Fprintln(w, "------\t--------\t-------\t------\t----------\t----------")

		for _, key := range keys {
			keyID := fmt.Sprintf("%v", key["key_id"])
			if len(keyID) > 12 {
				keyID = keyID[:12] + "..."
			}

			keyName := fmt.Sprintf("%v", key["key_name"])
			if len(keyName) > 20 {
				keyName = keyName[:17] + "..."
			}

			apiKey := fmt.Sprintf("%v", key["api_key"])
			if len(apiKey) > 20 {
				apiKey = apiKey[:17] + "..."
			}

			enabled := key["enabled"].(bool)
			status := "Disabled"
			if enabled {
				status = "Enabled"
			}

			expiresAt := ""
			if key["expires_at"] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", key["expires_at"])); err == nil {
					expiresAt = t.Format("2006-01-02")
				}
			}

			createdAt := ""
			if key["created_at"] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", key["created_at"])); err == nil {
					createdAt = t.Format("2006-01-02 15:04")
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", keyID, keyName, apiKey, status, expiresAt, createdAt)
		}
		w.Flush()

		return nil
	},
}

var keysGetCmd = &cobra.Command{
	Use:   "get <service-hash>",
	Short: "Get API keys data for a service",
	Long:  `Get complete API keys data for a service including service info and all keys.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceHash := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		// Create API client
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		// Make request
		response, err := apiClient.GetWithAuth(fmt.Sprintf("/services/%s/keys", serviceHash), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get service keys data: %w", err)
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Pretty print service info
		if service, ok := response["service"].(map[string]interface{}); ok {
			fmt.Println("Service Information:")
			fmt.Printf("  Hash:   %v\n", service["service_hash"])
			fmt.Printf("  Name:   %v\n", service["service_name"])
			fmt.Printf("  Active: %v\n\n", service["active"])
		}

		// Print keys
		if keys, ok := response["keys"].([]interface{}); ok && len(keys) > 0 {
			fmt.Println("API Keys:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "  KEY ID\tKEY NAME\tSTATUS\tEXPIRES AT")
			fmt.Fprintln(w, "  ------\t--------\t------\t----------")

			for _, item := range keys {
				if key, ok := item.(map[string]interface{}); ok {
					keyID := fmt.Sprintf("%v", key["key_id"])
					if len(keyID) > 12 {
						keyID = keyID[:12] + "..."
					}

					keyName := fmt.Sprintf("%v", key["key_name"])

					enabled := key["enabled"].(bool)
					status := "Disabled"
					if enabled {
						status = "Enabled"
					}

					expiresAt := ""
					if key["expires_at"] != nil {
						if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", key["expires_at"])); err == nil {
							expiresAt = t.Format("2006-01-02")
						}
					}

					fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", keyID, keyName, status, expiresAt)
				}
			}
			w.Flush()
		} else {
			fmt.Println("No API keys found.")
		}

		return nil
	},
}

var keysAddCmd = &cobra.Command{
	Use:   "add <service-hash>",
	Short: "Add a new API key to a service",
	Long:  `Add a new API key to a service with a name and expiration period.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		serviceHash := args[0]

		// Get flags
		keyName, _ := cmd.Flags().GetString("name")
		expirationDays, _ := cmd.Flags().GetInt("expiration")

		// Validate required fields
		if keyName == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("key name is required (use --name)")
		}

		if expirationDays <= 0 {
			cmd.SilenceUsage = true
			return fmt.Errorf("expiration days must be greater than 0 (use --expiration)")
		}

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		// Create API client
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		// Prepare payload
		payload := map[string]interface{}{
			"key_name":        keyName,
			"expiration_days": expirationDays,
		}

		log.Infof("Adding API key: %s (expires in %d days)", keyName, expirationDays)

		// Make request
		response, err := apiClient.PostWithAuth(fmt.Sprintf("/services/%s/keys", serviceHash), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to add API key: %w", err)
		}

		fmt.Printf("✓ API key added successfully\n")
		fmt.Printf("Key ID:     %v\n", response["key_id"])
		fmt.Printf("Key Name:   %v\n", response["key_name"])
		fmt.Printf("API Key:    %v\n", response["api_key"])
		fmt.Printf("Expires At: %v\n", response["expires_at"])
		enabledStatus := "Disabled"
		if enabled, ok := response["enabled"].(bool); ok && enabled {
			enabledStatus = "Enabled"
		}
		fmt.Printf("Status:     %s\n", enabledStatus)
		fmt.Printf("\n⚠️  Important: Save the API key now. It won't be shown again in full.\n")

		return nil
	},
}

var keysToggleCmd = &cobra.Command{
	Use:   "toggle <service-hash> <key-id>",
	Short: "Toggle an API key (enable/disable)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		serviceHash := args[0]
		keyID := args[1]

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		// Create API client
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		log.Infof("Toggling API key: %s", keyID)

		// Make PUT request
		response, err := apiClient.PutWithAuth(fmt.Sprintf("/services/%s/keys/%s/toggle", serviceHash, keyID), nil, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to toggle API key: %w", err)
		}

		fmt.Printf("✓ API key toggled successfully\n")
		fmt.Printf("Key ID:    %v\n", response["key_id"])
		fmt.Printf("Key Name:  %v\n", response["key_name"])
		enabledStatus := "Disabled"
		if enabled, ok := response["enabled"].(bool); ok && enabled {
			enabledStatus = "Enabled"
		}
		fmt.Printf("Status:    %s\n", enabledStatus)

		return nil
	},
}

var keysEnableCmd = &cobra.Command{
	Use:   "enable <service-hash> <key-id>",
	Short: "Enable an API key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceHash := args[0]
		keyID := args[1]

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		// Create API client
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		// Make request (toggle endpoint toggles the current state)
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/services/%s/keys/%s/toggle", serviceHash, keyID), nil, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to toggle API key: %w", err)
		}

		fmt.Printf("✓ API key toggled\n")
		fmt.Printf("Note: The toggle endpoint switches the current state. Use 'get' or 'list' to verify the new status.\n")
		return nil
	},
}

var keysDisableCmd = &cobra.Command{
	Use:   "disable <service-hash> <key-id>",
	Short: "Disable an API key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceHash := args[0]
		keyID := args[1]

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		// Create API client
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		// Make request (toggle endpoint toggles the current state)
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/services/%s/keys/%s/toggle", serviceHash, keyID), nil, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to toggle API key: %w", err)
		}

		fmt.Printf("✓ API key toggled\n")
		fmt.Printf("Note: The toggle endpoint switches the current state. Use 'get' or 'list' to verify the new status.\n")
		return nil
	},
}

var keysDeleteCmd = &cobra.Command{
	Use:     "delete <service-hash> <key-id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete an API key",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		serviceHash := args[0]
		keyID := args[1]

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete API key %s? (y/N): ", keyID)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Deletion cancelled.")
				return nil
			}
		}

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		// Create API client
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		log.Infof("Deleting API key: %s", keyID)

		// Make request
		_, err = apiClient.DeleteWithAuth(fmt.Sprintf("/services/%s/keys/%s", serviceHash, keyID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to delete API key: %w", err)
		}

		fmt.Printf("✓ API key deleted successfully\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)

	// Add subcommands
	keysCmd.AddCommand(keysListCmd)
	keysCmd.AddCommand(keysGetCmd)
	keysCmd.AddCommand(keysAddCmd)
	keysCmd.AddCommand(keysToggleCmd)
	keysCmd.AddCommand(keysEnableCmd)
	keysCmd.AddCommand(keysDisableCmd)
	keysCmd.AddCommand(keysDeleteCmd)

	// List command flags
	keysListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Get command flags
	keysGetCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Add command flags
	keysAddCmd.Flags().StringP("name", "n", "", "Name of the API key (required)")
	keysAddCmd.Flags().IntP("expiration", "e", 365, "Expiration period in days (required)")
	keysAddCmd.MarkFlagRequired("name")
	keysAddCmd.MarkFlagRequired("expiration")

	// Delete command flags
	keysDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
}
