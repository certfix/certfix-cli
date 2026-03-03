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

var personalTokensCmd = &cobra.Command{
	Use:     "personal-tokens",
	Aliases: []string{"pat", "tokens", "token"},
	Short:   "Manage personal access tokens",
	Long:    `Manage personal access tokens (PATs) for API authentication.`,
}

var patListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all personal tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth("/personal-tokens", token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list personal tokens: %w", err)
		}

		var tokens []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if t, ok := item.(map[string]interface{}); ok {
						tokens = append(tokens, t)
					}
				}
			}
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(tokens, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(tokens) == 0 {
			fmt.Println("No personal tokens found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tLAST USED\tEXPIRES AT\tCREATED AT")
		fmt.Fprintln(w, "--\t----\t------\t---------\t----------\t----------")

		for _, t := range tokens {
			id := fmt.Sprintf("%v", t["token_id"])
			name := fmt.Sprintf("%v", t["name"])
			if len(name) > 25 {
				name = name[:22] + "..."
			}
			status := "Active"
			if revoked, ok := t["revoked"].(bool); ok && revoked {
				status = "Revoked"
			}
			lastUsed := "Never"
			if t["last_used_at"] != nil && fmt.Sprintf("%v", t["last_used_at"]) != "<nil>" {
				if lu, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", t["last_used_at"])); err == nil {
					lastUsed = lu.Format("2006-01-02 15:04")
				}
			}
			expiresAt := "Never"
			if t["expires_at"] != nil && fmt.Sprintf("%v", t["expires_at"]) != "<nil>" {
				if exp, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", t["expires_at"])); err == nil {
					expiresAt = exp.Format("2006-01-02")
				}
			}
			createdAt := ""
			if t["created_at"] != nil {
				if ca, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", t["created_at"])); err == nil {
					createdAt = ca.Format("2006-01-02 15:04")
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", id, name, status, lastUsed, expiresAt, createdAt)
		}
		w.Flush()

		return nil
	},
}

var patCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new personal token",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		expiresIn, _ := cmd.Flags().GetInt("expires-in")
		outputFormat, _ := cmd.Flags().GetString("output")

		if name == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("name is required (use --name)")
		}

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		payload := map[string]interface{}{
			"name": name,
		}
		if expiresIn > 0 {
			payload["expires_in_days"] = expiresIn
		}

		response, err := apiClient.PostWithAuth("/personal-tokens", payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to create personal token: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Personal token created successfully\n")
		fmt.Printf("ID:    %v\n", response["token_id"])
		fmt.Printf("Name:  %v\n", response["name"])
		fmt.Printf("Token: %v\n", response["token"])
		fmt.Println("\nIMPORTANT: Store this token safely. It will not be shown again.")

		return nil
	},
}

var patRevokeCmd = &cobra.Command{
	Use:   "revoke <token-id>",
	Short: "Revoke a personal token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenID := args[0]
		force, _ := cmd.Flags().GetBool("force")
		outputFormat, _ := cmd.Flags().GetString("output")

		if !force {
			fmt.Printf("Are you sure you want to revoke token %s? (y/N): ", tokenID)
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

		response, err := apiClient.PatchWithAuth(fmt.Sprintf("/personal-tokens/%s/revoke", tokenID), nil, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to revoke personal token: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Personal token revoked successfully\n")
		return nil
	},
}

var patDeleteCmd = &cobra.Command{
	Use:     "delete <token-id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a personal token",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Are you sure you want to delete token %s? (y/N): ", tokenID)
			var ans string
			fmt.Scanln(&ans)
			if strings.ToLower(ans) != "y" && strings.ToLower(ans) != "yes" {
				fmt.Println("Deletion cancelled.")
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

		_, err = apiClient.DeleteWithAuth(fmt.Sprintf("/personal-tokens/%s", tokenID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to delete personal token: %w", err)
		}

		fmt.Printf("✓ Personal token deleted successfully\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(personalTokensCmd)
	personalTokensCmd.AddCommand(patListCmd)
	personalTokensCmd.AddCommand(patCreateCmd)
	personalTokensCmd.AddCommand(patRevokeCmd)
	personalTokensCmd.AddCommand(patDeleteCmd)

	patListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	patCreateCmd.Flags().StringP("name", "n", "", "Name for the personal token (required)")
	patCreateCmd.Flags().IntP("expires-in", "e", 0, "Expiration in days (0 = never expires)")
	patCreateCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	patCreateCmd.MarkFlagRequired("name")

	patRevokeCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	patRevokeCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	patDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}
