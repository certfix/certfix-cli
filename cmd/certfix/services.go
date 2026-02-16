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

var servicesRotateCmd = &cobra.Command{
	Use:   "rotate <service-hash[,service-hash,...]>",
	Short: "Rotate certificate(s) for one or more services",
	Long:  `Rotate the certificate for one or more services by hash. Example: certfix service rotate id1,id2,id3`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hashes := strings.Split(args[0], ",")
		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)
		var failed []string
		for _, hash := range hashes {
			hash = strings.TrimSpace(hash)
			if hash == "" { continue }
			fmt.Printf("Rotating certificate for service: %s... ", hash)
			_, err := apiClient.PostWithAuth("/services/"+hash+"/certificates/rotate", map[string]interface{}{}, token)
			if err != nil {
				fmt.Printf("Failed: %v\n", err)
				failed = append(failed, hash)
			} else {
				fmt.Printf("OK\n")
			}
		}
		if len(failed) > 0 {
			return fmt.Errorf("Failed to rotate for: %s", strings.Join(failed, ", "))
		}
		return nil
	},
}

var servicesCmd = &cobra.Command{
	Use:     "services",
	Aliases: []string{"service", "svc"},
	Short:   "Manage services",
	Long:    `Manage services including listing, creating, updating, activating/deactivating, and deleting services.`,
}

var servicesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all services",
	Long:    `List all services with optional filtering by active status or service group.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		// Get flags
		activeOnly, _ := cmd.Flags().GetBool("active")
		groupID, _ := cmd.Flags().GetString("group")
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
		if activeOnly {
			apiEndpoint = "/services/active"
		} else if groupID != "" {
			apiEndpoint = fmt.Sprintf("/services/group/%s", groupID)
		} else {
			apiEndpoint = "/services"
		}

		log.Debugf("GET %s%s", endpoint, apiEndpoint)

		// Make request
		response, err := apiClient.GetWithAuth(apiEndpoint, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list services: %w", err)
		}

		// Parse response
		var services []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if svc, ok := item.(map[string]interface{}); ok {
						services = append(services, svc)
					}
				}
			}
		}

		if len(services) == 0 {
			fmt.Println("No services found.")
			return nil
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(services, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "HASH\tNAME\tGROUP\tPOLICY\tSTATUS\tCREATED AT")
		fmt.Fprintln(w, "----\t----\t-----\t------\t------\t----------")

		for _, svc := range services {
			hash := fmt.Sprintf("%v", svc["service_hash"])
			if len(hash) > 12 {
				hash = hash[:12] + "..."
			}
			name := fmt.Sprintf("%v", svc["service_name"])
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			
			groupName := "N/A"
			if svc["service_group_name"] != nil && svc["service_group_name"] != "<nil>" {
				groupName = fmt.Sprintf("%v", svc["service_group_name"])
				if len(groupName) > 20 {
					groupName = groupName[:17] + "..."
				}
			}
			
			policyName := "N/A"
			if svc["politica_name"] != nil && svc["politica_name"] != "<nil>" {
				policyName = fmt.Sprintf("%v", svc["politica_name"])
				if len(policyName) > 20 {
					policyName = policyName[:17] + "..."
				}
			}
			
			active := svc["active"].(bool)
			status := "Inactive"
			if active {
				status = "Active"
			}
			
			createdAt := ""
			if svc["created_at"] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", svc["created_at"])); err == nil {
					createdAt = t.Format("2006-01-02 15:04")
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", hash, name, groupName, policyName, status, createdAt)
		}
		w.Flush()

		return nil
	},
}

var servicesGetCmd = &cobra.Command{
	Use:   "get <service-hash>",
	Short: "Get details of a specific service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceHash := args[0]
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
		response, err := apiClient.GetWithAuth(fmt.Sprintf("/services/%s", serviceHash), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get service: %w", err)
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Pretty print
		fmt.Printf("Hash:         %v\n", response["service_hash"])
		fmt.Printf("Name:         %v\n", response["service_name"])
		
		groupName := "N/A"
		if response["service_group_name"] != nil && response["service_group_name"] != "<nil>" {
			groupName = fmt.Sprintf("%v", response["service_group_name"])
		}
		groupID := "N/A"
		if response["service_group_id"] != nil && response["service_group_id"] != "<nil>" {
			groupID = fmt.Sprintf("%v", response["service_group_id"])
		}
		fmt.Printf("Group:        %s (%s)\n", groupName, groupID)
		
		policyName := "N/A"
		if response["politica_name"] != nil && response["politica_name"] != "<nil>" {
			policyName = fmt.Sprintf("%v", response["politica_name"])
		}
		policyID := "N/A"
		if response["politica_id"] != nil && response["politica_id"] != "<nil>" {
			policyID = fmt.Sprintf("%v", response["politica_id"])
		}
		fmt.Printf("Policy:       %s (%s)\n", policyName, policyID)
		
		webhookURL := "N/A"
		if response["webhook_url"] != nil && response["webhook_url"] != "<nil>" {
			webhookURL = fmt.Sprintf("%v", response["webhook_url"])
		}
		fmt.Printf("Webhook URL:  %s\n", webhookURL)
		
		active := response["active"].(bool)
		status := "Inactive"
		if active {
			status = "Active"
		}
		fmt.Printf("Status:       %s\n", status)
		
		if response["created_at"] != nil {
			fmt.Printf("Created At:   %v\n", response["created_at"])
		}
		if response["updated_at"] != nil {
			fmt.Printf("Updated At:   %v\n", response["updated_at"])
		}

		return nil
	},
}

var servicesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new service",
	Long: `Create a new service with specified name, webhook URL, service group, and policy.

