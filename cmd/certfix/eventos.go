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

var eventosCmd = &cobra.Command{
	Use:     "events",
	Aliases: []string{"event", "eventos", "evento"},
	Short:   "Manage events",
	Long:    `Manage events including listing, creating, updating, enabling/disabling, and deleting events.`,
}

var eventosListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all events",
	Long:    `List all events with optional filtering by severity or enabled status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		// Get flags
		severity, _ := cmd.Flags().GetString("severity")
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
			apiEndpoint = "/events/enabled"
		} else if severity != "" {
			apiEndpoint = fmt.Sprintf("/events/severity/%s", severity)
		} else {
			apiEndpoint = "/events"
		}

		log.Debugf("GET %s%s", endpoint, apiEndpoint)

		// Make request
		response, err := apiClient.GetWithAuth(apiEndpoint, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list events: %w", err)
		}

		// Parse response
		var eventos []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if evento, ok := item.(map[string]interface{}); ok {
						eventos = append(eventos, evento)
					}
				}
			}
		}

		if len(eventos) == 0 {
			fmt.Println("No events found.")
			return nil
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(eventos, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tEXTERNAL ID\tCOUNTER\tSEVERITY\tSTATUS\tCREATED AT")
		fmt.Fprintln(w, "----\t----\t-----------\t-------\t--------\t------\t----------")

		for _, evento := range eventos {
			id := fmt.Sprintf("%v", evento["event_id"])
			name := fmt.Sprintf("%v", evento["name"])
			severity := strings.ToUpper(fmt.Sprintf("%v", evento["severity"]))
			enabled := evento["enabled"].(bool)
			status := "Inactive"
			if enabled {
				status = "Active"
			}
			createdAt := ""
			if evento["created_at"] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", evento["created_at"])); err == nil {
					createdAt = t.Format("2006-01-02 15:04")
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%s\t%s\t%s\n", id, name, evento["external_id"], evento["counter"], severity, status, createdAt)
		}
		w.Flush()

		return nil
	},
}

var eventosGetCmd = &cobra.Command{
	Use:   "get <event-id>",
	Short: "Get details of a specific event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eventoID := args[0]
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
		response, err := apiClient.GetWithAuth(fmt.Sprintf("/events/%s", eventoID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get event: %w", err)
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Pretty print
		fmt.Printf("ID:          %v\n", response["event_id"])
		fmt.Printf("Name:        %v\n", response["name"])
		fmt.Printf("Severity:    %v\n", strings.ToUpper(fmt.Sprintf("%v", response["severity"])))
		enabled := response["enabled"].(bool)
		status := "Inactive"
		if enabled {
			status = "Active"
		}
		fmt.Printf("Status:      %s\n", status)
		fmt.Printf("External ID: %v\n", response["external_id"])
		fmt.Printf("Counter:     %v\n", response["counter"])
		fmt.Printf("Reset Time:  %v %v\n", response["reset_time_value"], response["reset_time_unit"])
		if response["last_event_at"] != nil {
			fmt.Printf("Last Event:  %v\n", response["last_event_at"])
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

var eventosCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new event",
	Long:  `Create a new event with specified name, severity, and status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		severity, _ := cmd.Flags().GetString("severity")
		enabled, _ := cmd.Flags().GetBool("enabled")
		resetUnit, _ := cmd.Flags().GetString("reset-unit")
		resetValue, _ := cmd.Flags().GetInt("reset-value")

		// Validate required fields
		if name == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("name is required (use --name)")
		}
		if severity == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("severity is required (use --severity)")
		}

		// Validate severity
		validSeverities := []string{"low", "medium", "high", "critical"}
		severityValid := false
		for _, v := range validSeverities {
			if strings.ToLower(severity) == v {
				severityValid = true
				break
			}
		}
		if !severityValid {
			cmd.SilenceUsage = true
			return fmt.Errorf("invalid severity: %s (must be one of: low, medium, high, critical)", severity)
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
			"name":             name,
			"severity":         strings.ToLower(severity),
			"enabled":          enabled,
			"reset_time_unit":  resetUnit,
			"reset_time_value": resetValue,
		}

		log.Infof("Creating event: %s", name)

		// Make request
		response, err := apiClient.PostWithAuth("/events", payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to create event: %w", err)
		}

		fmt.Printf("✓ Event created successfully\n")
		fmt.Printf("ID:       %v\n", response["event_id"])
		fmt.Printf("Name:     %v\n", response["name"])
		fmt.Printf("Severity: %v\n", strings.ToUpper(fmt.Sprintf("%v", response["severity"])))
		enabledStatus := "Inactive"
		if response["enabled"].(bool) {
			enabledStatus = "Active"
		}
		fmt.Printf("Status:   %s\n", enabledStatus)

		return nil
	},
}

