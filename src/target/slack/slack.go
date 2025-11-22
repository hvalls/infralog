package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"infralog/config"
	"infralog/target"
	"infralog/tfstate"
	"net/http"
	"strings"
)

type SlackTarget struct {
	webhookURL string
	channel    string
	username   string
	iconEmoji  string
}

type slackMessage struct {
	Channel   string  `json:"channel,omitempty"`
	Username  string  `json:"username,omitempty"`
	IconEmoji string  `json:"icon_emoji,omitempty"`
	Text      string  `json:"text"`
	Blocks    []block `json:"blocks,omitempty"`
}

type block struct {
	Type string      `json:"type"`
	Text *textObject `json:"text,omitempty"`
}

type textObject struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func New(cfg config.SlackConfig) (*SlackTarget, error) {
	if cfg.WebhookURL == "" {
		return nil, fmt.Errorf("slack webhook URL is required")
	}

	return &SlackTarget{
		webhookURL: cfg.WebhookURL,
		channel:    cfg.Channel,
		username:   cfg.Username,
		iconEmoji:  cfg.IconEmoji,
	}, nil
}

func (t *SlackTarget) Write(p *target.Payload) error {
	msg := t.buildMessage(p)

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling slack message: %w", err)
	}

	resp, err := http.Post(t.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack request failed with status code: %d", resp.StatusCode)
	}

	return nil
}

func (t *SlackTarget) buildMessage(p *target.Payload) slackMessage {
	var blocks []block

	// Header
	blocks = append(blocks, block{
		Type: "header",
		Text: &textObject{
			Type: "plain_text",
			Text: "Terraform State Changes Detected",
		},
	})

	// Context - state file info and timestamp
	tfs := p.Metadata.TFState
	stateInfo := fmt.Sprintf("*Bucket:* %s | *Key:* %s | *Region:* %s\n*Time:* %s",
		tfs.S3.Bucket, tfs.S3.Key, tfs.S3.Region,
		p.Metadata.Timestamp.Format("2006-01-02 15:04:05 UTC"))
	blocks = append(blocks, block{
		Type: "section",
		Text: &textObject{
			Type: "mrkdwn",
			Text: stateInfo,
		},
	})

	// Divider
	blocks = append(blocks, block{Type: "divider"})

	// Resource changes
	if len(p.Diffs.ResourceDiffs) > 0 {
		resourceText := t.formatResourceDiffs(p.Diffs.ResourceDiffs)
		blocks = append(blocks, block{
			Type: "section",
			Text: &textObject{
				Type: "mrkdwn",
				Text: resourceText,
			},
		})
	}

	// Output changes
	if len(p.Diffs.OutputDiffs) > 0 {
		outputText := t.formatOutputDiffs(p.Diffs.OutputDiffs)
		blocks = append(blocks, block{
			Type: "section",
			Text: &textObject{
				Type: "mrkdwn",
				Text: outputText,
			},
		})
	}

	msg := slackMessage{
		Text:   t.buildFallbackText(p.Diffs),
		Blocks: blocks,
	}

	if t.channel != "" {
		msg.Channel = t.channel
	}
	if t.username != "" {
		msg.Username = t.username
	}
	if t.iconEmoji != "" {
		msg.IconEmoji = t.iconEmoji
	}

	return msg
}

func (t *SlackTarget) formatResourceDiffs(diffs []tfstate.ResourceDiff) string {
	var sb strings.Builder
	sb.WriteString("*Resource Changes*\n\n")

	for _, diff := range diffs {
		emoji := statusEmoji(string(diff.Status))
		sb.WriteString(fmt.Sprintf("%s `%s.%s` - %s\n",
			emoji, diff.ResourceType, diff.ResourceName, diff.Status))

		if len(diff.AttributeDiffs) > 0 && diff.Status == tfstate.DiffStatusChanged {
			for attr, valueDiff := range diff.AttributeDiffs {
				sb.WriteString(fmt.Sprintf("    • `%s`: `%v` → `%v`\n",
					attr, valueDiff.Before, valueDiff.After))
			}
		}
	}

	return sb.String()
}

func (t *SlackTarget) formatOutputDiffs(diffs []tfstate.OutputDiff) string {
	var sb strings.Builder
	sb.WriteString("*Output Changes*\n\n")

	for _, diff := range diffs {
		emoji := statusEmoji(string(diff.Status))
		sb.WriteString(fmt.Sprintf("%s `%s` - %s\n",
			emoji, diff.OutputName, diff.Status))

		if diff.Status == tfstate.DiffStatusChanged {
			sb.WriteString(fmt.Sprintf("    • `%v` → `%v`\n",
				diff.ValueDiff.Before, diff.ValueDiff.After))
		}
	}

	return sb.String()
}

func (t *SlackTarget) buildFallbackText(d *tfstate.StateDiff) string {
	resourceCount := len(d.ResourceDiffs)
	outputCount := len(d.OutputDiffs)

	parts := []string{}
	if resourceCount > 0 {
		parts = append(parts, fmt.Sprintf("%d resource(s)", resourceCount))
	}
	if outputCount > 0 {
		parts = append(parts, fmt.Sprintf("%d output(s)", outputCount))
	}

	return fmt.Sprintf("Terraform state changes detected: %s changed", strings.Join(parts, ", "))
}

func statusEmoji(status string) string {
	switch status {
	case tfstate.DiffStatusAdded:
		return ":large_green_circle:"
	case tfstate.DiffStatusRemoved:
		return ":red_circle:"
	case tfstate.DiffStatusChanged:
		return ":large_yellow_circle:"
	default:
		return ":white_circle:"
	}
}
