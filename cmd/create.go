package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"slices"

	"github.com/DenisBytes/gonstrukt/cmd/types"
	"github.com/spf13/cobra"
)

func CreateCmd() *cobra.Command {
	var config types.ServiceConfig
	var serviceTypeStr, databaseStr, cacheStr, configStr, rateLimiterStr, observabilityStr string

	cmd := &cobra.Command{
		Use:   "create <git_repo_url>",
		Short: "Create a new Go service with specified configuration",
		Long: `Create a new Go service (gateway or auth) with database, caching,
configuration, observability, and rate limiting options.`,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var validationErrors []error

			if err := validateServiceName(args[0]); err != nil {
				if errors.Is(err, ErrServiceNameRequired) || errors.Is(err, ErrInvalidServiceName) {
					validationErrors = append(validationErrors, err)
				} else {
					validationErrors = append(validationErrors, fmt.Errorf("service name validation failed: %w", err))
				}
			} else {
				config.Name = args[0]
			}

			if err := validateAndSetServiceType(&config, serviceTypeStr); err != nil {
				if IsMissingRequiredFieldError(err) || IsInvalidOptionError(err) {
					validationErrors = append(validationErrors, err)
				} else {
					validationErrors = append(validationErrors, fmt.Errorf("service type validation failed: %w", err))
				}
			}

			if err := validateAndSetDatabase(&config, databaseStr); err != nil {
				if IsMissingRequiredFieldError(err) || IsInvalidOptionError(err) {
					validationErrors = append(validationErrors, err)
				} else {
					validationErrors = append(validationErrors, fmt.Errorf("database validation failed: %w", err))
				}
			}

			if err := validateAndSetConfig(&config, configStr); err != nil {
				if IsMissingRequiredFieldError(err) || IsInvalidOptionError(err) {
					validationErrors = append(validationErrors, err)
				} else {
					validationErrors = append(validationErrors, fmt.Errorf("config validation failed: %w", err))
				}
			}

			if err := validateAndSetCache(&config, cacheStr); err != nil {
				if IsMissingRequiredFieldError(err) || IsInvalidOptionError(err) {
					validationErrors = append(validationErrors, err)
				} else {
					validationErrors = append(validationErrors, fmt.Errorf("cache validation failed: %w", err))
				}
			}

			if err := validateAndSetRateLimiter(&config, rateLimiterStr); err != nil {
				if IsMissingRequiredFieldError(err) || IsInvalidOptionError(err) {
					validationErrors = append(validationErrors, err)
				} else {
					validationErrors = append(validationErrors, fmt.Errorf("rate limiter validation failed: %w", err))
				}
			}

			if err := validateAndSetObservability(&config, observabilityStr); err != nil {
				if IsInvalidOptionError(err) {
					validationErrors = append(validationErrors, err)
				} else {
					validationErrors = append(validationErrors, fmt.Errorf("observability validation failed: %w", err))
				}
			}

			if len(validationErrors) > 0 {
				joinedErr := errors.Join(validationErrors...)
				usage := cmd.UsageString()
				return NewCliError(joinedErr, usage)
			}

			// printServiceConfig(config)

			return nil
		},
	}

	cmd.Flags().StringVarP(&serviceTypeStr, "service-type", "s", "", "Service type (gateway, auth) [required]")
	cmd.Flags().StringVarP(&databaseStr, "database", "d", "", "Database type (psql) [required]")
	cmd.Flags().StringVar(&cacheStr, "cache", "", "Cache type (memory, redis, valkey) [required for gateway, optional for auth]")
	cmd.Flags().StringVarP(&configStr, "config", "", "", "Configuration source (yaml, vault) [required]")
	cmd.Flags().StringVarP(&rateLimiterStr, "rate-limiter", "r", "", "Rate limiting algorithm (token-bucket, approximated-sliding-window) [required for gateway, optional for auth]")
	cmd.Flags().StringVarP(&observabilityStr, "observability", "o", "otlp", "Observability type (otlp, none) [optional, defaults to otlp]")

	cmd.MarkFlagRequired("service-type")
	cmd.MarkFlagRequired("database")
	cmd.MarkFlagRequired("config")

	cmd.RegisterFlagCompletionFunc("service-type", serviceTypeCompletion)
	cmd.RegisterFlagCompletionFunc("database", databaseCompletion)
	cmd.RegisterFlagCompletionFunc("cache", cacheCompletion)
	cmd.RegisterFlagCompletionFunc("config", configCompletion)
	cmd.RegisterFlagCompletionFunc("rate-limiter", rateLimiterCompletion)
	cmd.RegisterFlagCompletionFunc("observability", observabilityCompletion)

	return cmd
}

func validateServiceName(name string) error {
	if name == "" {
		return ErrServiceNameRequired
	}

	pattern := `^github\.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$`
	matched, err := regexp.MatchString(pattern, name)
	if err != nil {
		return NewValidationError("service-name", name, "failed to validate format", err)
	}
	if !matched {
		return NewInvalidFormatError("service-name", name, "github.com/<username/org>/<project_name>")
	}
	return nil
}

