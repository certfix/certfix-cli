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

var serviceGroupsCmd = &cobra.Command{
	Use:     "service-groups",
	Aliases: []string{"service-group", "svc-groups", "svc-group"},
	Short:   "Manage service groups",
	Long:    `Manage service groups including listing, creating, updating, enabling/disabling, and deleting service groups.`,
}

var serviceGroupsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all service groups",
	Long:    `List all service groups with optional filtering by enabled status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		// Get flags
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
			apiEndpoint = "/service-groups/enabled"
		} else {
			apiEndpoint = "/service-groups"
		}

		log.Debugf("GET %s%s", endpoint, apiEndpoint)

		// Make request
		response, err := apiClient.GetWithAuth(apiEndpoint, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list service groups: %w", err)
		}

		// Parse response
		var serviceGroups []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if sg, ok := item.(map[string]interface{}); ok {
						serviceGroups = append(serviceGroups, sg)
					}
				}
			}
		}

		if len(serviceGroups) == 0 {
			fmt.Println("No service groups found.")
			return nil
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(serviceGroups, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION\tSTATUS\tCREATED AT")
		fmt.Fprintln(w, "----\t----\t-----------\t------\t----------")

		for _, sg := range serviceGroups {
			id := fmt.Sprintf("%v", sg["service_group_id"])
			name := fmt.Sprintf("%v", sg["name"])
			description := fmt.Sprintf("%v", sg["description"])
			if len(description) > 50 {
				description = description[:47] + "..."
			}
			enabled := sg["enabled"].(bool)
			status := "Inactive"
			if enabled {
				status = "Active"
			}
			createdAt := ""
			if sg["created_at"] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", sg["created_at"])); err == nil {
					createdAt = t.Format("2006-01-02 15:04")
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, name, description, status, createdAt)
		}
		w.Flush()

		return nil
	},
}

var serviceGroupsGetCmd = &cobra.Command{
	Use:   "get <service-group-id>",
	Short: "Get details of a specific service group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceGroupID := args[0]
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
		response, err := apiClient.GetWithAuth(fmt.Sprintf("/service-groups/%s", serviceGroupID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get service group: %w", err)
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Pretty print
		fmt.Printf("ID:          %v\n", response["service_group_id"])
		fmt.Printf("Name:        %v\n", response["name"])
		fmt.Printf("Description: %v\n", response["description"])
		enabled := response["enabled"].(bool)
		status := "Inactive"
		if enabled {
			status = "Active"
		}
		fmt.Printf("Status:      %s\n", status)
		if response["created_at"] != nil {
			fmt.Printf("Created At:  %v\n", response["created_at"])
		}
		if response["updated_at"] != nil {
			fmt.Printf("Updated At:  %v\n", response["updated_at"])
		}

		return nil
	},
}

var serviceGroupsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new service group",
	Long:  `Create a new service group with specified name, description, and status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		enabled, _ := cmd.Flags().GetBool("enabled")

		// Validate required fields
		if name == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("name is required (use --name)")
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
			"name":        name,
			"description": description,
			"enabled":     enabled,
		}

		log.Infof("Creating service group: %s", name)

		// Make request
		response, err := apiClient.PostWithAuth("/service-groups", payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to create service group: %w", err)
		}

		fmt.Printf("✓ Service group created successfully\n")
		fmt.Printf("ID:          %v\n", response["service_group_id"])
		fmt.Printf("Name:        %v\n", response["name"])
		fmt.Printf("Description: %v\n", response["description"])
		enabledStatus := "Inactive"
		if response["enabled"].(bool) {
			enabledStatus = "Active"
		}
		fmt.Printf("Status:      %s\n", enabledStatus)

		return nil
	},
}

