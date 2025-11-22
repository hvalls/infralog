package webhook

import (
	"infralog/config"
	"infralog/target"
	"infralog/tfstate"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		cfg            config.WebhookConfig
		expectedMethod string
		expectError    bool
		errorMsg       string
	}{
		{
			name:           "Valid POST webhook",
			cfg:            config.WebhookConfig{URL: "http://example.com", Method: "POST"},
			expectedMethod: "POST",
			expectError:    false,
		},
		{
			name:           "Valid PUT webhook",
			cfg:            config.WebhookConfig{URL: "http://example.com", Method: "PUT"},
			expectedMethod: "PUT",
			expectError:    false,
		},
		{
			name:           "Empty method defaults to POST",
			cfg:            config.WebhookConfig{URL: "http://example.com", Method: ""},
			expectedMethod: "POST",
			expectError:    false,
		},
		{
			name:        "Invalid method",
			cfg:         config.WebhookConfig{URL: "http://example.com", Method: "GET"},
			expectError: true,
			errorMsg:    "invalid method: GET. Method must be POST or PUT",
		},
		{
			name:           "Lowercase method is normalized",
			cfg:            config.WebhookConfig{URL: "http://example.com", Method: "post"},
			expectedMethod: "POST",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wh, err := New(tt.cfg)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				if wh != nil {
					t.Error("Expected nil target but got a value")
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q but got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if wh == nil {
					t.Error("Expected non-nil target but got nil")
					return
				}
				if wh.method != tt.expectedMethod {
					t.Errorf("Expected Method %q but got %q", tt.expectedMethod, wh.method)
				}
			}
		})
	}
}

func TestWrite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh, err := New(config.WebhookConfig{URL: server.URL, Method: "POST"})
	if err != nil {
		t.Fatalf("Failed to create webhook target: %v", err)
	}

	payload := target.NewPayload(&tfstate.StateDiff{}, config.TFState{})

	if err := wh.Write(payload); err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
}

func TestWrite_NonRetryableError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	wh, err := New(config.WebhookConfig{URL: server.URL, Method: "POST"})
	if err != nil {
		t.Fatalf("Failed to create webhook target: %v", err)
	}

	payload := target.NewPayload(&tfstate.StateDiff{}, config.TFState{})

	err = wh.Write(payload)
	if err == nil {
		t.Error("Expected an error but got none")
	}
}

func TestWrite_RetriesOnServerError(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh, err := New(config.WebhookConfig{
		URL:    server.URL,
		Method: "POST",
		Retry: config.RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 1, // 1ms for fast tests
			MaxDelay:     10,
			StatusCodes:  []int{503},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create webhook target: %v", err)
	}

	payload := target.NewPayload(&tfstate.StateDiff{}, config.TFState{})

	if err := wh.Write(payload); err != nil {
		t.Errorf("Expected success after retries but got: %v", err)
	}

	if attempts.Load() != 3 {
		t.Errorf("Expected 3 attempts but got %d", attempts.Load())
	}
}

func TestWrite_ExhaustsRetries(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	wh, err := New(config.WebhookConfig{
		URL:    server.URL,
		Method: "POST",
		Retry: config.RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 1,
			MaxDelay:     10,
			StatusCodes:  []int{503},
		},
	})
	if err != nil {
		t.Fatalf("Failed to create webhook target: %v", err)
	}

	payload := target.NewPayload(&tfstate.StateDiff{}, config.TFState{})

	err = wh.Write(payload)
	if err == nil {
		t.Error("Expected an error after exhausting retries")
	}

	if attempts.Load() != 3 {
		t.Errorf("Expected 3 attempts but got %d", attempts.Load())
	}
}

func TestCalculateDelay(t *testing.T) {
	wh := &WebhookTarget{
		retry: config.RetryConfig{
			InitialDelay: 1000,
			MaxDelay:     30000,
		},
	}

	// Test exponential growth (with some tolerance for jitter)
	delay1 := wh.calculateDelay(1)
	delay2 := wh.calculateDelay(2)
	delay3 := wh.calculateDelay(3)

	// First delay should be around 1000ms (±25% jitter)
	if delay1 < 750*1e6 || delay1 > 1250*1e6 {
		t.Errorf("First delay %v outside expected range [750ms, 1250ms]", delay1)
	}

	// Second delay should be around 2000ms (±25% jitter)
	if delay2 < 1500*1e6 || delay2 > 2500*1e6 {
		t.Errorf("Second delay %v outside expected range [1500ms, 2500ms]", delay2)
	}

	// Third delay should be around 4000ms (±25% jitter)
	if delay3 < 3000*1e6 || delay3 > 5000*1e6 {
		t.Errorf("Third delay %v outside expected range [3000ms, 5000ms]", delay3)
	}
}

func TestCalculateDelay_CappedAtMax(t *testing.T) {
	wh := &WebhookTarget{
		retry: config.RetryConfig{
			InitialDelay: 1000,
			MaxDelay:     5000,
		},
	}

	// At attempt 10, exponential would be 1000 * 2^9 = 512000ms
	// But should be capped at 5000ms (±25% jitter)
	delay := wh.calculateDelay(10)

	if delay < 3750*1e6 || delay > 6250*1e6 {
		t.Errorf("Delay %v should be capped around 5000ms (±25%%)", delay)
	}
}

func TestShouldRetry(t *testing.T) {
	wh := &WebhookTarget{
		retry: config.RetryConfig{
			StatusCodes: []int{500, 502, 503, 504},
		},
	}

	tests := []struct {
		statusCode int
		expected   bool
	}{
		{500, true},
		{502, true},
		{503, true},
		{504, true},
		{400, false},
		{401, false},
		{404, false},
		{200, false},
	}

	for _, tt := range tests {
		result := wh.shouldRetry(tt.statusCode)
		if result != tt.expected {
			t.Errorf("shouldRetry(%d) = %v, expected %v", tt.statusCode, result, tt.expected)
		}
	}
}
