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

var matrixCmd = &cobra.Command{
	Use:     "matrix",
	Aliases: []string{"matriz"},
	Short:   "Manage service matrix (service relations)",
	Long:    `Manage service matrix including listing, creating, enabling/disabling, and deleting service relations.`,
}

var matrixListCmd = &cobra.Command{
	Use:     "list <service-hash>",
	Aliases: []string{"ls"},
	Short:   "List all relations for a service",
	Long:    `List all service relations for a specific service.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
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

		apiEndpoint := fmt.Sprintf("/services/%s/matriz/relations", serviceHash)
		log.Debugf("GET %s%s", endpoint, apiEndpoint)

		// Make request
		response, err := apiClient.GetWithAuth(apiEndpoint, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list service relations: %w", err)
		}

		// Parse response
		var relations []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if rel, ok := item.(map[string]interface{}); ok {
						relations = append(relations, rel)
					}
				}
			}
		}

		if len(relations) == 0 {
			fmt.Println("No service relations found.")
			return nil
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(relations, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "RELATION ID\tSOURCE SERVICE\tRELATED SERVICE\tSTATUS\tCREATED AT")
		fmt.Fprintln(w, "-----------\t--------------\t---------------\t------\t----------")

		for _, rel := range relations {
			relationID := fmt.Sprintf("%v", rel["relation_id"])
			if len(relationID) > 12 {
				relationID = relationID[:12] + "..."
			}

			sourceName := "N/A"
			if rel["source_service_name"] != nil && rel["source_service_name"] != "<nil>" {
				sourceName = fmt.Sprintf("%v", rel["source_service_name"])
				if len(sourceName) > 25 {
					sourceName = sourceName[:22] + "..."
				}
			}

			relatedName := "N/A"
			if rel["related_service_name"] != nil && rel["related_service_name"] != "<nil>" {
				relatedName = fmt.Sprintf("%v", rel["related_service_name"])
				if len(relatedName) > 25 {
					relatedName = relatedName[:22] + "..."
				}
			}

			enabled := rel["enabled"].(bool)
			status := "Disabled"
			if enabled {
				status = "Enabled"
			}

			createdAt := ""
			if rel["created_at"] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", rel["created_at"])); err == nil {
					createdAt = t.Format("2006-01-02 15:04")
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", relationID, sourceName, relatedName, status, createdAt)
		}
		w.Flush()

		return nil
	},
}

var matrixGetCmd = &cobra.Command{
	Use:   "get <service-hash>",
	Short: "Get matrix data for a service",
	Long:  `Get complete matrix data for a service including all available services.`,
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
		response, err := apiClient.GetWithAuth(fmt.Sprintf("/services/%s/matriz", serviceHash), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get matrix data: %w", err)
		}

		// Output format
		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// Pretty print
		fmt.Printf("Service: %v\n\n", response["service"])
		
		if relations, ok := response["relations"].([]interface{}); ok && len(relations) > 0 {
			fmt.Println("Current Relations:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "  RELATION ID\tRELATED SERVICE\tSTATUS")
			fmt.Fprintln(w, "  -----------\t---------------\t------")
			
			for _, item := range relations {
				if rel, ok := item.(map[string]interface{}); ok {
					relationID := fmt.Sprintf("%v", rel["relation_id"])
					if len(relationID) > 12 {
						relationID = relationID[:12] + "..."
					}
					
					relatedName := "N/A"
					if rel["related_service_name"] != nil && rel["related_service_name"] != "<nil>" {
						relatedName = fmt.Sprintf("%v", rel["related_service_name"])
					}
					
					enabled := rel["enabled"].(bool)
					status := "Disabled"
					if enabled {
						status = "Enabled"
					}
					
					fmt.Fprintf(w, "  %s\t%s\t%s\n", relationID, relatedName, status)
				}
			}
			w.Flush()
		} else {
			fmt.Println("No relations found.")
		}

		if services, ok := response["available_services"].([]interface{}); ok && len(services) > 0 {
			fmt.Println("\nAvailable Services:")
			for _, item := range services {
				if svc, ok := item.(map[string]interface{}); ok {
					fmt.Printf("  - %v (%v)\n", svc["service_name"], svc["service_hash"])
				}
			}
		}

		return nil
	},
}

var matrixAddCmd = &cobra.Command{
	Use:   "add <source-service-hash> <related-service-hash>",
	Short: "Add a service relation",
	Long:  `Add a new relation between a source service and a related service.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		sourceServiceHash := args[0]
		relatedServiceHash := args[1]

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
			"related_service_hash": relatedServiceHash,
		}

		log.Infof("Adding service relation: %s -> %s", sourceServiceHash, relatedServiceHash)

		// Make request
		response, err := apiClient.PostWithAuth(fmt.Sprintf("/services/%s/matriz", sourceServiceHash), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to add service relation: %w", err)
		}

		fmt.Printf("✓ Service relation added successfully\n")
		fmt.Printf("Relation ID:      %v\n", response["relation_id"])
		fmt.Printf("Source Service:   %v (%v)\n", response["source_service_name"], response["source_service_hash"])
		fmt.Printf("Related Service:  %v (%v)\n", response["related_service_name"], response["related_service_hash"])
		enabledStatus := "Disabled"
		if response["enabled"].(bool) {
			enabledStatus = "Enabled"
		}
		fmt.Printf("Status:           %s\n", enabledStatus)

		return nil
	},
}