var eventosUpdateCmd = &cobra.Command{
	Use:   "update <event-id>",
	Short: "Update an existing event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		eventoID := args[0]

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		severity, _ := cmd.Flags().GetString("severity")
		enabled := cmd.Flags().Changed("enabled")
		enabledValue, _ := cmd.Flags().GetBool("enabled")
		resetUnit, _ := cmd.Flags().GetString("reset-unit")
		resetValue, _ := cmd.Flags().GetInt("reset-value")

		// Build update payload
		payload := make(map[string]interface{})

		if name != "" {
			payload["name"] = name
		}

		if severity != "" {
			// Validate severity
			validSeverities := []string{"low", "medium", "high", "critical"}
			severityValid := false
			for _, v := range validSeverities {
				if strings.ToLower(severity) == v {
					severityValid = true
					break
				}
			}
			if !severityValid {
				cmd.SilenceUsage = true
				return fmt.Errorf("invalid severity: %s (must be one of: low, medium, high, critical)", severity)
			}
			payload["severity"] = strings.ToLower(severity)
		}

		if enabled {
			payload["enabled"] = enabledValue
		}

		if cmd.Flags().Changed("reset-unit") {
			payload["reset_time_unit"] = resetUnit
		}

		if cmd.Flags().Changed("reset-value") {
			payload["reset_time_value"] = resetValue
		}

		if len(payload) == 0 {
			cmd.SilenceUsage = true
			return fmt.Errorf("no fields to update (use --name, --severity, or --enabled)")
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

		log.Infof("Updating event: %s", eventoID)

		// Make PUT request
		response, err := apiClient.PutWithAuth(fmt.Sprintf("/events/%s", eventoID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to update event: %w", err)
		}

		fmt.Printf("✓ Event updated successfully\n")
		fmt.Printf("ID:       %v\n", response["event_id"])
		fmt.Printf("Name:     %v\n", response["name"])
		fmt.Printf("Severity: %v\n", strings.ToUpper(fmt.Sprintf("%v", response["severity"])))
		enabledStatus := "Inactive"
		if response["enabled"].(bool) {
			enabledStatus = "Active"
		}
		fmt.Printf("Status:   %s\n", enabledStatus)

		return nil
	},
}

var eventosEnableCmd = &cobra.Command{
	Use:   "enable <event-id>",
	Short: "Enable an event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eventoID := args[0]

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
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/events/%s", eventoID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to enable event: %w", err)
		}

		fmt.Printf("✓ Event enabled successfully\n")
		return nil
	},
}

var eventosDisableCmd = &cobra.Command{
	Use:   "disable <event-id>",
	Short: "Disable an event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eventoID := args[0]

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
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/events/%s", eventoID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to disable event: %w", err)
		}

		fmt.Printf("✓ Event disabled successfully\n")
		return nil
	},
}

var eventosDeleteCmd = &cobra.Command{
	Use:     "delete <event-id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete an event",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		eventoID := args[0]

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete event %s? (y/N): ", eventoID)
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

		log.Infof("Deleting event: %s", eventoID)

		// Make request
		_, err = apiClient.DeleteWithAuth(fmt.Sprintf("/events/%s", eventoID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to delete event: %w", err)
		}

		fmt.Printf("✓ Event deleted successfully\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(eventosCmd)

	// Add subcommands
	eventosCmd.AddCommand(eventosListCmd)
	eventosCmd.AddCommand(eventosGetCmd)
	eventosCmd.AddCommand(eventosCreateCmd)
	eventosCmd.AddCommand(eventosUpdateCmd)
	eventosCmd.AddCommand(eventosEnableCmd)
	eventosCmd.AddCommand(eventosDisableCmd)
	eventosCmd.AddCommand(eventosDeleteCmd)

	// List command flags
	eventosListCmd.Flags().StringP("severity", "s", "", "Filter by severity (low, medium, high, critical)")
	eventosListCmd.Flags().BoolP("enabled", "e", false, "Show only enabled events")
	eventosListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Get command flags
	eventosGetCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Create command flags
	eventosCreateCmd.Flags().StringP("name", "n", "", "Name of the event (required)")
	eventosCreateCmd.Flags().StringP("severity", "s", "", "Severity level: low, medium, high, critical (required)")
	eventosCreateCmd.Flags().BoolP("enabled", "e", true, "Enable the event immediately (default: true)")
	eventosCreateCmd.Flags().String("reset-unit", "hours", "Reset unit: minutes, hours, days")
	eventosCreateCmd.Flags().Int("reset-value", 0, "Reset counter if no events within this value (0 = never)")
	eventosCreateCmd.MarkFlagRequired("name")
	eventosCreateCmd.MarkFlagRequired("severity")

	// Update command flags
	eventosUpdateCmd.Flags().StringP("name", "n", "", "New name for the event")
	eventosUpdateCmd.Flags().StringP("severity", "s", "", "New severity level: low, medium, high, critical")
	eventosUpdateCmd.Flags().BoolP("enabled", "e", false, "Enable or disable the event")
	eventosUpdateCmd.Flags().String("reset-unit", "", "New reset unit: minutes, hours, days")
	eventosUpdateCmd.Flags().Int("reset-value", 0, "New reset counter value")

	// Delete command flags
	eventosDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
}
