package certfix

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/client"
	"github.com/spf13/cobra"
)

var integrationKeysCmd = &cobra.Command{
	Use:     "integration-keys",
	Aliases: []string{"ik", "integration-key"},
	Short:   "Manage integration keys for external events",
	Long:    `Manage integration keys used for secure external event ingestion.`,
}

var ikListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all integration keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth("/integration-keys", token)
		if err != nil {
			return fmt.Errorf("failed to list integration keys: %w", err)
		}

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

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(keys, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tLAST USED\tEXPIRES AT")
		fmt.Fprintln(w, "----\t----\t------\t---------\t----------")

		for _, k := range keys {
			lastUsed := "Never"
			if k["last_used_at"] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", k["last_used_at"])); err == nil {
					lastUsed = t.Format("2006-01-02 15:04")
				}
			}
			expiresAt := "Never"
			if k["expires_at"] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", k["expires_at"])); err == nil {
					expiresAt = t.Format("2006-01-02 15:04")
				}
			}
			status := "Disabled"
			if k["enabled"].(bool) {
				status = "Enabled"
			}

			fmt.Fprintf(w, "%v\t%v\t%s\t%s\t%s\n", k["key_id"], k["name"], status, lastUsed, expiresAt)
		}
		w.Flush()
		return nil
	},
}

var ikCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new integration key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		expiresIn, _ := cmd.Flags().GetInt("expires-in")

		token, err := auth.GetToken()
		if err != nil {
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		payload := map[string]interface{}{
			"name":            name,
			"expires_in_days": expiresIn,
		}

		response, err := apiClient.PostWithAuth("/integration-keys", payload, token)
		if err != nil {
			return fmt.Errorf("failed to create integration key: %w", err)
		}

		fmt.Printf("✓ Integration key created successfully\n")
		fmt.Printf("Name: %v\n", response["name"])
		fmt.Printf("Key:  %v\n", response["key"])
		fmt.Println("\nIMPORTANT: Store this key safely. It will not be shown again.")
		return nil
	},
}

var ikDeleteCmd = &cobra.Command{
	Use:   "delete <key-id>",
	Short: "Delete an integration key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyID := args[0]
		token, err := auth.GetToken()
		if err != nil {
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		_, err = apiClient.DeleteWithAuth(fmt.Sprintf("/integration-keys/%s", keyID), token)
		if err != nil {
			return fmt.Errorf("failed to delete integration key: %w", err)
		}

		fmt.Printf("✓ Integration key deleted successfully\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(integrationKeysCmd)
	integrationKeysCmd.AddCommand(ikListCmd)
	integrationKeysCmd.AddCommand(ikCreateCmd)
	integrationKeysCmd.AddCommand(ikDeleteCmd)

	ikListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	ikCreateCmd.Flags().IntP("expires-in", "e", 0, "Expiration in days (0 = never)")
}
