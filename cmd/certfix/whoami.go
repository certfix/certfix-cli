package certfix

import (
	"encoding/json"
	"fmt"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/client"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the currently authenticated user",
	Long:  `Display information about the currently authenticated user based on the stored session token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth("/me", token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get current user: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("User ID:    %v\n", response["user_id"])
		fmt.Printf("Email:      %v\n", response["email"])

		isSuper, _ := response["is_super_user"].(bool)
		superStr := "No"
		if isSuper {
			superStr = "Yes"
		}
		fmt.Printf("Super User: %s\n", superStr)

		enabled, _ := response["enabled"].(bool)
		enabledStr := "No"
		if enabled {
			enabledStr = "Yes"
		}
		fmt.Printf("Enabled:    %s\n", enabledStr)


		return nil
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
	whoamiCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}
