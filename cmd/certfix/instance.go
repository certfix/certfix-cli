package certfix

import (
	"fmt"

	"github.com/certfix/certfix-cli/internal/api"
	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Manage Certfix instances",
	Long:  `Create, configure, and manage Certfix instances.`,
}

var instanceCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new instance",
	Long:  `Create a new Certfix instance with the specified name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		instanceType, _ := cmd.Flags().GetString("type")
		region, _ := cmd.Flags().GetString("region")

		log := logger.GetLogger()
		log.Infof("Creating instance: %s", name)

		// Check authentication
		if !auth.IsAuthenticated() {
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		// Create instance
		client := api.NewClient()
		instance, err := client.CreateInstance(name, instanceType, region)
		if err != nil {
			log.WithError(err).Error("Failed to create instance")
			return fmt.Errorf("failed to create instance: %w", err)
		}

		fmt.Printf("Instance '%s' created successfully\n", instance.Name)
		fmt.Printf("ID: %s\n", instance.ID)
		return nil
	},
}

var instanceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all instances",
	Long:  `List all Certfix instances in your account.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		log.Info("Listing instances")

		// Check authentication
		if !auth.IsAuthenticated() {
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		client := api.NewClient()
		instances, err := client.ListInstances()
		if err != nil {
			log.WithError(err).Error("Failed to list instances")
			return fmt.Errorf("failed to list instances: %w", err)
		}

		if len(instances) == 0 {
			fmt.Println("No instances found")
			return nil
		}

		fmt.Println("Instances:")
		for _, instance := range instances {
			fmt.Printf("  - %s (ID: %s, Status: %s)\n", instance.Name, instance.ID, instance.Status)
		}
		return nil
	},
}

var instanceDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete an instance",
	Long:  `Delete a Certfix instance by ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		log := logger.GetLogger()
		log.Infof("Deleting instance: %s", id)

		// Check authentication
		if !auth.IsAuthenticated() {
			return fmt.Errorf("not authenticated, please run 'certfix login' first")
		}

		client := api.NewClient()
		if err := client.DeleteInstance(id); err != nil {
			log.WithError(err).Error("Failed to delete instance")
			return fmt.Errorf("failed to delete instance: %w", err)
		}

		fmt.Printf("Instance '%s' deleted successfully\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(instanceCmd)
	instanceCmd.AddCommand(instanceCreateCmd)
	instanceCmd.AddCommand(instanceListCmd)
	instanceCmd.AddCommand(instanceDeleteCmd)

	instanceCreateCmd.Flags().StringP("type", "t", "standard", "Instance type")
	instanceCreateCmd.Flags().StringP("region", "r", "us-east-1", "Instance region")
}