You can optionally specify a custom hash using --hash. If provided, the hash must be unique
and will be validated before creating the service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		serviceHash, _ := cmd.Flags().GetString("hash")
		webhookURL, _ := cmd.Flags().GetString("webhook")
		groupID, _ := cmd.Flags().GetString("group")
		policyID, _ := cmd.Flags().GetString("policy")
		active, _ := cmd.Flags().GetBool("active")

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

		// If hash is provided, check for duplicates
		if serviceHash != "" {
			log.Debugf("Checking if hash already exists: %s", serviceHash)
			_, err := apiClient.GetWithAuth(fmt.Sprintf("/services/%s", serviceHash), token)
			if err == nil {
				// Service exists with this hash
				cmd.SilenceUsage = true
				return fmt.Errorf("service hash '%s' already exists. Please choose a different hash", serviceHash)
			}
			// If error is not found (404), we can proceed
			log.Debugf("Hash is available: %s", serviceHash)
		}

		// Prepare payload
		payload := map[string]interface{}{
			"service_name": name,
			"active":       active,
		}

		if serviceHash != "" {
			payload["service_hash"] = serviceHash
		}

		if webhookURL != "" {
			payload["webhook_url"] = webhookURL
		}
		if groupID != "" {
			payload["service_group_id"] = groupID
		}
		if policyID != "" {
			payload["politica_id"] = policyID
		}

		log.Infof("Creating service: %s", name)

		// Make request
		response, err := apiClient.PostWithAuth("/services", payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to create service: %w", err)
		}

		fmt.Printf("✓ Service created successfully\n")
		fmt.Printf("Hash:         %v\n", response["service_hash"])
		fmt.Printf("Name:         %v\n", response["service_name"])
		
		groupName := "N/A"
		if response["service_group_name"] != nil && response["service_group_name"] != "<nil>" {
			groupName = fmt.Sprintf("%v", response["service_group_name"])
		}
		fmt.Printf("Group:        %s\n", groupName)
		
		policyName := "N/A"
		if response["politica_name"] != nil && response["politica_name"] != "<nil>" {
			policyName = fmt.Sprintf("%v", response["politica_name"])
		}
		fmt.Printf("Policy:       %s\n", policyName)
		
		activeStatus := "Inactive"
		if response["active"].(bool) {
			activeStatus = "Active"
		}
		fmt.Printf("Status:       %s\n", activeStatus)

		return nil
	},
}

var servicesUpdateCmd = &cobra.Command{
	Use:   "update <service-hash>",
	Short: "Update an existing service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		serviceHash := args[0]

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		webhookURL, _ := cmd.Flags().GetString("webhook")
		groupID, _ := cmd.Flags().GetString("group")
		policyID, _ := cmd.Flags().GetString("policy")
		active := cmd.Flags().Changed("active")
		activeValue, _ := cmd.Flags().GetBool("active")
		clearWebhook, _ := cmd.Flags().GetBool("clear-webhook")
		clearGroup, _ := cmd.Flags().GetBool("clear-group")
		clearPolicy, _ := cmd.Flags().GetBool("clear-policy")

		// Build update payload
		payload := make(map[string]interface{})

		if name != "" {
			payload["service_name"] = name
		}

		if webhookURL != "" {
			payload["webhook_url"] = webhookURL
		} else if clearWebhook {
			payload["webhook_url"] = nil
		}

		if groupID != "" {
			payload["service_group_id"] = groupID
		} else if clearGroup {
			payload["service_group_id"] = nil
		}

		if policyID != "" {
			payload["politica_id"] = policyID
		} else if clearPolicy {
			payload["politica_id"] = nil
		}

		if active {
			payload["active"] = activeValue
		}

		if len(payload) == 0 {
			cmd.SilenceUsage = true
			return fmt.Errorf("no fields to update (use --name, --webhook, --group, --policy, --active, or clear flags)")
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

		log.Infof("Updating service: %s", serviceHash)

		// Make PUT request
		response, err := apiClient.PutWithAuth(fmt.Sprintf("/services/%s", serviceHash), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to update service: %w", err)
		}

		fmt.Printf("✓ Service updated successfully\n")
		fmt.Printf("Hash:         %v\n", response["service_hash"])
		fmt.Printf("Name:         %v\n", response["service_name"])
		
		groupName := "N/A"
		if response["service_group_name"] != nil && response["service_group_name"] != "<nil>" {
			groupName = fmt.Sprintf("%v", response["service_group_name"])
		}
		fmt.Printf("Group:        %s\n", groupName)
		
		policyName := "N/A"
		if response["politica_name"] != nil && response["politica_name"] != "<nil>" {
			policyName = fmt.Sprintf("%v", response["politica_name"])
		}
		fmt.Printf("Policy:       %s\n", policyName)
		
		activeStatus := "Inactive"
		if response["active"].(bool) {
			activeStatus = "Active"
		}
		fmt.Printf("Status:       %s\n", activeStatus)

		return nil
	},
}

