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
				fmt.Printf("%-35s %v\n", label+":", v)
			}
		}

		printStat("Active Services", "activeServices")
		printStat("Active Instances", "activeInstances")
		printStat("Active Certificates", "activeCertificates")
		printStat("Total Certificates Generated", "totalCertificatesGenerated")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}