var serviceGroupsUpdateCmd = &cobra.Command{
	Use:   "update <service-group-id>",
	Short: "Update an existing service group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		serviceGroupID := args[0]

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		enabled := cmd.Flags().Changed("enabled")
		enabledValue, _ := cmd.Flags().GetBool("enabled")

		// Build update payload
		payload := make(map[string]interface{})

		if name != "" {
			payload["name"] = name
		}

		if description != "" {
			payload["description"] = description
		}

		if enabled {
			payload["enabled"] = enabledValue
		}

		if len(payload) == 0 {
			cmd.SilenceUsage = true
			return fmt.Errorf("no fields to update (use --name, --description, or --enabled)")
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

		log.Infof("Updating service group: %s", serviceGroupID)

		// Make PUT request
		response, err := apiClient.PutWithAuth(fmt.Sprintf("/service-groups/%s", serviceGroupID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to update service group: %w", err)
		}

		fmt.Printf("✓ Service group updated successfully\n")
		fmt.Printf("ID:          %v\n", response["service_group_id"])
		fmt.Printf("Name:        %v\n", response["name"])
		fmt.Printf("Description: %v\n", response["description"])
		enabledStatus := "Inactive"
		if response["enabled"].(bool) {
			enabledStatus = "Active"
		}
		fmt.Printf("Status:      %s\n", enabledStatus)

		return nil
	},
}

var serviceGroupsEnableCmd = &cobra.Command{
	Use:   "enable <service-group-id>",
	Short: "Enable a service group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceGroupID := args[0]

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
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/service-groups/%s", serviceGroupID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to enable service group: %w", err)
		}

		fmt.Printf("✓ Service group enabled successfully\n")
		return nil
	},
}

var serviceGroupsDisableCmd = &cobra.Command{
	Use:   "disable <service-group-id>",
	Short: "Disable a service group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceGroupID := args[0]

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
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/service-groups/%s", serviceGroupID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to disable service group: %w", err)
		}

		fmt.Printf("✓ Service group disabled successfully\n")
		return nil
	},
}

var serviceGroupsDeleteCmd = &cobra.Command{
	Use:     "delete <service-group-id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a service group",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		serviceGroupID := args[0]

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete service group %s? (y/N): ", serviceGroupID)
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

		log.Infof("Deleting service group: %s", serviceGroupID)

		// Make request
		_, err = apiClient.DeleteWithAuth(fmt.Sprintf("/service-groups/%s", serviceGroupID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to delete service group: %w", err)
		}

		fmt.Printf("✓ Service group deleted successfully\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serviceGroupsCmd)

	// Add subcommands
	serviceGroupsCmd.AddCommand(serviceGroupsListCmd)
	serviceGroupsCmd.AddCommand(serviceGroupsGetCmd)
	serviceGroupsCmd.AddCommand(serviceGroupsCreateCmd)
	serviceGroupsCmd.AddCommand(serviceGroupsUpdateCmd)
	serviceGroupsCmd.AddCommand(serviceGroupsEnableCmd)
	serviceGroupsCmd.AddCommand(serviceGroupsDisableCmd)
	serviceGroupsCmd.AddCommand(serviceGroupsDeleteCmd)

	// List command flags
	serviceGroupsListCmd.Flags().BoolP("enabled", "e", false, "Show only enabled service groups")
	serviceGroupsListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Get command flags
	serviceGroupsGetCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Create command flags
	serviceGroupsCreateCmd.Flags().StringP("name", "n", "", "Name of the service group (required)")
	serviceGroupsCreateCmd.Flags().StringP("description", "d", "", "Description of the service group")
	serviceGroupsCreateCmd.Flags().BoolP("enabled", "e", true, "Enable the service group immediately (default: true)")
	serviceGroupsCreateCmd.MarkFlagRequired("name")

	// Update command flags
	serviceGroupsUpdateCmd.Flags().StringP("name", "n", "", "New name for the service group")
	serviceGroupsUpdateCmd.Flags().StringP("description", "d", "", "New description for the service group")
	serviceGroupsUpdateCmd.Flags().BoolP("enabled", "e", false, "Enable or disable the service group")

	// Delete command flags
	serviceGroupsDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
}
