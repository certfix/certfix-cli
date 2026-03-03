package certfix

import (
	"encoding/json"
	"fmt"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/client"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show dashboard statistics",
	Long:  `Display an overview of the system: services, instances, certificates, policies, and more.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth("/dashboard/stats", token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get dashboard stats: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Println("=== CertFix Dashboard ===")
		fmt.Println()

		printStat := func(label string, key string) {
			if v, ok := response[key]; ok && v != nil {
				fmt.Printf("%-30s %v\n", label+":", v)
			}
		}

		printStat("Total Services", "total_services")
		printStat("Active Services", "active_services")
		printStat("Total Instances", "total_instances")
		printStat("Online Instances", "online_instances")
		printStat("Total Certificates", "total_certificates")
		printStat("Valid Certificates", "valid_certificates")
		printStat("Revoked Certificates", "revoked_certificates")
		printStat("Expiring Soon (7d)", "expiring_soon")
		printStat("Total Policies", "total_policies")
		printStat("Active Policies", "active_policies")
		printStat("Total Service Groups", "total_service_groups")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}
