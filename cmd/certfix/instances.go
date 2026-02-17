package certfix

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/certfix/certfix-cli/internal/api"
	"github.com/spf13/cobra"
)

var instancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage instances",
	Long:  `Manage instances including listing instances by service key.`,
}

var instancesListCmd = &cobra.Command{
	Use:   "list <key-id>",
	Short: "List all instances by service key",
	Long:  `List all instances associated with a specific service key ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyID := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")

		// Create API client
		// endpoint := config.GetAPIEndpoint()
		// Using internal/api client wrapper
		apiClient := api.NewClient()

		instances, err := apiClient.ListInstancesByKey(keyID)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list instances: %w", err)
		}

		// Apply "Lost" logic to all instances before output
		for _, instance := range instances {
			lastSeen, _ := instance["last_seen_at"].(string)
			if lastSeen != "" {
				lastSeenTime, err := time.Parse(time.RFC3339, lastSeen)
				if err == nil && time.Since(lastSeenTime) > 5*time.Minute {
					instance["status"] = "Lost"
				}
			}
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(instances, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "HOSTNAME\tOS\tIP ADDRESS\tSTATUS\tREGISTERED\tLAST SEEN\tVERSION")
		fmt.Fprintln(w, "--------\t--\t----------\t------\t----------\t---------\t-------")

		for _, instance := range instances {
			// Helper to safe string
			s := func(k string) string {
				if v, ok := instance[k]; ok && v != nil {
					return fmt.Sprintf("%v", v)
				}
				return "N/A"
			}

			hostname := s("hostname")
			osType := s("os_type")
			arch := s("architecture")
			osInfo := fmt.Sprintf("%s / %s", osType, arch)

			ip := s("ip_address")
			status := s("status")

			registered := s("first_registered_at")
			if t, err := time.Parse(time.RFC3339, registered); err == nil {
				registered = t.Format("2006-01-02 15:04")
			}

			lastSeen := s("last_seen_at")
			if t, err := time.Parse(time.RFC3339, lastSeen); err == nil {
				lastSeen = t.Format("2006-01-02 15:04")
			}

			version := s("agent_version")

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", hostname, osInfo, ip, status, registered, lastSeen, version)
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(instancesCmd)
	instancesCmd.AddCommand(instancesListCmd)

	instancesListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}
