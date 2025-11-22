package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"infralog/config"
	"infralog/tfstate"
	"math"
	"math/rand"
	"net/http"
	"slices"
	"strings"
	"time"
)

type WebhookTarget struct {
	url    string
	method string
	retry  config.RetryConfig
}

func New(cfg config.WebhookConfig) (*WebhookTarget, error) {
	method := strings.ToUpper(cfg.Method)
	if method == "" {
		method = "POST"
	} else if method != "POST" && method != "PUT" {
		return nil, fmt.Errorf("invalid method: %s. Method must be POST or PUT", cfg.Method)
	}

	return &WebhookTarget{
		url:    cfg.URL,
		method: method,
		retry:  cfg.Retry.WithDefaults(),
	}, nil
}

func (t *WebhookTarget) Write(d *tfstate.StateDiff, tfs config.TFState) error {
	jsonData, err := getJSONBody(d, tfs)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 1; attempt <= t.retry.MaxAttempts; attempt++ {
		statusCode, err := t.doRequest(jsonData)
		if err != nil {
			lastErr = err
			if attempt < t.retry.MaxAttempts {
				t.sleep(attempt)
			}
			continue
		}

		if statusCode >= 200 && statusCode < 300 {
			return nil
		}

		lastErr = fmt.Errorf("webhook request failed with status code: %d", statusCode)

		if !t.shouldRetry(statusCode) {
			return lastErr
		}

		if attempt < t.retry.MaxAttempts {
			t.sleep(attempt)
		}
	}

	return fmt.Errorf("webhook request failed after %d attempts: %w", t.retry.MaxAttempts, lastErr)
}

func (t *WebhookTarget) doRequest(jsonData []byte) (int, error) {
	req, err := http.NewRequest(t.method, t.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error making webhook request: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func (t *WebhookTarget) shouldRetry(statusCode int) bool {
	return slices.Contains(t.retry.StatusCodes, statusCode)
}

func (t *WebhookTarget) sleep(attempt int) {
	delay := t.calculateDelay(attempt)
	time.Sleep(delay)
}

func (t *WebhookTarget) calculateDelay(attempt int) time.Duration {
	// Exponential backoff: initialDelay * 2^(attempt-1)
	backoff := float64(t.retry.InitialDelay) * math.Pow(2, float64(attempt-1))

	// Cap at max delay
	if backoff > float64(t.retry.MaxDelay) {
		backoff = float64(t.retry.MaxDelay)
	}

	// Add jitter (±25%)
	jitter := backoff * 0.25 * (rand.Float64()*2 - 1)
	delay := backoff + jitter

	return time.Duration(delay) * time.Millisecond
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
