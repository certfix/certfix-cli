package certfix

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/certfix/certfix-cli/internal/config"
	"github.com/certfix/certfix-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Certfix CLI settings",
	Long: `Interactive configuration wizard for Certfix CLI.
Set up your API endpoint URL and other essential settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.GetLogger()
		
		// Get flags
		apiURL, _ := cmd.Flags().GetString("api-url")
		timeout, _ := cmd.Flags().GetInt("timeout")
		retryAttempts, _ := cmd.Flags().GetInt("retry-attempts")

		// Check if any flags were provided
		hasFlags := cmd.Flags().Changed("api-url") || 
					cmd.Flags().Changed("timeout") || 
					cmd.Flags().Changed("retry-attempts")

		// If no flags provided, run interactive configuration
		if !hasFlags {
			return interactiveConfigure()
		}

		// Validate and set API URL if provided
		if cmd.Flags().Changed("api-url") {
			if err := validateURL(apiURL); err != nil {
				log.WithError(err).Error("Invalid API URL")
				return fmt.Errorf("invalid API URL: %w", err)
			}

			if err := config.Set("endpoint", apiURL); err != nil {
				log.WithError(err).Error("Failed to set API URL")
				return fmt.Errorf("failed to set API URL: %w", err)
			}

			log.Infof("API URL set to: %s", apiURL)
			fmt.Printf("✓ API URL configured: %s\n", apiURL)
		}

		// Set timeout if provided
		if cmd.Flags().Changed("timeout") {
			if timeout <= 0 {
				return fmt.Errorf("timeout must be greater than 0")
			}

			if err := config.Set("timeout", fmt.Sprintf("%d", timeout)); err != nil {
				log.WithError(err).Error("Failed to set timeout")
				return fmt.Errorf("failed to set timeout: %w", err)
			}

			log.Infof("Timeout set to: %d seconds", timeout)
			fmt.Printf("✓ Timeout configured: %d seconds\n", timeout)
		}

		// Set retry attempts if provided
		if cmd.Flags().Changed("retry-attempts") {
			if retryAttempts < 0 {
				return fmt.Errorf("retry attempts must be 0 or greater")
			}

			if err := config.Set("retry_attempts", fmt.Sprintf("%d", retryAttempts)); err != nil {
				log.WithError(err).Error("Failed to set retry attempts")
				return fmt.Errorf("failed to set retry attempts: %w", err)
			}

			log.Infof("Retry attempts set to: %d", retryAttempts)
			fmt.Printf("✓ Retry attempts configured: %d\n", retryAttempts)
		}

		fmt.Println("\nConfiguration saved successfully!")
		return nil
	},
}

// validateURL validates that the provided URL is well-formed
func validateURL(apiURL string) error {
	// Ensure URL has a scheme
	if !strings.HasPrefix(apiURL, "http://") && !strings.HasPrefix(apiURL, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	return nil
}

// interactiveConfigure runs an interactive configuration wizard
func interactiveConfigure() error {
	log := logger.GetLogger()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Certfix CLI Configuration")
	fmt.Println("=========================")
	fmt.Println()

	// Get current configuration
	configs, _ := config.List()
	
	// Configure API URL
	currentEndpoint := "https://api.certfix.io"
	if endpoint, ok := configs["endpoint"]; ok {
		currentEndpoint = fmt.Sprintf("%v", endpoint)
	}
	
	fmt.Printf("API URL [%s]: ", currentEndpoint)
	apiURL, _ := reader.ReadString('\n')
	apiURL = strings.TrimSpace(apiURL)
	
	if apiURL == "" {
		apiURL = currentEndpoint
	}
	
	// Validate and set API URL
	if err := validateURL(apiURL); err != nil {
		log.WithError(err).Error("Invalid API URL")
		return fmt.Errorf("invalid API URL: %w", err)
	}
	
	if err := config.Set("endpoint", apiURL); err != nil {
		log.WithError(err).Error("Failed to set API URL")
		return fmt.Errorf("failed to set API URL: %w", err)
	}
	
	fmt.Printf("✓ API URL configured: %s\n", apiURL)

	// Configure timeout
	currentTimeout := 30
	if timeout, ok := configs["timeout"]; ok {
		if t, ok := timeout.(int); ok {
			currentTimeout = t
		}
	}
	
	fmt.Printf("Timeout in seconds [%d]: ", currentTimeout)
	timeoutStr, _ := reader.ReadString('\n')
	timeoutStr = strings.TrimSpace(timeoutStr)
	
	timeout := currentTimeout
	if timeoutStr != "" {
		t, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout value: must be a number")
		}
		if t <= 0 {
			return fmt.Errorf("timeout must be greater than 0")
		}
		timeout = t
	}
	
	if err := config.Set("timeout", fmt.Sprintf("%d", timeout)); err != nil {
		log.WithError(err).Error("Failed to set timeout")
		return fmt.Errorf("failed to set timeout: %w", err)
	}
	
	fmt.Printf("✓ Timeout configured: %d seconds\n", timeout)

	// Configure retry attempts
	currentRetry := 3
	if retry, ok := configs["retry_attempts"]; ok {
		if r, ok := retry.(int); ok {
			currentRetry = r
		}
	}
	
	fmt.Printf("Retry attempts [%d]: ", currentRetry)
	retryStr, _ := reader.ReadString('\n')
	retryStr = strings.TrimSpace(retryStr)
	
	retryAttempts := currentRetry
	if retryStr != "" {
		r, err := strconv.Atoi(retryStr)
		if err != nil {
			return fmt.Errorf("invalid retry attempts value: must be a number")
		}
		if r < 0 {
			return fmt.Errorf("retry attempts must be 0 or greater")
		}
		retryAttempts = r
	}
	
	if err := config.Set("retry_attempts", fmt.Sprintf("%d", retryAttempts)); err != nil {
		log.WithError(err).Error("Failed to set retry attempts")
		return fmt.Errorf("failed to set retry attempts: %w", err)
	}
	
	fmt.Printf("✓ Retry attempts configured: %d\n", retryAttempts)

	fmt.Println("\nConfiguration saved successfully!")
	return nil
}

// showCurrentConfig displays the current configuration
func showCurrentConfig() error {
	log := logger.GetLogger()
	log.Info("Displaying current configuration")

	configs, err := config.List()
	if err != nil {
		return fmt.Errorf("failed to retrieve configuration: %w", err)
	}

	fmt.Println("Current Certfix CLI Configuration:")
	fmt.Println("==================================")
	
	// Display key configurations
	if endpoint, ok := configs["endpoint"]; ok {
		fmt.Printf("API URL:         %v\n", endpoint)
	}
	if timeout, ok := configs["timeout"]; ok {
		fmt.Printf("Timeout:         %v seconds\n", timeout)
	}
	if retryAttempts, ok := configs["retry_attempts"]; ok {
		fmt.Printf("Retry Attempts:  %v\n", retryAttempts)
	}

	fmt.Println("\nTo change settings, use:")
	fmt.Println("  certfix configure --api-url <url>")
	fmt.Println("  certfix configure --timeout <seconds>")
	fmt.Println("  certfix configure --retry-attempts <count>")

	return nil
}

func init() {
	rootCmd.AddCommand(configureCmd)

	configureCmd.Flags().StringP("api-url", "a", "", "API endpoint URL (e.g., https://api.certfix.io)")
	configureCmd.Flags().IntP("timeout", "t", 0, "Request timeout in seconds")
	configureCmd.Flags().IntP("retry-attempts", "r", 0, "Number of retry attempts for failed requests")
}
