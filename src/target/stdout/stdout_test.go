package stdout

import (
	"bytes"
	"encoding/json"
	"infralog/config"
	"infralog/target"
	"infralog/tfstate"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		cfg            config.StdoutConfig
		expectedFormat string
	}{
		{
			name:           "Default format is text",
			cfg:            config.StdoutConfig{Enabled: true},
			expectedFormat: FormatText,
		},
		{
			name:           "JSON format",
			cfg:            config.StdoutConfig{Enabled: true, Format: "json"},
			expectedFormat: FormatJSON,
		},
		{
			name:           "Text format explicit",
			cfg:            config.StdoutConfig{Enabled: true, Format: "text"},
			expectedFormat: FormatText,
		},
		{
			name:           "Case insensitive JSON",
			cfg:            config.StdoutConfig{Enabled: true, Format: "JSON"},
			expectedFormat: FormatJSON,
		},
		{
			name:           "Invalid format defaults to text",
			cfg:            config.StdoutConfig{Enabled: true, Format: "xml"},
			expectedFormat: FormatText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := New(tt.cfg)
			if target.format != tt.expectedFormat {
				t.Errorf("Expected format %q, got %q", tt.expectedFormat, target.format)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	stdoutTarget := &StdoutTarget{format: FormatJSON, writer: &buf}

	diff := &tfstate.StateDiff{
		ResourceDiffs: []tfstate.ResourceDiff{
			{
				ResourceType: "aws_instance",
				ResourceName: "web",
				Status:       tfstate.DiffStatusChanged,
				AttributeDiffs: map[string]tfstate.ValueDiff{
					"instance_type": {Before: "t2.micro", After: "t2.small"},
				},
			},
			{
				ResourceType: "aws_s3_bucket",
				ResourceName: "data",
				Status:       tfstate.DiffStatusAdded,
			},
		},
		OutputDiffs: []tfstate.OutputDiff{
			{
				OutputName: "endpoint",
				Status:     tfstate.DiffStatusChanged,
				ValueDiff:  tfstate.ValueDiff{Before: "old.example.com", After: "new.example.com"},
			},
		},
	}
	tfs := config.TFState{}
	tfs.S3.Bucket = "my-bucket"
	tfs.S3.Key = "terraform.tfstate"
	tfs.S3.Region = "us-east-1"

	payload := target.NewPayload(diff, tfs)
	if err := stdoutTarget.Write(payload); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 3 lines: 2 resource changes + 1 output change
	if len(lines) != 3 {
		t.Errorf("Expected 3 JSON lines, got %d", len(lines))
	}

	// Verify each line is valid JSON with expected structure
	for i, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
			continue
		}

		// Verify common fields
		if entry.Level != "info" {
			t.Errorf("Line %d: expected level 'info', got %q", i, entry.Level)
		}
		if entry.Source != "s3://my-bucket/terraform.tfstate" {
			t.Errorf("Line %d: expected source 's3://my-bucket/terraform.tfstate', got %q", i, entry.Source)
		}
		if entry.Timestamp.IsZero() {
			t.Errorf("Line %d: timestamp should not be zero", i)
		}
	}

	// Verify first line (resource changed)
	var first LogEntry
	json.Unmarshal([]byte(lines[0]), &first)
	if first.EventType != "resource_change" {
		t.Errorf("Expected event_type 'resource_change', got %q", first.EventType)
	}
	if first.ResourceType != "aws_instance" {
		t.Errorf("Expected resource_type 'aws_instance', got %q", first.ResourceType)
	}
	if first.Status != "changed" {
		t.Errorf("Expected status 'changed', got %q", first.Status)
	}
	if first.Changes == nil || first.Changes["instance_type"].Before != "t2.micro" {
		t.Error("Expected changes with instance_type before value")
	}

	// Verify third line (output changed)
	var third LogEntry
	json.Unmarshal([]byte(lines[2]), &third)
	if third.EventType != "output_change" {
		t.Errorf("Expected event_type 'output_change', got %q", third.EventType)
	}
	if third.OutputName != "endpoint" {
		t.Errorf("Expected output_name 'endpoint', got %q", third.OutputName)
	}
}

func TestWriteText(t *testing.T) {
	var buf bytes.Buffer
	stdoutTarget := &StdoutTarget{format: FormatText, writer: &buf}

	diff := &tfstate.StateDiff{
		ResourceDiffs: []tfstate.ResourceDiff{
			{
				ResourceType: "aws_instance",
				ResourceName: "web",
				Status:       tfstate.DiffStatusAdded,
			},
			{
				ResourceType: "aws_s3_bucket",
				ResourceName: "data",
				Status:       tfstate.DiffStatusRemoved,
			},
			{
				ResourceType: "aws_rds_instance",
				ResourceName: "db",
				Status:       tfstate.DiffStatusChanged,
				AttributeDiffs: map[string]tfstate.ValueDiff{
					"instance_class": {Before: "db.t2.micro", After: "db.t2.small"},
				},
			},
		},
		OutputDiffs: []tfstate.OutputDiff{
			{
				OutputName: "endpoint",
				Status:     tfstate.DiffStatusChanged,
				ValueDiff:  tfstate.ValueDiff{Before: "old.example.com", After: "new.example.com"},
			},
		},
	}
	tfs := config.TFState{}
	tfs.S3.Bucket = "my-bucket"
	tfs.S3.Key = "terraform.tfstate"
	tfs.S3.Region = "us-east-1"

	payload := target.NewPayload(diff, tfs)
	if err := stdoutTarget.Write(payload); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()

	// Verify header
	if !strings.Contains(output, "TERRAFORM STATE CHANGES DETECTED") {
		t.Error("Expected header in output")
	}

	// Verify bucket info
	if !strings.Contains(output, "my-bucket") {
		t.Error("Expected bucket name in output")
	}

	// Verify resource changes
	if !strings.Contains(output, "[+] aws_instance.web") {
		t.Error("Expected added resource with [+] symbol")
	}
	if !strings.Contains(output, "[-] aws_s3_bucket.data") {
		t.Error("Expected removed resource with [-] symbol")
	}
	if !strings.Contains(output, "[~] aws_rds_instance.db") {
		t.Error("Expected changed resource with [~] symbol")
	}

	// Verify attribute diff
	if !strings.Contains(output, "instance_class") {
		t.Error("Expected attribute name in output")
	}

	// Verify output changes
	if !strings.Contains(output, "OUTPUT CHANGES") {
		t.Error("Expected OUTPUT CHANGES section")
	}
	if !strings.Contains(output, "endpoint") {
		t.Error("Expected output name in output")
	}
}

func TestStatusSymbol(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{tfstate.DiffStatusAdded, "[+]"},
		{tfstate.DiffStatusRemoved, "[-]"},
		{tfstate.DiffStatusChanged, "[~]"},
		{"unknown", "[?]"},
	}

	for _, tt := range tests {
		result := statusSymbol(tt.status)
		if result != tt.expected {
			t.Errorf("statusSymbol(%q) = %q, expected %q", tt.status, result, tt.expected)
		}
	}
}
