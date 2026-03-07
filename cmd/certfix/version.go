package certfix

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// Version is set during build time
	Version = "0.0.1"
	// BuildDate is set during build time via ldflags
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Display the current version of Certfix CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Certfix CLI v%s (built %s)\n", strings.TrimPrefix(Version, "v"), BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
