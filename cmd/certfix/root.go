package certfix

import (
	"fmt"
	"os"

	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "certfix",
	Short: "Certfix CLI - Manage your certificates and application configurations",
	Long: `Certfix CLI is a command-line interface tool for managing certificates,
application configurations, and infrastructure operations.

Similar to AWS CLI or Azure CLI, it provides authenticated access to your
Certfix services after login and configuration.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger.InitLogger(verbose)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.certfix.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func initConfig() {
	config.InitConfig(cfgFile)
}
