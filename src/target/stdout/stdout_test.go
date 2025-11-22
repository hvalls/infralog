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
	if err := stdoutTarget.Write(payload); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Verify structure
	if _, ok := parsed["diffs"]; !ok {
		t.Error("Expected 'diffs' in output")
	}
	if _, ok := parsed["metadata"]; !ok {
		t.Error("Expected 'metadata' in output")
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
					"instance_class": {OldValue: "db.t2.micro", NewValue: "db.t2.small"},
				},
			},
		},
		OutputDiffs: []tfstate.OutputDiff{
			{
				OutputName: "endpoint",
				Status:     tfstate.DiffStatusChanged,
				ValueDiff:  tfstate.ValueDiff{OldValue: "old.example.com", NewValue: "new.example.com"},
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
