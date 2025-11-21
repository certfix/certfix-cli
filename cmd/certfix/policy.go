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

// Strategy mapping: display labels to enum values
var strategyEnumMapping = map[string]string{
	"Eventos":             "eventos",
	"Gradual":             "gradual",
	"Janela de Manutenção": "janela_manutencao",
}

var policyCmd = &cobra.Command{
	Use:     "policy",
	Aliases: []string{"policies", "politica", "politicas"},
	Short:   "Manage policies",
	Long:    `Manage policies including listing, creating, updating, enabling/disabling, and deleting policies.`,
}

var policyListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all policies",
	Long:    `List all policies with optional filtering by strategy or enabled status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		// Get flags
		strategy, _ := cmd.Flags().GetString("strategy")
		enabledOnly, _ := cmd.Flags().GetBool("enabled")
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

		// Determine endpoint
		var apiEndpoint string
		if enabledOnly {
			apiEndpoint = "/politicas/enabled"
		} else if strategy != "" {
			apiEndpoint = fmt.Sprintf("/politicas/strategy/%s", strategy)
		} else {
			apiEndpoint = "/politicas"
		}

		log.Debugf("GET %s%s", endpoint, apiEndpoint)

		// Make request
		response, err := apiClient.GetWithAuth(apiEndpoint, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list policies: %w", err)
		}

		// Parse response
		var policies []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if policy, ok := item.(map[string]interface{}); ok {
						policies = append(policies, policy)
					}
				}
			}
		}

		if len(policies) == 0 {
			fmt.Println("No policies found.")
			return nil
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(policies, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTRATEGY\tSTATUS\tCREATED AT")
		fmt.Fprintln(w, "----\t----\t--------\t------\t----------")

		for _, policy := range policies {
			id := fmt.Sprintf("%v", policy["politica_id"])
			name := fmt.Sprintf("%v", policy["name"])
			strategy := fmt.Sprintf("%v", policy["strategy"])
			enabled := policy["enabled"].(bool)
			status := "Inactive"
			if enabled {
				status = "Active"
			}
			createdAt := ""
			if policy["created_at"] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", policy["created_at"])); err == nil {
					createdAt = t.Format("2006-01-02 15:04")
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, name, strategy, status, createdAt)
		}
		w.Flush()

		return nil
	},
}

var policyGetCmd = &cobra.Command{
	Use:   "get <policy-id>",
	Short: "Get details of a specific policy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		policyID := args[0]
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
		response, err := apiClient.GetWithAuth(fmt.Sprintf("/politicas/%s", policyID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get policy: %w", err)
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Pretty print
		fmt.Printf("ID:          %v\n", response["politica_id"])
		fmt.Printf("Name:        %v\n", response["name"])
		fmt.Printf("Strategy:    %v\n", response["strategy"])
		enabled := response["enabled"].(bool)
		status := "Inactive"
		if enabled {
			status = "Active"
		}
		fmt.Printf("Status:      %s\n", status)
		
		if response["cron_config"] != nil {
			fmt.Println("Cron Config:")
			cronConfig := response["cron_config"].(map[string]interface{})
			fmt.Printf("  Minute:    %v\n", cronConfig["minute"])
			fmt.Printf("  Hour:      %v\n", cronConfig["hour"])
			fmt.Printf("  Day:       %v\n", cronConfig["day"])
			fmt.Printf("  Month:     %v\n", cronConfig["month"])
			fmt.Printf("  Weekday:   %v\n", cronConfig["weekday"])
		}
		
		if response["event_config"] != nil {
			fmt.Println("Event Config:")
			eventConfig := response["event_config"].(map[string]interface{})
			fmt.Printf("  Event ID:  %v\n", eventConfig["evento_id"])
			fmt.Printf("  Total:     %v\n", eventConfig["total_eventos"])
		}
		
		if response["created_at"] != nil {
			fmt.Printf("Created At:  %v\n", response["created_at"])
		}
		if response["updated_at"] != nil {
			fmt.Printf("Updated At:  %v\n", response["updated_at"])
		}

		return nil
	},
}

var policyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		strategy, _ := cmd.Flags().GetString("strategy")
		enabled, _ := cmd.Flags().GetBool("enabled")

		// Cron flags
		cronMinute, _ := cmd.Flags().GetString("cron-minute")
		cronHour, _ := cmd.Flags().GetString("cron-hour")
		cronDay, _ := cmd.Flags().GetString("cron-day")
		cronMonth, _ := cmd.Flags().GetString("cron-month")
		cronWeekday, _ := cmd.Flags().GetString("cron-weekday")

		// Event flags
		eventID, _ := cmd.Flags().GetString("event-id")
		eventTotal, _ := cmd.Flags().GetInt("event-total")

		// Validate required fields
		if name == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("name is required")
		}
		if strategy == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("strategy is required")
		}

		// Validate strategy
		validStrategies := []string{"Gradual", "Janela de Manutenção", "Eventos"}
		strategyValid := false
		for _, v := range validStrategies {
			if strategy == v {
				strategyValid = true
				break
			}
		}
		if !strategyValid {
			cmd.SilenceUsage = true
			return fmt.Errorf("invalid strategy: %s (must be one of: Gradual, Janela de Manutenção, Eventos)", strategy)
		}

		// Map to enum value
		var enumStrategy string
		if enumStrat, exists := strategyEnumMapping[strategy]; exists {
			enumStrategy = enumStrat
		} else {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to map strategy to enum value")
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
			"name":     name,
			"strategy": enumStrategy,
			"enabled":  enabled,
		}

		// Add cron config if provided (for Gradual or Janela de Manutenção)
		if cronMinute != "" || cronHour != "" || cronDay != "" || cronMonth != "" || cronWeekday != "" {
			payload["cron_config"] = map[string]interface{}{
				"minute":  cronMinute,
				"hour":    cronHour,
				"day":     cronDay,
				"month":   cronMonth,
				"weekday": cronWeekday,
			}
		}

		// Add event config if provided (for Eventos strategy)
		if eventID != "" {
			payload["event_config"] = map[string]interface{}{
				"evento_id":     eventID,
				"total_eventos": eventTotal,
			}
		}

		log.Infof("Creating policy: %s", name)

		// Make request
		response, err := apiClient.PostWithAuth("/politicas", payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to create policy: %w", err)
		}

		fmt.Printf("✓ Policy created successfully\n")
		fmt.Printf("ID:       %v\n", response["politica_id"])
		fmt.Printf("Name:     %v\n", response["name"])
		fmt.Printf("Strategy: %v\n", response["strategy"])
		enabledStatus := "Inactive"
		if response["enabled"].(bool) {
			enabledStatus = "Active"
		}
		fmt.Printf("Status:   %s\n", enabledStatus)

		return nil
	},
}

var policyUpdateCmd = &cobra.Command{
	Use:   "update <policy-id>",
	Short: "Update an existing policy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		policyID := args[0]

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		strategy, _ := cmd.Flags().GetString("strategy")
		enabled := cmd.Flags().Changed("enabled")
		enabledValue, _ := cmd.Flags().GetBool("enabled")

		// Cron flags
		cronMinute, _ := cmd.Flags().GetString("cron-minute")
		cronHour, _ := cmd.Flags().GetString("cron-hour")
		cronDay, _ := cmd.Flags().GetString("cron-day")
		cronMonth, _ := cmd.Flags().GetString("cron-month")
		cronWeekday, _ := cmd.Flags().GetString("cron-weekday")

		// Event flags
		eventID, _ := cmd.Flags().GetString("event-id")
		eventTotal := cmd.Flags().Changed("event-total")
		eventTotalValue, _ := cmd.Flags().GetInt("event-total")

		// Build update payload
		payload := make(map[string]interface{})

		if name != "" {
			payload["name"] = name
		}

		if strategy != "" {
			// Validate strategy
			validStrategies := []string{"Gradual", "Janela de Manutenção", "Eventos"}
			strategyValid := false
			for _, v := range validStrategies {
				if strategy == v {
					strategyValid = true
					break
				}
			}
			if !strategyValid {
				cmd.SilenceUsage = true
				return fmt.Errorf("invalid strategy: %s (must be one of: Gradual, Janela de Manutenção, Eventos)", strategy)
			}
			// Map to enum value
			if enumStrat, exists := strategyEnumMapping[strategy]; exists {
				payload["strategy"] = enumStrat
			} else {
				cmd.SilenceUsage = true
				return fmt.Errorf("failed to map strategy to enum value")
			}
		}

		if enabled {
			payload["enabled"] = enabledValue
		}

		// Add cron config if any cron flag is provided
		if cronMinute != "" || cronHour != "" || cronDay != "" || cronMonth != "" || cronWeekday != "" {
			payload["cron_config"] = map[string]interface{}{
				"minute":  cronMinute,
				"hour":    cronHour,
				"day":     cronDay,
				"month":   cronMonth,
				"weekday": cronWeekday,
			}
		}

		// Add event config if provided
		if eventID != "" || eventTotal {
			eventConfig := make(map[string]interface{})
			if eventID != "" {
				eventConfig["evento_id"] = eventID
			}
			if eventTotal {
				eventConfig["total_eventos"] = eventTotalValue
			}
			payload["event_config"] = eventConfig
		}

		if len(payload) == 0 {
			cmd.SilenceUsage = true
			return fmt.Errorf("no fields to update")
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

		log.Infof("Updating policy: %s", policyID)

		// Make PUT request
		response, err := apiClient.PutWithAuth(fmt.Sprintf("/politicas/%s", policyID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to update policy: %w", err)
		}

		fmt.Printf("✓ Policy updated successfully\n")
		fmt.Printf("ID:       %v\n", response["politica_id"])
		fmt.Printf("Name:     %v\n", response["name"])
		fmt.Printf("Strategy: %v\n", response["strategy"])
		enabledStatus := "Inactive"
		if response["enabled"].(bool) {
			enabledStatus = "Active"
		}
		fmt.Printf("Status:   %s\n", enabledStatus)

		return nil
	},
}

var policyEnableCmd = &cobra.Command{
	Use:   "enable <policy-id>",
	Short: "Enable a policy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		policyID := args[0]

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
			"enabled": true,
		}

		// Make request
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/politicas/%s", policyID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to enable policy: %w", err)
		}

		fmt.Printf("✓ Policy enabled successfully\n")
		return nil
	},
}

var policyDisableCmd = &cobra.Command{
	Use:   "disable <policy-id>",
	Short: "Disable a policy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		policyID := args[0]

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
			"enabled": false,
		}

		// Make request
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/politicas/%s", policyID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to disable policy: %w", err)
		}

		fmt.Printf("✓ Policy disabled successfully\n")
		return nil
	},
}

var policyDeleteCmd = &cobra.Command{
	Use:     "delete <policy-id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a policy",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		policyID := args[0]

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete policy %s? (y/N): ", policyID)
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

		log.Infof("Deleting policy: %s", policyID)

		// Make request
		_, err = apiClient.DeleteWithAuth(fmt.Sprintf("/politicas/%s", policyID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to delete policy: %w", err)
		}

		fmt.Printf("✓ Policy deleted successfully\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(policyCmd)

	// Add subcommands
	policyCmd.AddCommand(policyListCmd)
	policyCmd.AddCommand(policyGetCmd)
	policyCmd.AddCommand(policyCreateCmd)
	policyCmd.AddCommand(policyUpdateCmd)
	policyCmd.AddCommand(policyEnableCmd)
	policyCmd.AddCommand(policyDisableCmd)
	policyCmd.AddCommand(policyDeleteCmd)

	// List command flags
	policyListCmd.Flags().StringP("strategy", "s", "", "Filter by strategy (Gradual, Janela de Manutenção, Eventos)")
	policyListCmd.Flags().BoolP("enabled", "e", false, "Show only enabled policies")
	policyListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Get command flags
	policyGetCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Create command flags
	policyCreateCmd.Flags().StringP("name", "n", "", "Name of the policy (required)")
	policyCreateCmd.Flags().StringP("strategy", "s", "", "Strategy: Gradual, Janela de Manutenção, or Eventos (required)")
	policyCreateCmd.Flags().BoolP("enabled", "e", true, "Enable the policy immediately (default: true)")
	
	// Cron configuration flags (for Gradual and Janela de Manutenção)
	policyCreateCmd.Flags().String("cron-minute", "*", "Cron minute (0-59 or *)")
	policyCreateCmd.Flags().String("cron-hour", "*", "Cron hour (0-23 or *)")
	policyCreateCmd.Flags().String("cron-day", "*", "Cron day (1-31 or *)")
	policyCreateCmd.Flags().String("cron-month", "*", "Cron month (1-12 or *)")
	policyCreateCmd.Flags().String("cron-weekday", "*", "Cron weekday (0-7 or *)")
	
	// Event configuration flags (for Eventos strategy)
	policyCreateCmd.Flags().String("event-id", "", "Event ID for Eventos strategy")
	policyCreateCmd.Flags().Int("event-total", 1, "Total events for Eventos strategy")
	
	policyCreateCmd.MarkFlagRequired("name")
	policyCreateCmd.MarkFlagRequired("strategy")

	// Update command flags
	policyUpdateCmd.Flags().StringP("name", "n", "", "New name for the policy")
	policyUpdateCmd.Flags().StringP("strategy", "s", "", "New strategy: Gradual, Janela de Manutenção, or Eventos")
	policyUpdateCmd.Flags().BoolP("enabled", "e", false, "Enable or disable the policy")
	
	// Cron configuration flags
	policyUpdateCmd.Flags().String("cron-minute", "", "Cron minute (0-59 or *)")
	policyUpdateCmd.Flags().String("cron-hour", "", "Cron hour (0-23 or *)")
	policyUpdateCmd.Flags().String("cron-day", "", "Cron day (1-31 or *)")
	policyUpdateCmd.Flags().String("cron-month", "", "Cron month (1-12 or *)")
	policyUpdateCmd.Flags().String("cron-weekday", "", "Cron weekday (0-7 or *)")
	
	// Event configuration flags
	policyUpdateCmd.Flags().String("event-id", "", "Event ID for Eventos strategy")
	policyUpdateCmd.Flags().Int("event-total", 0, "Total events for Eventos strategy")

	// Delete command flags
	policyDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
}
