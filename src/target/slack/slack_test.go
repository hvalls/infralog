package slack

import (
	"encoding/json"
	"infralog/config"
	"infralog/target"
	"infralog/tfplan"
	"net/http"
	"net/http/httptest"
	"strings"
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

	plan := &tfplan.Plan{
		ResourceChanges: []tfplan.ResourceChange{
			{
				Type: "aws_instance",
				Name: "web",
				Change: tfplan.Change{
					Actions: []string{"update"},
					Before: map[string]interface{}{
						"instance_type": "t2.micro",
					},
					After: map[string]interface{}{
						"instance_type": "t2.small",
					},
				},
			},
		},
	}
	payload := target.NewPayload(plan)
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

	plan := &tfplan.Plan{
		ResourceChanges: []tfplan.ResourceChange{
			{
				Type: "aws_s3_bucket",
				Name: "data",
				Change: tfplan.Change{
					Actions: []string{"create"},
				},
			},
		},
	}

	payload := target.NewPayload(plan)
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

	payload := target.NewPayload(&tfplan.Plan{})
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
		{"added", ":large_green_circle:"},
		{"removed", ":red_circle:"},
		{"changed", ":large_yellow_circle:"},
		{"replaced", ":large_yellow_circle:"},
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
		plan     *tfplan.Plan
		contains string
	}{
		{
			name: "Resources only",
			plan: &tfplan.Plan{
				ResourceChanges: []tfplan.ResourceChange{{}, {}},
			},
			contains: "2 resource(s)",
		},
		{
			name: "Outputs only",
			plan: &tfplan.Plan{
				OutputChanges: map[string]tfplan.OutputChange{
					"output1": {},
				},
			},
			contains: "1 output(s)",
		},
		{
			name: "Both resources and outputs",
			plan: &tfplan.Plan{
				ResourceChanges: []tfplan.ResourceChange{{}},
				OutputChanges: map[string]tfplan.OutputChange{
					"output1": {},
					"output2": {},
				},
			},
			contains: "1 resource(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := target.buildFallbackText(tt.plan)
			if result == "" {
				t.Error("Expected non-empty fallback text")
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestFormatResourceChanges(t *testing.T) {
	target := &SlackTarget{}

	changes := []tfplan.ResourceChange{
		{
			Type: "aws_instance",
			Name: "web",
			Change: tfplan.Change{
				Actions: []string{"update"},
				Before: map[string]interface{}{
					"instance_type": "t2.micro",
				},
				After: map[string]interface{}{
					"instance_type": "t2.small",
				},
			},
		},
		{
			Type: "aws_s3_bucket",
			Name: "data",
			Change: tfplan.Change{
				Actions: []string{"create"},
			},
		},
	}

	result := target.formatResourceChanges(changes)

	if result == "" {
		t.Error("Expected non-empty result")
	}
	if !strings.Contains(result, "aws_instance.web") {
		t.Error("Expected result to contain aws_instance.web")
	}
	if !strings.Contains(result, "aws_s3_bucket.data") {
		t.Error("Expected result to contain aws_s3_bucket.data")
	}
	if !strings.Contains(result, "instance_type") {
		t.Error("Expected result to contain instance_type attribute")
	}
}
