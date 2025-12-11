package certfix

import (
	"fmt"
	"os"

	"github.com/certfix/certfix-cli/internal/auth"
	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/client"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/certfix/certfix-cli/pkg/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var applyCmd = &cobra.Command{
	Use:   "apply <config-file.yml>",
	Short: "Apply configuration from YAML file",
	Long: `Apply a complete CertFix configuration from a YAML file.

The configuration file can contain:
- Events
- Policies
- Service Groups
- Services (with API keys and relations)

Resources will be created in order, and if an error occurs, all created 
resources will be rolled back automatically.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		configFile := args[0]

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		skipExisting, _ := cmd.Flags().GetBool("skip-existing")

		// Read YAML file
		fmt.Printf("Reading configuration from: %s\n", configFile)
		data, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		// Parse YAML
		var certfixConfig models.CertfixConfig
		if err := yaml.Unmarshal(data, &certfixConfig); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}

		fmt.Println("Configuration loaded successfully")
		fmt.Printf("  - Events: %d\n", len(certfixConfig.Events))
		fmt.Printf("  - Policies: %d\n", len(certfixConfig.Policies))
		fmt.Printf("  - Service Groups: %d\n", len(certfixConfig.ServiceGroups))
		fmt.Printf("  - Services: %d\n", len(certfixConfig.Services))

		if dryRun {
			fmt.Println("\n=== DRY RUN MODE - No changes will be made ===")

			// Show what would be created
			if len(certfixConfig.Events) > 0 {
				fmt.Println("Events to create:")
				for _, e := range certfixConfig.Events {
					fmt.Printf("  ✓ %s (severity: %s, enabled: %v)\n", e.Name, e.Severity, e.Enabled)
				}
				fmt.Println()
			}

			if len(certfixConfig.Policies) > 0 {
				fmt.Println("Policies to create:")
				for _, p := range certfixConfig.Policies {
					fmt.Printf("  ✓ %s (strategy: %s, enabled: %v)\n", p.Name, p.Strategy, p.Enabled)
					if len(p.CronConfig) > 0 {
						fmt.Printf("      Cron: %v\n", p.CronConfig)
					}
					if len(p.EventConfig) > 0 {
						fmt.Printf("      Event Config: %v\n", p.EventConfig)
					}
				}
				fmt.Println()
			}

			if len(certfixConfig.ServiceGroups) > 0 {
				fmt.Println("Service Groups to create:")
				for _, g := range certfixConfig.ServiceGroups {
					desc := g.Description
					if desc == "" {
						desc = "(no description)"
					}
					fmt.Printf("  ✓ %s - %s (enabled: %v)\n", g.Name, desc, g.Enabled)
				}
				fmt.Println()
			}

			if len(certfixConfig.Services) > 0 {
				fmt.Println("Services to create:")
				for _, s := range certfixConfig.Services {
					fmt.Printf("  ✓ %s (hash: %s)\n", s.Name, s.Hash)
					if s.GroupName != "" {
						fmt.Printf("      Group: %s\n", s.GroupName)
					}
					if s.PolicyName != "" {
						fmt.Printf("      Policy: %s\n", s.PolicyName)
					}
					if s.WebhookURL != "" {
						fmt.Printf("      Webhook: %s\n", s.WebhookURL)
					}
					if len(s.Keys) > 0 {
						fmt.Printf("      Keys: %d\n", len(s.Keys))
						for _, k := range s.Keys {
							fmt.Printf("        - %s (expiration: %d days)\n", k.Name, k.ExpirationDays)
						}
					}
					if len(s.Relations) > 0 {
						fmt.Printf("      Relations: %d\n", len(s.Relations))
						for _, r := range s.Relations {
							fmt.Printf("        - %s (type: %s)\n", r.TargetHash, r.Type)
						}
					}
					fmt.Println()
				}
			}

			total := len(certfixConfig.Events) + len(certfixConfig.Policies) + len(certfixConfig.ServiceGroups) + len(certfixConfig.Services)
			fmt.Printf("Total resources: %d\n", total)
			return nil
		}

		// Get authentication token
		token, err := auth.GetToken()
		if err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		// Create API client
		endpoint := config.GetAPIEndpoint()
		apiClient := client.NewHTTPClient(endpoint)

		// Track created resources for rollback
		var createdResources []models.CreatedResource

		// Defer rollback on error
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Panic occurred: %v", r)
				rollbackResources(apiClient, token, createdResources)
				panic(r)
			}
		}()

		// Apply configuration
		err = applyConfiguration(&certfixConfig, apiClient, token, &createdResources, skipExisting)
		if err != nil {
			log.Errorf("Error during apply: %v", err)
			log.Infof("Rolling back created resources...")
			rollbackResources(apiClient, token, createdResources)
			return err
		}

		log.Infof("✓ Configuration applied successfully!")
		log.Infof("Total resources created: %d", len(createdResources))

		return nil
	},
}

func applyConfiguration(config *models.CertfixConfig, apiClient *client.HTTPClient, token string, createdResources *[]models.CreatedResource, skipExisting bool) error {
	log := logger.GetLogger()

	// 1. Create Events
	log.Infof("\n=== Creating Events ===")
	for i, event := range config.Events {
		log.Infof("[%d/%d] Creating event: %s", i+1, len(config.Events), event.Name)

		if err := createEvent(apiClient, token, event, createdResources, skipExisting); err != nil {
			return fmt.Errorf("failed to create event '%s': %w", event.Name, err)
		}
	}

	// 2. Create Policies
	log.Infof("\n=== Creating Policies ===")
	for i, policy := range config.Policies {
		log.Infof("[%d/%d] Creating policy: %s", i+1, len(config.Policies), policy.Name)

		if err := createPolicy(apiClient, token, policy, createdResources, skipExisting); err != nil {
			return fmt.Errorf("failed to create policy '%s': %w", policy.Name, err)
		}
	}

	// 3. Create Service Groups
	log.Infof("\n=== Creating Service Groups ===")
	for i, group := range config.ServiceGroups {
		log.Infof("[%d/%d] Creating service group: %s", i+1, len(config.ServiceGroups), group.Name)

		if err := createServiceGroup(apiClient, token, group, createdResources, skipExisting); err != nil {
			return fmt.Errorf("failed to create service group '%s': %w", group.Name, err)
		}
	}

	// 4. Create Services (without keys and relations)
	log.Infof("\n=== Creating Services ===")
	for i, service := range config.Services {
		log.Infof("[%d/%d] Creating service: %s (%s)", i+1, len(config.Services), service.Name, service.Hash)

		if err := createService(apiClient, token, service, createdResources, skipExisting); err != nil {
			return fmt.Errorf("failed to create service '%s': %w", service.Hash, err)
		}
	}

	// 5. Create Service Keys
	log.Infof("\n=== Creating Service Keys ===")
	for _, service := range config.Services {
		if len(service.Keys) > 0 {
			log.Infof("Creating %d keys for service: %s", len(service.Keys), service.Hash)

			for i, key := range service.Keys {
				log.Infof("  [%d/%d] Creating key: %s", i+1, len(service.Keys), key.Name)

				if err := createServiceKey(apiClient, token, service.Hash, key, createdResources); err != nil {
					return fmt.Errorf("failed to create key '%s' for service '%s': %w", key.Name, service.Hash, err)
				}
			}
		}
	}

	// 6. Create Service Relations
	log.Infof("\n=== Creating Service Relations ===")
	for _, service := range config.Services {
		if len(service.Relations) > 0 {
			log.Infof("Creating %d relations for service: %s", len(service.Relations), service.Hash)

			for i, relation := range service.Relations {
				log.Infof("  [%d/%d] Creating relation: %s -> %s", i+1, len(service.Relations), service.Hash, relation.TargetHash)

				if err := createServiceRelation(apiClient, token, service.Hash, relation, createdResources); err != nil {
					return fmt.Errorf("failed to create relation from '%s' to '%s': %w", service.Hash, relation.TargetHash, err)
				}
			}
		}
	}

	return nil
}

func createEvent(apiClient *client.HTTPClient, token string, event models.EventConfig, createdResources *[]models.CreatedResource, skipExisting bool) error {
	log := logger.GetLogger()

	// Note: Skip existence check for now - events API doesn't support hash-based lookup

	payload := map[string]interface{}{
		"name":     event.Name,
		"severity": event.Severity,
		"enabled":  event.Enabled,
	}

	_, err := apiClient.PostWithAuth("/events", payload, token)
	if err != nil {
		return err
	}

	*createdResources = append(*createdResources, models.CreatedResource{
		Type: "event",
		Hash: event.Name,
	})

	log.Infof("  ✓ Created successfully")
	return nil
}

func createPolicy(apiClient *client.HTTPClient, token string, policy models.PolicyConfig, createdResources *[]models.CreatedResource, skipExisting bool) error {
	log := logger.GetLogger()

	// Check if exists (skip for now, will check by list)

	payload := map[string]interface{}{
		"name":     policy.Name,
		"strategy": policy.Strategy,
		"enabled":  policy.Enabled,
	}

	// Add optional cron config
	if len(policy.CronConfig) > 0 {
		payload["cron_config"] = policy.CronConfig
	}

	// Add optional event config
	if len(policy.EventConfig) > 0 {
		payload["event_config"] = policy.EventConfig
	}

	_, err := apiClient.PostWithAuth("/politicas", payload, token)
	if err != nil {
		return err
	}

	*createdResources = append(*createdResources, models.CreatedResource{
		Type: "policy",
		Hash: policy.Name,
	})

	log.Infof("  ✓ Created successfully")
	return nil
}

func createServiceGroup(apiClient *client.HTTPClient, token string, group models.ServiceGroupConfig, createdResources *[]models.CreatedResource, skipExisting bool) error {
	log := logger.GetLogger()

	// Check if exists (skip for now, will check by list)

	payload := map[string]interface{}{
		"name":        group.Name,
		"description": group.Description,
		"enabled":     group.Enabled,
	}

	_, err := apiClient.PostWithAuth("/service-groups", payload, token)
	if err != nil {
		return err
	}

	*createdResources = append(*createdResources, models.CreatedResource{
		Type: "service_group",
		Hash: group.Name,
	})

	log.Infof("  ✓ Created successfully")
	return nil
}

func createService(apiClient *client.HTTPClient, token string, service models.ServiceConfig, createdResources *[]models.CreatedResource, skipExisting bool) error {
	log := logger.GetLogger()

	// Check if exists
	_, err := apiClient.GetWithAuth(fmt.Sprintf("/services/%s", service.Hash), token)
	if err == nil {
		if skipExisting {
			log.Infof("  ⊙ Service already exists, skipping")
			return nil
		}
		return fmt.Errorf("service already exists")
	}

	payload := map[string]interface{}{
		"service_hash": service.Hash,
		"service_name": service.Name,
		"active":       service.Active,
	}

	if service.WebhookURL != "" {
		payload["webhook_url"] = service.WebhookURL
	}

	// Look up service group ID by name
	if service.GroupName != "" {
		response, err := apiClient.GetWithAuth(fmt.Sprintf("/service-groups/name/%s", service.GroupName), token)
		if err != nil {
			return fmt.Errorf("failed to find service group '%s': %w", service.GroupName, err)
		}
		if groupID, ok := response["service_group_id"].(string); ok {
			payload["service_group_id"] = groupID
		}
	}

	// Look up policy ID by name
	if service.PolicyName != "" {
		response, err := apiClient.GetWithAuth("/politicas", token)
		if err != nil {
			return fmt.Errorf("failed to get políticas: %w", err)
		}
		// Check if response is an array
		if isArray, ok := response["_is_array"].(bool); ok && isArray {
			if arrayData, ok := response["_array_data"].([]interface{}); ok {
				for _, item := range arrayData {
					if p, ok := item.(map[string]interface{}); ok {
						if pName, ok := p["name"].(string); ok && pName == service.PolicyName {
							if pID, ok := p["politica_id"].(string); ok {
								payload["politica_id"] = pID
								break
							}
						}
					}
				}
			}
		}
	}

	_, err = apiClient.PostWithAuth("/services", payload, token)
	if err != nil {
		return err
	}

	*createdResources = append(*createdResources, models.CreatedResource{
		Type: "service",
		Hash: service.Hash,
	})

	log.Infof("  ✓ Created successfully")
	return nil
}

func createServiceKey(apiClient *client.HTTPClient, token string, serviceHash string, key models.ServiceKeyConfig, createdResources *[]models.CreatedResource) error {
	log := logger.GetLogger()

	payload := map[string]interface{}{
		"key_name": key.Name,
		"enabled":  key.Enabled,
	}

	if key.ExpirationDays > 0 {
		payload["expiration_days"] = key.ExpirationDays
	}

	response, err := apiClient.PostWithAuth(fmt.Sprintf("/services/%s/keys", serviceHash), payload, token)
	if err != nil {
		return err
	}

	keyID := ""
	if id, ok := response["key_id"].(string); ok {
		keyID = id
	}

	*createdResources = append(*createdResources, models.CreatedResource{
		Type: "key",
		Hash: serviceHash,
		ID:   keyID,
	})

	log.Infof("    ✓ Key created")
	return nil
}

func createServiceRelation(apiClient *client.HTTPClient, token string, sourceHash string, relation models.ServiceRelationConfig, createdResources *[]models.CreatedResource) error {
	log := logger.GetLogger()

	payload := map[string]interface{}{
		"related_service_hash": relation.TargetHash,
	}

	if relation.Type != "" {
		payload["relation_type"] = relation.Type
	}

	_, err := apiClient.PostWithAuth(fmt.Sprintf("/services/%s/matriz", sourceHash), payload, token)
	if err != nil {
		return err
	}

	*createdResources = append(*createdResources, models.CreatedResource{
		Type: "relation",
		Hash: sourceHash,
		ID:   relation.TargetHash,
	})

	log.Infof("    ✓ Relation created")
	return nil
}

func rollbackResources(apiClient *client.HTTPClient, token string, resources []models.CreatedResource) {
	log := logger.GetLogger()

	if len(resources) == 0 {
		return
	}

	log.Infof("\n=== Rolling Back Resources ===")
	log.Infof("Deleting %d resources in reverse order...", len(resources))

	// Delete in reverse order
	for i := len(resources) - 1; i >= 0; i-- {
		resource := resources[i]

		switch resource.Type {
		case "relation":
			log.Infof("  Deleting relation: %s -> %s", resource.Hash, resource.ID)
			_, err := apiClient.DeleteWithAuth(fmt.Sprintf("/services/%s/matriz/%s", resource.Hash, resource.ID), token)
			if err != nil {
				log.Warnf("  ⚠ Failed to delete relation: %v", err)
			} else {
				log.Infof("  ✓ Deleted")
			}

		case "key":
			log.Infof("  Deleting key: %s (service: %s)", resource.ID, resource.Hash)
			_, err := apiClient.DeleteWithAuth(fmt.Sprintf("/services/%s/keys/%s", resource.Hash, resource.ID), token)
			if err != nil {
				log.Warnf("  ⚠ Failed to delete key: %v", err)
			} else {
				log.Infof("  ✓ Deleted")
			}

		case "service":
			log.Infof("  Deleting service: %s", resource.Hash)
			_, err := apiClient.DeleteWithAuth(fmt.Sprintf("/services/%s", resource.Hash), token)
			if err != nil {
				log.Warnf("  ⚠ Failed to delete service: %v", err)
			} else {
				log.Infof("  ✓ Deleted")
			}

		case "service_group":
			log.Infof("  Deleting service group: %s", resource.Hash)
			_, err := apiClient.DeleteWithAuth(fmt.Sprintf("/service-groups/%s", resource.Hash), token)
			if err != nil {
				log.Warnf("  ⚠ Failed to delete service group: %v", err)
			} else {
				log.Infof("  ✓ Deleted")
			}

		case "politica":
			log.Infof("  Deleting política: %s", resource.Hash)
			_, err := apiClient.DeleteWithAuth(fmt.Sprintf("/policy/%s", resource.Hash), token)
			if err != nil {
				log.Warnf("  ⚠ Failed to delete política: %v", err)
			} else {
				log.Infof("  ✓ Deleted")
			}

		case "evento":
			log.Infof("  Deleting evento: %s", resource.Hash)
			_, err := apiClient.DeleteWithAuth(fmt.Sprintf("/eventos/%s", resource.Hash), token)
			if err != nil {
				log.Warnf("  ⚠ Failed to delete evento: %v", err)
			} else {
				log.Infof("  ✓ Deleted")
			}
		}
	}

	log.Infof("Rollback completed")
}

func init() {
	rootCmd.AddCommand(applyCmd)

	applyCmd.Flags().Bool("dry-run", false, "Show what would be created without making changes")
	applyCmd.Flags().Bool("skip-existing", false, "Skip resources that already exist instead of failing")
}
