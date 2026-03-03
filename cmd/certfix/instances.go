package certfix

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/certfix/certfix-cli/internal/api"
	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/client"
	"github.com/spf13/cobra"
)

var instancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage instances",
	Long:  `Manage service instances including listing, getting details, viewing logs, and deleting instances.`,
}

// instanceTableWriter writes a tabular list of instances.
func instanceTableWriter(instances []map[string]interface{}) {
	// Apply "Lost" logic: mark as Lost if last_seen_at > 5 minutes ago
	for _, instance := range instances {
		lastSeen, _ := instance["last_seen_at"].(string)
		if lastSeen != "" {
			lastSeenTime, err := time.Parse(time.RFC3339, lastSeen)
			if err == nil && time.Since(lastSeenTime) > 5*time.Minute {
				instance["status"] = "Lost"
			}
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tHOSTNAME\tOS\tIP ADDRESS\tSTATUS\tLAST SEEN\tVERSION")
	fmt.Fprintln(w, "--\t--------\t--\t----------\t------\t---------\t-------")

	for _, instance := range instances {
		s := func(k string) string {
			if v, ok := instance[k]; ok && v != nil {
				return fmt.Sprintf("%v", v)
			}
			return "N/A"
		}

		id := s("id")
		hostname := s("hostname")
		osType := s("os_type")
		arch := s("architecture")
		osInfo := fmt.Sprintf("%s / %s", osType, arch)
		ip := s("ip_address")
		status := s("status")
		lastSeen := s("last_seen_at")
		if t, err := time.Parse(time.RFC3339, lastSeen); err == nil {
			lastSeen = t.Format("2006-01-02 15:04")
		}
		version := s("agent_version")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", id, hostname, osInfo, ip, status, lastSeen, version)
	}
	w.Flush()
}

var instancesListCmd = &cobra.Command{
	Use:   "list <key-id>",
	Short: "List all instances by service key",
	Long:  `List all instances associated with a specific service key ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyID := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")

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

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(instances, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "HOSTNAME\tOS\tIP ADDRESS\tSTATUS\tREGISTERED\tLAST SEEN\tVERSION")
		fmt.Fprintln(w, "--------\t--\t----------\t------\t----------\t---------\t-------")

		for _, instance := range instances {
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

var instancesListAllCmd = &cobra.Command{
	Use:   "list-all",
	Short: "List all instances globally",
	Long:  `List all service instances across all services.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth("/instances", token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list instances: %w", err)
		}

		var instances []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if inst, ok := item.(map[string]interface{}); ok {
						instances = append(instances, inst)
					}
				}
			}
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(instances, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(instances) == 0 {
			fmt.Println("No instances found.")
			return nil
		}

		instanceTableWriter(instances)
		return nil
	},
}

var instancesListByServiceCmd = &cobra.Command{
	Use:   "list-by-service <service-hash>",
	Short: "List all instances for a service",
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

		response, err := apiClient.GetWithAuth(fmt.Sprintf("/services/%s/instances", serviceHash), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list instances: %w", err)
		}

		var instances []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if inst, ok := item.(map[string]interface{}); ok {
						instances = append(instances, inst)
					}
				}
			}
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(instances, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(instances) == 0 {
			fmt.Println("No instances found.")
			return nil
		}

		instanceTableWriter(instances)
		return nil
	},
}

var instancesGetCmd = &cobra.Command{
	Use:   "get <instance-id>",
	Short: "Get details of a specific instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceID := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth(fmt.Sprintf("/instances/%s", instanceID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get instance: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		s := func(k string) string {
			if v, ok := response[k]; ok && v != nil {
				return fmt.Sprintf("%v", v)
			}
			return "N/A"
		}

		fmt.Printf("ID:           %s\n", s("id"))
		fmt.Printf("Hostname:     %s\n", s("hostname"))
		fmt.Printf("OS:           %s / %s\n", s("os_type"), s("architecture"))
		fmt.Printf("IP Address:   %s\n", s("ip_address"))
		fmt.Printf("Status:       %s\n", s("status"))
		fmt.Printf("Agent Ver:    %s\n", s("agent_version"))
		fmt.Printf("Service Hash: %s\n", s("service_hash"))
		fmt.Printf("First Seen:   %s\n", s("first_registered_at"))
		fmt.Printf("Last Seen:    %s\n", s("last_seen_at"))

		return nil
	},
}

var instancesDeleteCmd = &cobra.Command{
	Use:     "delete <instance-id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete an instance",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Are you sure you want to delete instance %s? (y/N): ", instanceID)
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

		_, err = apiClient.DeleteWithAuth(fmt.Sprintf("/instances/%s", instanceID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to delete instance: %w", err)
		}

		fmt.Printf("✓ Instance deleted successfully\n")
		return nil
	},
}

var instancesLogsCmd = &cobra.Command{
	Use:   "logs <instance-id>",
	Short: "Get logs for a specific instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		instanceID := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")
		limit, _ := cmd.Flags().GetInt("limit")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		apiEndpoint := fmt.Sprintf("/instances/%s/logs", instanceID)
		if limit > 0 {
			apiEndpoint = fmt.Sprintf("%s?limit=%d", apiEndpoint, limit)
		}

		response, err := apiClient.GetWithAuth(apiEndpoint, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get instance logs: %w", err)
		}

		var logs []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if log, ok := item.(map[string]interface{}); ok {
						logs = append(logs, log)
					}
				}
			}
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(logs, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(logs) == 0 {
			fmt.Println("No logs found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "TIMESTAMP\tEVENT TYPE\tMESSAGE")
		fmt.Fprintln(w, "---------\t----------\t-------")

		for _, log := range logs {
			ts := fmt.Sprintf("%v", log["created_at"])
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				ts = t.Format("2006-01-02 15:04:05")
			}
			eventType := fmt.Sprintf("%v", log["event_type"])
			message := fmt.Sprintf("%v", log["message"])
			if len(message) > 80 {
				message = message[:77] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", ts, eventType, message)
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(instancesCmd)
	instancesCmd.AddCommand(instancesListCmd)
	instancesCmd.AddCommand(instancesListAllCmd)
	instancesCmd.AddCommand(instancesListByServiceCmd)
	instancesCmd.AddCommand(instancesGetCmd)
	instancesCmd.AddCommand(instancesDeleteCmd)
	instancesCmd.AddCommand(instancesLogsCmd)

	instancesListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	instancesListAllCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	instancesListByServiceCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	instancesGetCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	instancesDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	instancesLogsCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	instancesLogsCmd.Flags().IntP("limit", "l", 50, "Maximum number of log entries to show")
}
