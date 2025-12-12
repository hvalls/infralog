package config

import (
	"os"
	"strconv"
	"strings"

	"slices"

	"gopkg.in/yaml.v2"
)

// Environment variable names
const (
	envPrefix = "INFRALOG_"

	// Webhook target
	envWebhookURL                 = "INFRALOG_TARGET_WEBHOOK_URL"
	envWebhookMethod              = "INFRALOG_TARGET_WEBHOOK_METHOD"
	envWebhookRetryMaxAttempts    = "INFRALOG_TARGET_WEBHOOK_RETRY_MAX_ATTEMPTS"
	envWebhookRetryInitialDelayMS = "INFRALOG_TARGET_WEBHOOK_RETRY_INITIAL_DELAY_MS"
	envWebhookRetryMaxDelayMS     = "INFRALOG_TARGET_WEBHOOK_RETRY_MAX_DELAY_MS"
	envWebhookRetryRetryOnStatus  = "INFRALOG_TARGET_WEBHOOK_RETRY_RETRY_ON_STATUS"

	// Slack target
	envSlackWebhookURL = "INFRALOG_TARGET_SLACK_WEBHOOK_URL"
	envSlackChannel    = "INFRALOG_TARGET_SLACK_CHANNEL"
	envSlackUsername   = "INFRALOG_TARGET_SLACK_USERNAME"
	envSlackIconEmoji  = "INFRALOG_TARGET_SLACK_ICON_EMOJI"

	// Filters
	envFilterResourceTypes = "INFRALOG_FILTER_RESOURCE_TYPES"
	envFilterOutputs       = "INFRALOG_FILTER_OUTPUTS"
)

type Config struct {
	Target Target `yaml:"target"`
	Filter Filter `yaml:"filter"`
}

type Target struct {
	Webhook WebhookConfig `yaml:"webhook"`
	Slack   SlackConfig   `yaml:"slack"`
}

type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`    // Optional: override default channel
	Username   string `yaml:"username"`   // Optional: override bot username
	IconEmoji  string `yaml:"icon_emoji"` // Optional: override bot icon
}

type WebhookConfig struct {
	URL    string      `yaml:"url"`
	Method string      `yaml:"method"`
	Retry  RetryConfig `yaml:"retry"`
}

type RetryConfig struct {
	MaxAttempts  int   `yaml:"max_attempts"`
	InitialDelay int   `yaml:"initial_delay_ms"`
	MaxDelay     int   `yaml:"max_delay_ms"`
	StatusCodes  []int `yaml:"retry_on_status"`
}

// WithDefaults returns a RetryConfig with default values applied.
func (r RetryConfig) WithDefaults() RetryConfig {
	if r.MaxAttempts == 0 {
		r.MaxAttempts = 3
	}
	if r.InitialDelay == 0 {
		r.InitialDelay = 1000 // 1 second
	}
	if r.MaxDelay == 0 {
		r.MaxDelay = 30000 // 30 seconds
	}
	if r.StatusCodes == nil {
		r.StatusCodes = []int{500, 502, 503, 504}
	}
	return r
}

type Filter struct {
	ResourceTypes []string `yaml:"resource_types"`
	Outputs       []string `yaml:"outputs"`
}

// setStringFromEnv sets target to the env var value if the env var is set and non-empty.
func setStringFromEnv(target *string, envKey string) {
	if val := os.Getenv(envKey); val != "" {
		*target = val
	}
}

// setIntFromEnv sets target to the env var value (parsed as int) if the env var is set and valid.
func setIntFromEnv(target *int, envKey string) {
	if val := os.Getenv(envKey); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			*target = intVal
		}
	}
}

// setStringSliceFromEnv sets target to the env var value (comma-separated) if the env var is set.
func setStringSliceFromEnv(target *[]string, envKey string) {
	if val := os.Getenv(envKey); val != "" {
		parts := strings.Split(val, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			*target = result
		}
	}
}

// setIntSliceFromEnv sets target to the env var value (comma-separated ints) if the env var is set.
func setIntSliceFromEnv(target *[]int, envKey string) {
	if val := os.Getenv(envKey); val != "" {
		parts := strings.Split(val, ",")
		result := make([]int, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				if intVal, err := strconv.Atoi(trimmed); err == nil {
					result = append(result, intVal)
				}
			}
		}
		if len(result) > 0 {
			*target = result
		}
	}
}

// loadConfigFromEnv loads configuration from environment variables with INFRALOG_ prefix.
func loadConfigFromEnv(cfg *Config) {
	// Webhook target
	setStringFromEnv(&cfg.Target.Webhook.URL, envWebhookURL)
	setStringFromEnv(&cfg.Target.Webhook.Method, envWebhookMethod)

	// Webhook retry configuration
	setIntFromEnv(&cfg.Target.Webhook.Retry.MaxAttempts, envWebhookRetryMaxAttempts)
	setIntFromEnv(&cfg.Target.Webhook.Retry.InitialDelay, envWebhookRetryInitialDelayMS)
	setIntFromEnv(&cfg.Target.Webhook.Retry.MaxDelay, envWebhookRetryMaxDelayMS)
	setIntSliceFromEnv(&cfg.Target.Webhook.Retry.StatusCodes, envWebhookRetryRetryOnStatus)

	// Slack target
	setStringFromEnv(&cfg.Target.Slack.WebhookURL, envSlackWebhookURL)
	setStringFromEnv(&cfg.Target.Slack.Channel, envSlackChannel)
	setStringFromEnv(&cfg.Target.Slack.Username, envSlackUsername)
	setStringFromEnv(&cfg.Target.Slack.IconEmoji, envSlackIconEmoji)

	// Filters
	setStringSliceFromEnv(&cfg.Filter.ResourceTypes, envFilterResourceTypes)
	setStringSliceFromEnv(&cfg.Filter.Outputs, envFilterOutputs)
}

func LoadConfig(filename string) (*Config, error) {
	var config Config

	// Load from file if provided
	if filename != "" {
		file, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(file, &config); err != nil {
			return nil, err
		}
	}

	// Overlay environment variables (they take precedence over file config)
	loadConfigFromEnv(&config)

	return &config, nil
}

func (f *Filter) MatchesResourceType(resourceType string) bool {
	if f.ResourceTypes == nil {
		return true
	}
	if len(f.ResourceTypes) == 0 {
		return false
	}
	return slices.Contains(f.ResourceTypes, resourceType)
}

func (f *Filter) MatchesOutput(output string) bool {
	if f.Outputs == nil {
		return true
	}
	if len(f.Outputs) == 0 {
		return false
	}
	return slices.Contains(f.Outputs, output)
}
