package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"infralog/tfstate"
	"net/http"
)

type WebhookTarget struct {
	URL string
}

func New(url string) *WebhookTarget {
	return &WebhookTarget{
		URL: url,
	}
}

func (t *WebhookTarget) Write(d *tfstate.StateDiff) error {
	jsonData, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("error marshaling state diff: %w", err)
	}

	resp, err := http.Post(t.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error making webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook request failed with status code: %d", resp.StatusCode)
	}

	return nil
}
