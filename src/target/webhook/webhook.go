package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"infralog/config"
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

func (t *WebhookTarget) Write(d *tfstate.StateDiff, tfs config.TFState) error {
	jsonData, err := getJSONBody(d, tfs)
	if err != nil {
		return err
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

func getJSONBody(d *tfstate.StateDiff, tfs config.TFState) ([]byte, error) {
	body := struct {
		Diffs    *tfstate.StateDiff `json:"diffs"`
		Metadata struct {
			TFState config.TFState `json:"tfstate"`
		} `json:"metadata"`
	}{
		Diffs: d,
		Metadata: struct {
			TFState config.TFState `json:"tfstate"`
		}{
			TFState: tfs,
		},
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling webhook body: %w", err)
	}

	return jsonData, nil
}
