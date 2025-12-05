package config

import (
	"os"

	"slices"

	"gopkg.in/yaml.v2"
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

func LoadConfig(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}

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