var servicesActivateCmd = &cobra.Command{
	Use:   "activate <service-hash>",
	Short: "Activate a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceHash := args[0]

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
			"active": true,
		}

		// Make request
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/services/%s", serviceHash), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to activate service: %w", err)
		}

		fmt.Printf("✓ Service activated successfully\n")
		return nil
	},
}

var servicesDeactivateCmd = &cobra.Command{
	Use:   "deactivate <service-hash>",
	Short: "Deactivate a service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceHash := args[0]

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
			"active": false,
		}

		// Make request
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/services/%s", serviceHash), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to deactivate service: %w", err)
		}

		fmt.Printf("✓ Service deactivated successfully\n")
		return nil
	},
}

var servicesDeleteCmd = &cobra.Command{
	Use:     "delete <service-hash>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a service",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		serviceHash := args[0]

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete service %s? (y/N): ", serviceHash)
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

		log.Infof("Deleting service: %s", serviceHash)

		// Make request
		_, err = apiClient.DeleteWithAuth(fmt.Sprintf("/services/%s", serviceHash), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to delete service: %w", err)
		}

		fmt.Printf("✓ Service deleted successfully\n")
		return nil
	},
}

var servicesGenerateHashCmd = &cobra.Command{
	Use:   "generate-hash <service-name>",
	Short: "Generate a hash for a service name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]
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

		// Prepare payload
		payload := map[string]interface{}{
			"service_name": serviceName,
		}

		// Make request
		response, err := apiClient.PostWithAuth("/services/generate-hash", payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to generate hash: %w", err)
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Service Name: %s\n", serviceName)
		fmt.Printf("Service Hash: %v\n", response["service_hash"])

		return nil
	},
}

func init() {
	rootCmd.AddCommand(servicesCmd)

	// Add subcommands
	servicesCmd.AddCommand(servicesListCmd)
	servicesCmd.AddCommand(servicesGetCmd)
	servicesCmd.AddCommand(servicesCreateCmd)
	servicesCmd.AddCommand(servicesUpdateCmd)
	servicesCmd.AddCommand(servicesActivateCmd)
	servicesCmd.AddCommand(servicesDeactivateCmd)
	servicesCmd.AddCommand(servicesDeleteCmd)
	servicesCmd.AddCommand(servicesGenerateHashCmd)

		// Add rotate command
		servicesCmd.AddCommand(servicesRotateCmd)

	// List command flags
	servicesListCmd.Flags().BoolP("active", "a", false, "Show only active services")
	servicesListCmd.Flags().StringP("group", "g", "", "Filter by service group ID")
	servicesListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Get command flags
	servicesGetCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Create command flags
	servicesCreateCmd.Flags().StringP("name", "n", "", "Name of the service (required)")
	servicesCreateCmd.Flags().String("hash", "", "Custom service hash (optional, must be unique)")
	servicesCreateCmd.Flags().StringP("webhook", "w", "", "Webhook URL for the service")
	servicesCreateCmd.Flags().StringP("group", "g", "", "Service group ID")
	servicesCreateCmd.Flags().StringP("policy", "p", "", "Policy ID")
	servicesCreateCmd.Flags().BoolP("active", "a", true, "Activate the service immediately (default: true)")
	servicesCreateCmd.MarkFlagRequired("name")

	// Update command flags
	servicesUpdateCmd.Flags().StringP("name", "n", "", "New name for the service")
	servicesUpdateCmd.Flags().StringP("webhook", "w", "", "New webhook URL for the service")
	servicesUpdateCmd.Flags().StringP("group", "g", "", "New service group ID")
	servicesUpdateCmd.Flags().StringP("policy", "p", "", "New policy ID")
	servicesUpdateCmd.Flags().BoolP("active", "a", false, "Activate or deactivate the service")
	servicesUpdateCmd.Flags().Bool("clear-webhook", false, "Clear the webhook URL")
	servicesUpdateCmd.Flags().Bool("clear-group", false, "Clear the service group")
	servicesUpdateCmd.Flags().Bool("clear-policy", false, "Clear the policy")

	// Delete command flags
	servicesDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")

	// Generate hash command flags
	servicesGenerateHashCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
}
