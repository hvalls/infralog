package webhook

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		method         string
		expectedURL    string
		expectedMethod string
		expectError    bool
		errorMsg       string
	}{
		{
			name:           "Valid POST webhook",
			url:            "http://example.com",
			method:         "POST",
			expectedURL:    "http://example.com",
			expectedMethod: "POST",
			expectError:    false,
		},
		{
			name:           "Valid PUT webhook",
			url:            "http://example.com",
			method:         "PUT",
			expectedURL:    "http://example.com",
			expectedMethod: "PUT",
			expectError:    false,
		},
		{
			name:           "Empty method defaults to POST",
			url:            "http://example.com",
			method:         "",
			expectedURL:    "http://example.com",
			expectedMethod: "POST",
			expectError:    false,
		},
		{
			name:        "Invalid method",
			url:         "http://example.com",
			method:      "GET",
			expectError: true,
			errorMsg:    "invalid method: GET. Method must be POST or PUT",
		},
		{
			name:           "Case insensitive POST",
			url:            "http://example.com",
			method:         "post",
			expectedURL:    "http://example.com",
			expectedMethod: "post",
			expectError:    false,
		},
		{
			name:           "Case insensitive PUT",
			url:            "http://example.com",
			method:         "put",
			expectedURL:    "http://example.com",
			expectedMethod: "put",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := New(tt.url, tt.method)

			if tt.expectError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				if target != nil {
					t.Error("Expected nil target but got a value")
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q but got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if target == nil {
					t.Error("Expected non-nil target but got nil")
					return
				}
				if target.URL != tt.expectedURL {
					t.Errorf("Expected URL %q but got %q", tt.expectedURL, target.URL)
				}
				if target.Method != tt.expectedMethod {
					t.Errorf("Expected Method %q but got %q", tt.expectedMethod, target.Method)
				}
			}
		})
	}
}
