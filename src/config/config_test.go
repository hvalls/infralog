package config

import (
	"os"
	"testing"
)

func TestFilter_MatchesResourceType(t *testing.T) {
	tests := []struct {
		name         string
		filter       Filter
		resourceType string
		want         bool
	}{
		{
			name:         "nil resource types should match any resource",
			filter:       Filter{ResourceTypes: nil},
			resourceType: "aws_instance",
			want:         true,
		},
		{
			name:         "empty resource types should match no resource",
			filter:       Filter{ResourceTypes: []string{}},
			resourceType: "aws_instance",
			want:         false,
		},
		{
			name:         "should match when resource type is in the list",
			filter:       Filter{ResourceTypes: []string{"aws_instance", "aws_vpc"}},
			resourceType: "aws_instance",
			want:         true,
		},
		{
			name:         "should not match when resource type is not in the list",
			filter:       Filter{ResourceTypes: []string{"aws_vpc", "aws_subnet"}},
			resourceType: "aws_instance",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.MatchesResourceType(tt.resourceType); got != tt.want {
				t.Errorf("Filter.MatchesResourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_MatchesOutput(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		output string
		want   bool
	}{
		{
			name:   "nil outputs should match any output",
			filter: Filter{Outputs: nil},
			output: "instance_ip",
			want:   true,
		},
		{
			name:   "empty outputs should match no output",
			filter: Filter{Outputs: []string{}},
			output: "instance_ip",
			want:   false,
		},
		{
			name:   "should match when output is in the list",
			filter: Filter{Outputs: []string{"instance_ip", "vpc_id"}},
			output: "instance_ip",
			want:   true,
		},
		{
			name:   "should not match when output is not in the list",
			filter: Filter{Outputs: []string{"vpc_id", "subnet_id"}},
			output: "instance_ip",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.MatchesOutput(tt.output); got != tt.want {
				t.Errorf("Filter.MatchesOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Save original environment and restore after test
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			pair := splitEnv(env)
			if len(pair) == 2 {
				os.Setenv(pair[0], pair[1])
			}
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		want     Config
		wantDesc string
	}{
		{
			name: "webhook configuration from env",
			envVars: map[string]string{
				"INFRALOG_TARGET_WEBHOOK_URL":    "https://example.com/webhook",
				"INFRALOG_TARGET_WEBHOOK_METHOD": "POST",
			},
			want: Config{
				Target: Target{
					Webhook: WebhookConfig{
						URL:    "https://example.com/webhook",
						Method: "POST",
					},
				},
			},
			wantDesc: "should load webhook URL and method from env",
		},
		{
			name: "webhook retry configuration from env",
			envVars: map[string]string{
				"INFRALOG_TARGET_WEBHOOK_RETRY_MAX_ATTEMPTS":      "5",
				"INFRALOG_TARGET_WEBHOOK_RETRY_INITIAL_DELAY_MS":  "2000",
				"INFRALOG_TARGET_WEBHOOK_RETRY_MAX_DELAY_MS":      "60000",
				"INFRALOG_TARGET_WEBHOOK_RETRY_RETRY_ON_STATUS":   "500,502,503",
			},
			want: Config{
				Target: Target{
					Webhook: WebhookConfig{
						Retry: RetryConfig{
							MaxAttempts:  5,
							InitialDelay: 2000,
							MaxDelay:     60000,
							StatusCodes:  []int{500, 502, 503},
						},
					},
				},
			},
			wantDesc: "should load webhook retry config from env",
		},
		{
			name: "slack configuration from env",
			envVars: map[string]string{
				"INFRALOG_TARGET_SLACK_WEBHOOK_URL": "https://hooks.slack.com/services/xxx",
				"INFRALOG_TARGET_SLACK_CHANNEL":     "#infra",
				"INFRALOG_TARGET_SLACK_USERNAME":    "infralog-bot",
				"INFRALOG_TARGET_SLACK_ICON_EMOJI":  ":robot:",
			},
			want: Config{
				Target: Target{
					Slack: SlackConfig{
						WebhookURL: "https://hooks.slack.com/services/xxx",
						Channel:    "#infra",
						Username:   "infralog-bot",
						IconEmoji:  ":robot:",
					},
				},
			},
			wantDesc: "should load slack config from env",
		},
		{
			name: "filter configuration from env",
			envVars: map[string]string{
				"INFRALOG_FILTER_RESOURCE_TYPES": "aws_instance,aws_s3_bucket,aws_vpc",
				"INFRALOG_FILTER_OUTPUTS":        "public_ip,vpc_id",
			},
			want: Config{
				Filter: Filter{
					ResourceTypes: []string{"aws_instance", "aws_s3_bucket", "aws_vpc"},
					Outputs:       []string{"public_ip", "vpc_id"},
				},
			},
			wantDesc: "should load filter config from env",
		},
		{
			name: "filter with spaces in comma-separated list",
			envVars: map[string]string{
				"INFRALOG_FILTER_RESOURCE_TYPES": "aws_instance, aws_s3_bucket , aws_vpc",
			},
			want: Config{
				Filter: Filter{
					ResourceTypes: []string{"aws_instance", "aws_s3_bucket", "aws_vpc"},
				},
			},
			wantDesc: "should trim whitespace from comma-separated values",
		},
		{
			name: "all configuration from env",
			envVars: map[string]string{
				"INFRALOG_TARGET_WEBHOOK_URL":                     "https://example.com/webhook",
				"INFRALOG_TARGET_WEBHOOK_METHOD":                  "POST",
				"INFRALOG_TARGET_WEBHOOK_RETRY_MAX_ATTEMPTS":      "3",
				"INFRALOG_TARGET_WEBHOOK_RETRY_INITIAL_DELAY_MS":  "1000",
				"INFRALOG_TARGET_WEBHOOK_RETRY_MAX_DELAY_MS":      "30000",
				"INFRALOG_TARGET_WEBHOOK_RETRY_RETRY_ON_STATUS":   "500,502,503,504",
				"INFRALOG_TARGET_SLACK_WEBHOOK_URL":               "https://hooks.slack.com/services/xxx",
				"INFRALOG_TARGET_SLACK_CHANNEL":                   "#infra",
				"INFRALOG_FILTER_RESOURCE_TYPES":                  "aws_instance",
				"INFRALOG_FILTER_OUTPUTS":                         "public_ip",
			},
			want: Config{
				Target: Target{
					Webhook: WebhookConfig{
						URL:    "https://example.com/webhook",
						Method: "POST",
						Retry: RetryConfig{
							MaxAttempts:  3,
							InitialDelay: 1000,
							MaxDelay:     30000,
							StatusCodes:  []int{500, 502, 503, 504},
						},
					},
					Slack: SlackConfig{
						WebhookURL: "https://hooks.slack.com/services/xxx",
						Channel:    "#infra",
					},
				},
				Filter: Filter{
					ResourceTypes: []string{"aws_instance"},
					Outputs:       []string{"public_ip"},
				},
			},
			wantDesc: "should load complete config from env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for key, val := range tt.envVars {
				os.Setenv(key, val)
			}

			// Load config (no file)
			got, err := LoadConfig("")
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}

			// Check webhook config
			if got.Target.Webhook.URL != tt.want.Target.Webhook.URL {
				t.Errorf("Webhook.URL = %v, want %v", got.Target.Webhook.URL, tt.want.Target.Webhook.URL)
			}
			if got.Target.Webhook.Method != tt.want.Target.Webhook.Method {
				t.Errorf("Webhook.Method = %v, want %v", got.Target.Webhook.Method, tt.want.Target.Webhook.Method)
			}
			if got.Target.Webhook.Retry.MaxAttempts != tt.want.Target.Webhook.Retry.MaxAttempts {
				t.Errorf("Webhook.Retry.MaxAttempts = %v, want %v", got.Target.Webhook.Retry.MaxAttempts, tt.want.Target.Webhook.Retry.MaxAttempts)
			}
			if got.Target.Webhook.Retry.InitialDelay != tt.want.Target.Webhook.Retry.InitialDelay {
				t.Errorf("Webhook.Retry.InitialDelay = %v, want %v", got.Target.Webhook.Retry.InitialDelay, tt.want.Target.Webhook.Retry.InitialDelay)
			}
			if got.Target.Webhook.Retry.MaxDelay != tt.want.Target.Webhook.Retry.MaxDelay {
				t.Errorf("Webhook.Retry.MaxDelay = %v, want %v", got.Target.Webhook.Retry.MaxDelay, tt.want.Target.Webhook.Retry.MaxDelay)
			}
			if !intSliceEqual(got.Target.Webhook.Retry.StatusCodes, tt.want.Target.Webhook.Retry.StatusCodes) {
				t.Errorf("Webhook.Retry.StatusCodes = %v, want %v", got.Target.Webhook.Retry.StatusCodes, tt.want.Target.Webhook.Retry.StatusCodes)
			}

			// Check slack config
			if got.Target.Slack.WebhookURL != tt.want.Target.Slack.WebhookURL {
				t.Errorf("Slack.WebhookURL = %v, want %v", got.Target.Slack.WebhookURL, tt.want.Target.Slack.WebhookURL)
			}
			if got.Target.Slack.Channel != tt.want.Target.Slack.Channel {
				t.Errorf("Slack.Channel = %v, want %v", got.Target.Slack.Channel, tt.want.Target.Slack.Channel)
			}
			if got.Target.Slack.Username != tt.want.Target.Slack.Username {
				t.Errorf("Slack.Username = %v, want %v", got.Target.Slack.Username, tt.want.Target.Slack.Username)
			}
			if got.Target.Slack.IconEmoji != tt.want.Target.Slack.IconEmoji {
				t.Errorf("Slack.IconEmoji = %v, want %v", got.Target.Slack.IconEmoji, tt.want.Target.Slack.IconEmoji)
			}

			// Check filter config
			if !stringSliceEqual(got.Filter.ResourceTypes, tt.want.Filter.ResourceTypes) {
				t.Errorf("Filter.ResourceTypes = %v, want %v", got.Filter.ResourceTypes, tt.want.Filter.ResourceTypes)
			}
			if !stringSliceEqual(got.Filter.Outputs, tt.want.Filter.Outputs) {
				t.Errorf("Filter.Outputs = %v, want %v", got.Filter.Outputs, tt.want.Filter.Outputs)
			}
		})
	}
}

func TestEnvOverridesFile(t *testing.T) {
	// Save original environment and restore after test
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			pair := splitEnv(env)
			if len(pair) == 2 {
				os.Setenv(pair[0], pair[1])
			}
		}
	}()

	// Clear environment and set test values
	os.Clearenv()
	os.Setenv("INFRALOG_TARGET_WEBHOOK_URL", "https://env.example.com/webhook")
	os.Setenv("INFRALOG_FILTER_RESOURCE_TYPES", "aws_lambda_function")

	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	configContent := `target:
  webhook:
    url: "https://file.example.com/webhook"
    method: "POST"
filter:
  resource_types:
    - "aws_instance"
    - "aws_s3_bucket"
`
	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Load config
	got, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Env var should override file value
	if got.Target.Webhook.URL != "https://env.example.com/webhook" {
		t.Errorf("Webhook.URL = %v, want https://env.example.com/webhook (env should override file)", got.Target.Webhook.URL)
	}

	// File value should be preserved when no env var is set
	if got.Target.Webhook.Method != "POST" {
		t.Errorf("Webhook.Method = %v, want POST (file value should be preserved)", got.Target.Webhook.Method)
	}

	// Env var should override file array
	if len(got.Filter.ResourceTypes) != 1 || got.Filter.ResourceTypes[0] != "aws_lambda_function" {
		t.Errorf("Filter.ResourceTypes = %v, want [aws_lambda_function] (env should override file)", got.Filter.ResourceTypes)
	}
}

// Helper functions
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func splitEnv(env string) []string {
	parts := make([]string, 0, 2)
	if idx := indexOf(env, '='); idx >= 0 {
		parts = append(parts, env[:idx], env[idx+1:])
	}
	return parts
}

func indexOf(s string, c rune) int {
	for i, r := range s {
		if r == c {
			return i
		}
	}
	return -1
}
