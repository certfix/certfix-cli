package certfix

import (
	"fmt"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/client"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

// requireSuperuser fetches the current user via /me and returns an error if the
// user does not have superuser privileges. It also initialises the logger so that
// commands which define their own PersistentPreRunE do not skip the root-level
// logger initialisation.
func requireSuperuser(cmd *cobra.Command, args []string) error {
	logger.InitLogger(verbose)

	token, err := auth.GetToken()
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	endpoint := config.GetAPIEndpoint()
	apiClient := client.NewHTTPClient(endpoint)

	response, err := apiClient.GetWithAuth("/me", token)
	if err != nil {
		cmd.SilenceUsage = true
		return fmt.Errorf("failed to verify permissions: %w", err)
	}

	isSuper, _ := response["is_super_user"].(bool)
	if !isSuper {
		cmd.SilenceUsage = true
		return fmt.Errorf("permission denied: this command requires superuser privileges\n  Run 'certfix whoami' to check your current role or contact your administrator")
	}

	return nil
}
