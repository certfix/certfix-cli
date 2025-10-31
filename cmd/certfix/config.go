package certfix

import (
	"fmt"

	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  `Manage Certfix CLI configuration settings and instance configurations.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  `Set a configuration key-value pair in your Certfix configuration.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		log := logger.GetLogger()
		log.Infof("Setting configuration: %s = %s", key, value)

		if err := config.Set(key, value); err != nil {
			log.WithError(err).Error("Failed to set configuration")
			return fmt.Errorf("failed to set configuration: %w", err)
		}

		fmt.Printf("Configuration updated: %s = %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long:  `Retrieve a configuration value from your Certfix configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		log := logger.GetLogger()
		log.Infof("Getting configuration: %s", key)

		value, err := config.Get(key)
		if err != nil {
			log.WithError(err).Error("Failed to get configuration")
			return fmt.Errorf("failed to get configuration: %w", err)
		}

		fmt.Printf("%s = %s\n", key, value)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `List all configuration key-value pairs in your Certfix configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		log.Info("Listing all configurations")

		configs, err := config.List()
		if err != nil {
			log.WithError(err).Error("Failed to list configurations")
			return fmt.Errorf("failed to list configurations: %w", err)
		}

		if len(configs) == 0 {
			fmt.Println("No configurations found")
			return nil
		}

		fmt.Println("Current configurations:")
		for key, value := range configs {
			fmt.Printf("  %s = %v\n", key, value)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
}