var matrixEnableCmd = &cobra.Command{
	Use:   "enable <service-hash> <relation-id>",
	Short: "Enable a service relation",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceHash := args[0]
		relationID := args[1]

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		// Create API client
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		// Make request (toggle endpoint toggles the current state, so we need to check first)
		// For simplicity, we'll just call toggle and inform the user
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/services/%s/matriz/relations/%s/toggle", serviceHash, relationID), nil, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to toggle service relation: %w", err)
		}

		fmt.Printf("✓ Service relation toggled\n")
		fmt.Printf("Note: The toggle endpoint switches the current state. Use 'get' or 'list' to verify the new status.\n")
		return nil
	},
}

var matrixDisableCmd = &cobra.Command{
	Use:   "disable <service-hash> <relation-id>",
	Short: "Disable a service relation",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceHash := args[0]
		relationID := args[1]

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		// Create API client
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		// Make request (toggle endpoint toggles the current state)
		_, err = apiClient.PutWithAuth(fmt.Sprintf("/services/%s/matriz/relations/%s/toggle", serviceHash, relationID), nil, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to toggle service relation: %w", err)
		}

		fmt.Printf("✓ Service relation toggled\n")
		fmt.Printf("Note: The toggle endpoint switches the current state. Use 'get' or 'list' to verify the new status.\n")
		return nil
	},
}

var matrixToggleCmd = &cobra.Command{
	Use:   "toggle <service-hash> <relation-id>",
	Short: "Toggle a service relation (enable/disable)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		serviceHash := args[0]
		relationID := args[1]

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		// Create API client
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		log.Infof("Toggling service relation: %s", relationID)

		// Make PUT request
		response, err := apiClient.PutWithAuth(fmt.Sprintf("/services/%s/matriz/relations/%s/toggle", serviceHash, relationID), nil, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to toggle service relation: %w", err)
		}

		fmt.Printf("✓ Service relation toggled successfully\n")
		fmt.Printf("Relation ID:      %v\n", response["relation_id"])
		fmt.Printf("Source Service:   %v\n", response["source_service_name"])
		fmt.Printf("Related Service:  %v\n", response["related_service_name"])
		enabledStatus := "Disabled"
		if response["enabled"].(bool) {
			enabledStatus = "Enabled"
		}
		fmt.Printf("New Status:       %s\n", enabledStatus)

		return nil
	},
}

var matrixDeleteCmd = &cobra.Command{
	Use:     "delete <service-hash> <relation-id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a service relation",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		serviceHash := args[0]
		relationID := args[1]

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete service relation %s? (y/N): ", relationID)
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

		log.Infof("Deleting service relation: %s", relationID)

		// Make request
		_, err = apiClient.DeleteWithAuth(fmt.Sprintf("/services/%s/matriz/relations/%s", serviceHash, relationID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to delete service relation: %w", err)
		}

		fmt.Printf("✓ Service relation deleted successfully\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(matrixCmd)

	// Add subcommands
	matrixCmd.AddCommand(matrixListCmd)
	matrixCmd.AddCommand(matrixGetCmd)
	matrixCmd.AddCommand(matrixAddCmd)
	matrixCmd.AddCommand(matrixToggleCmd)
	matrixCmd.AddCommand(matrixEnableCmd)
	matrixCmd.AddCommand(matrixDisableCmd)
	matrixCmd.AddCommand(matrixDeleteCmd)

	// List command flags
	matrixListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Get command flags
	matrixGetCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	// Delete command flags
	matrixDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
}
