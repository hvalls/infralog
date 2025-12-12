package target

import (
	"encoding/json"
	"infralog/tfplan"
	"testing"
	"time"
)

func TestNewPayload(t *testing.T) {
	plan := &tfplan.Plan{
		TerraformVersion: "1.5.0",
		ResourceChanges:  []tfplan.ResourceChange{},
		OutputChanges:    map[string]tfplan.OutputChange{},
	}

	payload := NewPayload(plan)

	// Verify basic structure
	if payload == nil {
		t.Fatal("NewPayload returned nil")
	}

	if payload.Plan != plan {
		t.Error("Payload.Plan does not match input plan")
	}

	// Verify datetime is recent (within last second)
	now := time.Now().UTC()
	diff := now.Sub(payload.Datetime)
	if diff < 0 || diff > time.Second {
		t.Errorf("Payload.Datetime is not recent: %v (diff: %v)", payload.Datetime, diff)
	}

	// Metadata may or may not be present depending on whether we're in a git repo
	// We just verify the structure is valid
	if payload.Metadata != nil {
		t.Logf("Metadata present: %+v", payload.Metadata)
		if payload.Metadata.Git != nil {
			t.Logf("Git metadata: %+v", payload.Metadata.Git)
		}
	} else {
		t.Log("No metadata present (not in git repo or git unavailable)")
	}
}

func TestPayload_JSONSerialization_WithMetadata(t *testing.T) {
	plan := &tfplan.Plan{
		TerraformVersion: "1.5.0",
	}

	payload := NewPayload(plan)

	// Serialize to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Deserialize back
	var decoded Payload
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	// Verify structure integrity
	if decoded.Plan == nil {
		t.Error("Decoded payload has nil Plan")
	}

	if decoded.Datetime.IsZero() {
		t.Error("Decoded payload has zero Datetime")
	}

	// Log the JSON output for inspection
	t.Logf("JSON payload: %s", string(jsonData))
}

func TestPayload_JSONSerialization_OmitemptyBehavior(t *testing.T) {
	// Create a payload with nil metadata to test omitempty
	plan := &tfplan.Plan{
		TerraformVersion: "1.5.0",
	}

	payload := &Payload{
		Plan:     plan,
		Datetime: time.Now().UTC(),
		Metadata: nil, // Explicitly nil
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	jsonString := string(jsonData)
	t.Logf("JSON with nil metadata: %s", jsonString)

	// Verify that "metadata" key is omitted when nil
	var rawJSON map[string]interface{}
	if err := json.Unmarshal(jsonData, &rawJSON); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	if _, exists := rawJSON["metadata"]; exists {
		t.Error("Expected 'metadata' key to be omitted when nil, but it was present")
	}
}

func TestPayloadMetadata_OmitemptyGit(t *testing.T) {
	// Test that git field is omitted when nil
	metadata := &PayloadMetadata{
		Git: nil,
	}

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal metadata: %v", err)
	}

	jsonString := string(jsonData)
	t.Logf("JSON with nil git: %s", jsonString)

	// Should be empty object "{}"
	if jsonString != "{}" {
		t.Errorf("Expected empty object, got: %s", jsonString)
	}
}
