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
	"github.com/spf13/cobra"
)

var userGroupsCmd = &cobra.Command{
	Use:               "user-groups",
	Aliases:           []string{"ug", "user-group"},
	Short:             "Manage user groups",
	Long:              `Manage user groups including listing, creating, updating, enabling/disabling, and deleting groups.`,
	PersistentPreRunE: requireSuperuser,
}

var ugListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all user groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth("/user-groups", token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list user groups: %w", err)
		}

		var groups []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if g, ok := item.(map[string]interface{}); ok {
						groups = append(groups, g)
					}
				}
			}
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(groups, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(groups) == 0 {
			fmt.Println("No user groups found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tCREATED AT")
		fmt.Fprintln(w, "--\t----\t------\t----------")

		for _, g := range groups {
			id := fmt.Sprintf("%v", g["id"])
			name := fmt.Sprintf("%v", g["name"])
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			status := "Disabled"
			if enabled, ok := g["enabled"].(bool); ok && enabled {
				status = "Enabled"
			}
			createdAt := ""
			if g["created_at"] != nil {
				if t, err := time.Parse(time.RFC3339, fmt.Sprintf("%v", g["created_at"])); err == nil {
					createdAt = t.Format("2006-01-02 15:04")
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, name, status, createdAt)
		}
		w.Flush()

		return nil
	},
}

var ugGetCmd = &cobra.Command{
	Use:   "get <group-id>",
	Short: "Get details of a specific user group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupID := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth(fmt.Sprintf("/user-groups/%s", groupID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get user group: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("ID:         %v\n", response["id"])
		fmt.Printf("Name:       %v\n", response["name"])
		status := "Disabled"
		if enabled, ok := response["enabled"].(bool); ok && enabled {
			status = "Enabled"
		}
		fmt.Printf("Status:     %s\n", status)
		if response["created_at"] != nil {
			fmt.Printf("Created At: %v\n", response["created_at"])
		}

		return nil
	},
}

var ugCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user group",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		enabled, _ := cmd.Flags().GetBool("enabled")
		outputFormat, _ := cmd.Flags().GetString("output")

		if name == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("name is required (use --name)")
		}

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		payload := map[string]interface{}{
			"name":    name,
			"enabled": enabled,
		}

		response, err := apiClient.PostWithAuth("/user-groups", payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to create user group: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ User group created successfully\n")
		fmt.Printf("ID:   %v\n", response["id"])
		fmt.Printf("Name: %v\n", response["name"])

		return nil
	},
}

var ugUpdateCmd = &cobra.Command{
	Use:   "update <group-id>",
	Short: "Update an existing user group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupID := args[0]
		name, _ := cmd.Flags().GetString("name")
		enabledChanged := cmd.Flags().Changed("enabled")
		enabledValue, _ := cmd.Flags().GetBool("enabled")
		outputFormat, _ := cmd.Flags().GetString("output")

		payload := make(map[string]interface{})
		if name != "" {
			payload["name"] = name
		}
		if enabledChanged {
			payload["enabled"] = enabledValue
		}

		if len(payload) == 0 {
			cmd.SilenceUsage = true
			return fmt.Errorf("no fields to update (use --name or --enabled)")
		}

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.PutWithAuth(fmt.Sprintf("/user-groups/%s", groupID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to update user group: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ User group updated successfully\n")
		fmt.Printf("ID:   %v\n", response["id"])
		fmt.Printf("Name: %v\n", response["name"])

		return nil
	},
}

var ugDeleteCmd = &cobra.Command{
	Use:     "delete <group-id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a user group",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Are you sure you want to delete user group %s? (y/N): ", groupID)
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

		_, err = apiClient.DeleteWithAuth(fmt.Sprintf("/user-groups/%s", groupID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to delete user group: %w", err)
		}

		fmt.Printf("✓ User group deleted successfully\n")
		return nil
	},
}

var ugEnableCmd = &cobra.Command{
	Use:   "enable <group-id>",
	Short: "Enable a user group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupID := args[0]

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		_, err = apiClient.PutWithAuth(fmt.Sprintf("/user-groups/%s", groupID), map[string]interface{}{"enabled": true}, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to enable user group: %w", err)
		}

		fmt.Printf("✓ User group enabled successfully\n")
		return nil
	},
}

var ugDisableCmd = &cobra.Command{
	Use:   "disable <group-id>",
	Short: "Disable a user group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupID := args[0]

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		_, err = apiClient.PatchWithAuth(fmt.Sprintf("/user-groups/%s/disable", groupID), nil, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to disable user group: %w", err)
		}

		fmt.Printf("✓ User group disabled successfully\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(userGroupsCmd)
	userGroupsCmd.AddCommand(ugListCmd)
	userGroupsCmd.AddCommand(ugGetCmd)
	userGroupsCmd.AddCommand(ugCreateCmd)
	userGroupsCmd.AddCommand(ugUpdateCmd)
	userGroupsCmd.AddCommand(ugDeleteCmd)
	userGroupsCmd.AddCommand(ugEnableCmd)
	userGroupsCmd.AddCommand(ugDisableCmd)

	ugListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	ugGetCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	ugCreateCmd.Flags().StringP("name", "n", "", "Name of the user group (required)")
	ugCreateCmd.Flags().BoolP("enabled", "e", true, "Enable the group immediately (default: true)")
	ugCreateCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	ugCreateCmd.MarkFlagRequired("name")

	ugUpdateCmd.Flags().StringP("name", "n", "", "New name for the user group")
	ugUpdateCmd.Flags().BoolP("enabled", "e", false, "Enable or disable the group")
	ugUpdateCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	ugDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}