func validateAndSetServiceType(config *types.ServiceConfig, serviceTypeStr string) error {
	if serviceTypeStr == "" {
		return NewMissingRequiredFieldError("service-type", "", "")
	}

	validTypes := types.ValidServiceTypes()
	if slices.Contains(validTypes, serviceTypeStr) {
		config.ServiceType = types.ServiceType(serviceTypeStr)
		return nil
	}

	return NewInvalidOptionError("service-type", serviceTypeStr, validTypes)
}

func validateAndSetDatabase(config *types.ServiceConfig, databaseStr string) error {
	if databaseStr == "" {
		return NewMissingRequiredFieldError("database", "", "")
	}

	validDatabases := types.ValidDatabaseTypes()
	if slices.Contains(validDatabases, databaseStr) {
		config.Database = types.DatabaseType(databaseStr)
		return nil
	}

	return NewInvalidOptionError("database", databaseStr, validDatabases)
}

func validateAndSetCache(config *types.ServiceConfig, cacheStr string) error {
	if config.ServiceType == types.ServiceTypeGateway && cacheStr == "" {
		return NewMissingRequiredFieldError("cache", string(config.ServiceType), "caching is mandatory for performance")
	}

	if cacheStr != "" {
		validCaches := types.ValidCacheTypes()
		if slices.Contains(validCaches, cacheStr) {
			cacheType := types.CacheType(cacheStr)
			config.Cache = &cacheType
			return nil
		}
		return NewInvalidOptionError("cache", cacheStr, validCaches)
	}

	return nil
}

func validateAndSetConfig(config *types.ServiceConfig, configStr string) error {
	if configStr == "" {
		return NewMissingRequiredFieldError("config", "", "configuration source must be specified")
	}

	validConfigs := types.ValidConfigTypes()
	if slices.Contains(validConfigs, configStr) {
		config.Config = types.ConfigType(configStr)
		return nil
	}

	return NewInvalidOptionError("config", configStr, validConfigs)
}

func validateAndSetRateLimiter(config *types.ServiceConfig, rateLimiterStr string) error {
	if config.ServiceType == types.ServiceTypeGateway && rateLimiterStr == "" {
		return NewMissingRequiredFieldError("rate-limiter", string(config.ServiceType), "rate limiting is essential for gateway traffic control")
	}

	if rateLimiterStr != "" {
		validRateLimiters := types.ValidRateLimiterTypes()
		if slices.Contains(validRateLimiters, rateLimiterStr) {
			rateLimiter := types.RateLimiterType(rateLimiterStr)
			config.RateLimiter = &rateLimiter
			return nil
		}
		return NewInvalidOptionError("rate-limiter", rateLimiterStr, validRateLimiters)
	}

	return nil
}

func validateAndSetObservability(config *types.ServiceConfig, observabilityStr string) error {
	if observabilityStr == "" {
		observabilityStr = "otlp"
	}

	validObservability := types.ValidObservabilityTypes()
	if slices.Contains(validObservability, observabilityStr) {
		config.Observability = types.ObservabilityType(observabilityStr)
		return nil
	}

	return NewInvalidOptionError("observability", observabilityStr, validObservability)
}

// func printServiceConfig(config types.ServiceConfig) {
// 	fmt.Printf("Service Configuration:\n")
// 	fmt.Printf("  Name: %s\n", config.Name)
// 	fmt.Printf("  Service Type: %s\n", config.ServiceType)
// 	fmt.Printf("  Database: %s\n", config.Database)

// 	if config.Cache != nil {
// 		fmt.Printf("  Cache: %s\n", *config.Cache)
// 	} else {
// 		fmt.Printf("  Cache: none\n")
// 	}

// 	fmt.Printf("  Config Source: %s\n", config.Config)

// 	if config.RateLimiter != nil {
// 		fmt.Printf("  Rate Limiter: %s\n", *config.RateLimiter)
// 	} else {
// 		fmt.Printf("  Rate Limiter: none\n")
// 	}

// 	fmt.Printf("  Observability: %s\n", config.Observability)
// }

func serviceTypeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return types.ValidServiceTypes(), cobra.ShellCompDirectiveNoFileComp
}

func databaseCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return types.ValidDatabaseTypes(), cobra.ShellCompDirectiveNoFileComp
}

func cacheCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return types.ValidCacheTypes(), cobra.ShellCompDirectiveNoFileComp
}

func configCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return types.ValidConfigTypes(), cobra.ShellCompDirectiveNoFileComp
}

func rateLimiterCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return types.ValidRateLimiterTypes(), cobra.ShellCompDirectiveNoFileComp
}

func observabilityCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return types.ValidObservabilityTypes(), cobra.ShellCompDirectiveNoFileComp
}
