package slack

import (
	"encoding/json"
	"infralog/config"
	"infralog/target"
	"infralog/tfstate"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.SlackConfig
		expectError bool
	}{
		{
			name:        "Valid config with webhook URL",
			cfg:         config.SlackConfig{WebhookURL: "https://hooks.slack.com/services/xxx"},
			expectError: false,
		},
		{
			name:        "Empty webhook URL",
			cfg:         config.SlackConfig{WebhookURL: ""},
			expectError: true,
		},
		{
			name: "Config with all options",
			cfg: config.SlackConfig{
				WebhookURL: "https://hooks.slack.com/services/xxx",
				Channel:    "#alerts",
				Username:   "Infralog",
				IconEmoji:  ":terraform:",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := New(tt.cfg)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				if target != nil {
					t.Error("Expected nil target but got a value")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if target == nil {
					t.Error("Expected non-nil target but got nil")
				}
			}
		})
	}
}

func TestWrite_Success(t *testing.T) {
	var receivedBody slackMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}

		if err := json.NewDecoder(r.Body).Decode(&receivedBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	slackTarget, err := New(config.SlackConfig{WebhookURL: server.URL})
	if err != nil {
		t.Fatalf("Failed to create slack target: %v", err)
	}

	diff := &tfstate.StateDiff{
		ResourceDiffs: []tfstate.ResourceDiff{
			{
				ResourceType: "aws_instance",
				ResourceName: "web",
				Status:       tfstate.DiffStatusChanged,
				AttributeDiffs: map[string]tfstate.ValueDiff{
					"instance_type": {OldValue: "t2.micro", NewValue: "t2.small"},
				},
			},
		},
	}
	tfs := config.TFState{}
	tfs.S3.Bucket = "my-bucket"
	tfs.S3.Key = "terraform.tfstate"
	tfs.S3.Region = "us-east-1"

	payload := target.NewPayload(diff, tfs)
	if err := slackTarget.Write(payload); err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	if receivedBody.Text == "" {
		t.Error("Expected non-empty fallback text")
	}
	if len(receivedBody.Blocks) == 0 {
		t.Error("Expected blocks in message")
	}
}

func TestWrite_WithOptionalFields(t *testing.T) {
	var receivedBody slackMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	slackTarget, err := New(config.SlackConfig{
		WebhookURL: server.URL,
		Channel:    "#alerts",
		Username:   "Infralog Bot",
		IconEmoji:  ":robot:",
	})
	if err != nil {
		t.Fatalf("Failed to create slack target: %v", err)
	}

	diff := &tfstate.StateDiff{
		ResourceDiffs: []tfstate.ResourceDiff{
			{ResourceType: "aws_s3_bucket", ResourceName: "data", Status: tfstate.DiffStatusAdded},
		},
	}

	payload := target.NewPayload(diff, config.TFState{})
	if err := slackTarget.Write(payload); err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	if receivedBody.Channel != "#alerts" {
		t.Errorf("Expected channel #alerts, got %s", receivedBody.Channel)
	}
	if receivedBody.Username != "Infralog Bot" {
		t.Errorf("Expected username Infralog Bot, got %s", receivedBody.Username)
	}
	if receivedBody.IconEmoji != ":robot:" {
		t.Errorf("Expected icon_emoji :robot:, got %s", receivedBody.IconEmoji)
	}
}

func TestWrite_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	slackTarget, err := New(config.SlackConfig{WebhookURL: server.URL})
	if err != nil {
		t.Fatalf("Failed to create slack target: %v", err)
	}

	payload := target.NewPayload(&tfstate.StateDiff{}, config.TFState{})
	err = slackTarget.Write(payload)
	if err == nil {
		t.Error("Expected an error but got none")
	}
}

func TestStatusEmoji(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{tfstate.DiffStatusAdded, ":large_green_circle:"},
		{tfstate.DiffStatusRemoved, ":red_circle:"},
		{tfstate.DiffStatusChanged, ":large_yellow_circle:"},
		{"unknown", ":white_circle:"},
	}

	for _, tt := range tests {
		result := statusEmoji(tt.status)
		if result != tt.expected {
			t.Errorf("statusEmoji(%s) = %s, expected %s", tt.status, result, tt.expected)
		}
	}
}

func TestBuildFallbackText(t *testing.T) {
	target := &SlackTarget{}

	tests := []struct {
		name     string
		diff     *tfstate.StateDiff
		contains string
	}{
		{
			name: "Resources only",
			diff: &tfstate.StateDiff{
				ResourceDiffs: []tfstate.ResourceDiff{{}, {}},
			},
			contains: "2 resource(s)",
		},
		{
			name: "Outputs only",
			diff: &tfstate.StateDiff{
				OutputDiffs: []tfstate.OutputDiff{{}},
			},
			contains: "1 output(s)",
		},
		{
			name: "Both resources and outputs",
			diff: &tfstate.StateDiff{
				ResourceDiffs: []tfstate.ResourceDiff{{}},
				OutputDiffs:   []tfstate.OutputDiff{{}, {}},
			},
			contains: "1 resource(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := target.buildFallbackText(tt.diff)
			if result == "" {
				t.Error("Expected non-empty fallback text")
			}
		})
	}
}

func TestFormatResourceDiffs(t *testing.T) {
	target := &SlackTarget{}

	diffs := []tfstate.ResourceDiff{
		{
			ResourceType: "aws_instance",
			ResourceName: "web",
			Status:       tfstate.DiffStatusChanged,
			AttributeDiffs: map[string]tfstate.ValueDiff{
				"instance_type": {OldValue: "t2.micro", NewValue: "t2.small"},
			},
		},
		{
			ResourceType: "aws_s3_bucket",
			ResourceName: "data",
			Status:       tfstate.DiffStatusAdded,
		},
	}

	result := target.formatResourceDiffs(diffs)

	if result == "" {
		t.Error("Expected non-empty result")
	}
	if !contains(result, "aws_instance.web") {
		t.Error("Expected result to contain aws_instance.web")
	}
	if !contains(result, "aws_s3_bucket.data") {
		t.Error("Expected result to contain aws_s3_bucket.data")
	}
	if !contains(result, "instance_type") {
		t.Error("Expected result to contain instance_type attribute")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
