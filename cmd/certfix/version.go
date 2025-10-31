package certfix

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set during build time
	Version = "1.0.0"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Display the current version of Certfix CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Certfix CLI v%s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
