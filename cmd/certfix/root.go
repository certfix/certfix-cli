package certfix

import (
	"os"

	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "certfix",
	Short: "Certfix CLI - Manage your certificates and application configurations",
	Long: `Certfix CLI is a command-line interface tool for managing certificates,
application configurations, and infrastructure operations.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger.InitLogger(verbose)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func initConfig() {
	config.InitConfig("")
}
