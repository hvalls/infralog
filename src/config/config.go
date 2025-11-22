package config

import (
	"os"

	"slices"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Polling struct {
		Interval int `yaml:"interval"`
	} `yaml:"polling"`
	TFState     TFState       `yaml:"tfstate"`
	Target      Target        `yaml:"target"`
	Filter      Filter        `yaml:"filter"`
	Persistence Persistence   `yaml:"persistence"`
	Metrics     MetricsConfig `yaml:"metrics"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
}

// WithDefaults returns a MetricsConfig with default values applied.
func (m MetricsConfig) WithDefaults() MetricsConfig {
	if m.Address == "" {
		m.Address = ":8080"
	}
	return m
}

type Target struct {
	Webhook WebhookConfig `yaml:"webhook"`
	Slack   SlackConfig   `yaml:"slack"`
	Stdout  StdoutConfig  `yaml:"stdout"`
}

type StdoutConfig struct {
	Enabled bool   `yaml:"enabled"`
	Format  string `yaml:"format"` // "json" or "text" (default: "text")
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

type Persistence struct {
	StateFile string `yaml:"state_file"`
}

type TFState struct {
	S3    S3Config    `yaml:"s3" json:"s3"`
	Local LocalConfig `yaml:"local" json:"local"`
}

type S3Config struct {
	Bucket string `yaml:"bucket" json:"bucket"`
	Key    string `yaml:"key" json:"key"`
	Region string `yaml:"region" json:"region"`
}

type LocalConfig struct {
	Path string `yaml:"path" json:"path"`
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
