package certfix

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/client"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:               "users",
	Aliases:           []string{"user"},
	Short:             "Manage users",
	Long:              `Manage users including listing, creating, updating, deleting, and managing super user status.`,
	PersistentPreRunE: requireSuperuser,
}

var usersListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth("/users", token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to list users: %w", err)
		}

		var users []map[string]interface{}
		if response["_is_array"] != nil {
			if arr, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arr {
					if u, ok := item.(map[string]interface{}); ok {
						users = append(users, u)
					}
				}
			}
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(users, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(users) == 0 {
			fmt.Println("No users found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "USER ID\tEMAIL\tSUPER\tENABLED")
		fmt.Fprintln(w, "-------\t-----\t-----\t-------")

		for _, u := range users {
			id := fmt.Sprintf("%v", u["user_id"])
			email := fmt.Sprintf("%v", u["email"])
			if len(email) > 35 {
				email = email[:32] + "..."
			}
			isSuper := "No"
			if s, ok := u["is_super_user"].(bool); ok && s {
				isSuper = "Yes"
			}
			enabled := "No"
			if e, ok := u["enabled"].(bool); ok && e {
				enabled = "Yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, email, isSuper, enabled)
		}
		w.Flush()

		return nil
	},
}

var usersGetCmd = &cobra.Command{
	Use:   "get <user-id>",
	Short: "Get details of a specific user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		userID := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.GetWithAuth(fmt.Sprintf("/users/%s", userID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to get user: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("User ID:    %v\n", response["user_id"])
		fmt.Printf("Email:      %v\n", response["email"])
		isSuper := "No"
		if s, ok := response["is_super_user"].(bool); ok && s {
			isSuper = "Yes"
		}
		fmt.Printf("Super User: %s\n", isSuper)
		enabled := "No"
		if e, ok := response["enabled"].(bool); ok && e {
			enabled = "Yes"
		}
		fmt.Printf("Enabled:    %s\n", enabled)

		return nil
	},
}

var usersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		outputFormat, _ := cmd.Flags().GetString("output")

		if name == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("name is required (use --name)")
		}
		if email == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("email is required (use --email)")
		}
		if password == "" {
			cmd.SilenceUsage = true
			return fmt.Errorf("password is required (use --password)")
		}

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		payload := map[string]interface{}{
			"name":     name,
			"email":    email,
			"password": password,
		}

		response, err := apiClient.PostWithAuth("/users", payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to create user: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ User created successfully\n")

		return nil
	},
}

var usersUpdateCmd = &cobra.Command{
	Use:   "update <user-id>",
	Short: "Update an existing user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		userID := args[0]
		name, _ := cmd.Flags().GetString("name")
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		outputFormat, _ := cmd.Flags().GetString("output")

		payload := make(map[string]interface{})
		if name != "" {
			payload["name"] = name
		}
		if email != "" {
			payload["email"] = email
		}
		if password != "" {
			payload["password"] = password
		}

		if len(payload) == 0 {
			cmd.SilenceUsage = true
			return fmt.Errorf("no fields to update (use --name, --email, or --password)")
		}

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		response, err := apiClient.PutWithAuth(fmt.Sprintf("/users/%s", userID), payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to update user: %w", err)
		}

		if outputFormat == "json" {
			data, _ := json.MarshalIndent(response, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ User updated successfully\n")
		fmt.Printf("ID:    %v\n", response["id"])
		fmt.Printf("Name:  %v\n", response["name"])
		fmt.Printf("Email: %v\n", response["email"])

		return nil
	},
}

var usersDeleteCmd = &cobra.Command{
	Use:     "delete <user-id>",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a user",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		userID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Are you sure you want to delete user %s? (y/N): ", userID)
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

		_, err = apiClient.DeleteWithAuth(fmt.Sprintf("/users/%s", userID), token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to delete user: %w", err)
		}

		fmt.Printf("✓ User deleted successfully\n")
		return nil
	},
}

var usersSetSuperCmd = &cobra.Command{
	Use:   "set-super <email>",
	Short: "Grant super user privileges to a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		email := args[0]

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		payload := map[string]interface{}{
			"email": email,
		}

		_, err = apiClient.PostWithAuth("/users/super", payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to set super user: %w", err)
		}

		fmt.Printf("✓ Super user privileges granted to %s\n", email)
		return nil
	},
}

var usersRevokeSuperCmd = &cobra.Command{
	Use:   "revoke-super <email>",
	Short: "Revoke super user privileges from a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		email := args[0]

		token, err := auth.GetToken()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		payload := map[string]interface{}{
			"email": email,
		}

		_, err = apiClient.PostWithAuth("/users/super/not", payload, token)
		if err != nil {
			cmd.SilenceUsage = true
			return fmt.Errorf("failed to revoke super user: %w", err)
		}

		fmt.Printf("✓ Super user privileges revoked from %s\n", email)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(usersListCmd)
	usersCmd.AddCommand(usersGetCmd)
	usersCmd.AddCommand(usersCreateCmd)
	usersCmd.AddCommand(usersUpdateCmd)
	usersCmd.AddCommand(usersDeleteCmd)
	usersCmd.AddCommand(usersSetSuperCmd)
	usersCmd.AddCommand(usersRevokeSuperCmd)

	usersListCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	usersGetCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	usersCreateCmd.Flags().StringP("name", "n", "", "Full name of the user (required)")
	usersCreateCmd.Flags().StringP("email", "e", "", "Email address of the user (required)")
	usersCreateCmd.Flags().StringP("password", "p", "", "Password for the user (required)")
	usersCreateCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")
	usersCreateCmd.MarkFlagRequired("name")
	usersCreateCmd.MarkFlagRequired("email")
	usersCreateCmd.MarkFlagRequired("password")

	usersUpdateCmd.Flags().StringP("name", "n", "", "New name for the user")
	usersUpdateCmd.Flags().StringP("email", "e", "", "New email for the user")
	usersUpdateCmd.Flags().StringP("password", "p", "", "New password for the user")
	usersUpdateCmd.Flags().StringP("output", "o", "table", "Output format (table, json)")

	usersDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}
