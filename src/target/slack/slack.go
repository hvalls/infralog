package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"infralog/config"
	"infralog/target"
	"infralog/tfplan"
	"net/http"
	"sort"
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
			Text: "Terraform Plan Changes",
		},
	})

	// Context - timestamp and git metadata
	contextText := fmt.Sprintf("*Time:* %s", p.Datetime.Format("2006-01-02 15:04:05 UTC"))

	// Add git metadata if available
	if p.Metadata != nil && p.Metadata.Git != nil {
		git := p.Metadata.Git
		contextText += "\n\n*Git Context*\n"

		if git.Committer != "" {
			contextText += fmt.Sprintf("👤 *Committer:* %s\n", git.Committer)
		}
		if git.Branch != "" {
			contextText += fmt.Sprintf("🌿 *Branch:* `%s`\n", git.Branch)
		}
		if git.CommitSHA != "" {
			// Show short SHA (first 8 characters)
			shortSHA := git.CommitSHA
			if len(shortSHA) > 8 {
				shortSHA = shortSHA[:8]
			}
			contextText += fmt.Sprintf("📝 *Commit:* `%s`\n", shortSHA)
		}
		if git.RepoURL != "" {
			contextText += fmt.Sprintf("🔗 *Repository:* %s\n", git.RepoURL)
		}
	}

	blocks = append(blocks, block{
		Type: "section",
		Text: &textObject{
			Type: "mrkdwn",
			Text: contextText,
		},
	})

	// Divider
	blocks = append(blocks, block{Type: "divider"})

	// Resource changes
	if len(p.Plan.ResourceChanges) > 0 {
		resourceText := t.formatResourceChanges(p.Plan.ResourceChanges)
		blocks = append(blocks, block{
			Type: "section",
			Text: &textObject{
				Type: "mrkdwn",
				Text: resourceText,
			},
		})
	}

	// Output changes
	if len(p.Plan.OutputChanges) > 0 {
		outputText := t.formatOutputChanges(p.Plan.OutputChanges)
		blocks = append(blocks, block{
			Type: "section",
			Text: &textObject{
				Type: "mrkdwn",
				Text: outputText,
			},
		})
	}

	msg := slackMessage{
		Text:   t.buildFallbackText(p.Plan),
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

func (t *SlackTarget) formatResourceChanges(changes []tfplan.ResourceChange) string {
	var sb strings.Builder
	sb.WriteString("*Resource Changes*\n\n")

	for _, rc := range changes {
		status := actionsToStatus(rc.Change.Actions)
		emoji := statusEmoji(status)
		sb.WriteString(fmt.Sprintf("%s `%s.%s` - %s\n",
			emoji, rc.Type, rc.Name, status))

		// Show changed attributes for updates
		if status == "changed" || status == "replaced" {
			changes := extractChanges(rc.Change.Before, rc.Change.After)
			// Limit to first 5 changed attributes to avoid excessive Slack message length
			count := 0
			for attr, change := range changes {
				if count >= 5 {
					sb.WriteString(fmt.Sprintf("    • _...and %d more attributes_\n", len(changes)-5))
					break
				}
				sb.WriteString(fmt.Sprintf("    • `%s`: `%v` → `%v`\n",
					attr, change.Before, change.After))
				count++
			}
		}
	}

	return sb.String()
}

func (t *SlackTarget) formatOutputChanges(changes map[string]tfplan.OutputChange) string {
	var sb strings.Builder
	sb.WriteString("*Output Changes*\n\n")

	// Sort output names for consistent ordering
	names := make([]string, 0, len(changes))
	for name := range changes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		oc := changes[name]
		status := actionsToStatus(oc.Change.Actions)
		emoji := statusEmoji(status)
		sb.WriteString(fmt.Sprintf("%s `%s` - %s\n",
			emoji, name, status))

		if status == "changed" || status == "replaced" {
			sb.WriteString(fmt.Sprintf("    • `%v` → `%v`\n",
				oc.Change.Before, oc.Change.After))
		}
	}

	return sb.String()
}

func (t *SlackTarget) buildFallbackText(plan *tfplan.Plan) string {
	resourceCount := len(plan.ResourceChanges)
	outputCount := len(plan.OutputChanges)

	parts := []string{}
	if resourceCount > 0 {
		parts = append(parts, fmt.Sprintf("%d resource(s)", resourceCount))
	}
	if outputCount > 0 {
		parts = append(parts, fmt.Sprintf("%d output(s)", outputCount))
	}

	return fmt.Sprintf("Terraform plan changes detected: %s changed", strings.Join(parts, ", "))
}

// actionsToStatus maps Terraform plan actions to a readable status string.
func actionsToStatus(actions []string) string {
	if len(actions) == 0 {
		return "unknown"
	}

	// Sort actions to normalize ordering
	sortedActions := make([]string, len(actions))
	copy(sortedActions, actions)
	sort.Strings(sortedActions)

	// Single action cases
	if len(sortedActions) == 1 {
		switch sortedActions[0] {
		case "create":
			return "added"
		case "delete":
			return "removed"
		case "update":
			return "changed"
		default:
			return sortedActions[0]
		}
	}

	// Multiple actions (typically replace operations: create + delete)
	if len(sortedActions) == 2 {
		if sortedActions[0] == "create" && sortedActions[1] == "delete" {
			return "replaced"
		}
	}

	return "changed"
}

func statusEmoji(status string) string {
	switch status {
	case "added":
		return ":large_green_circle:"
	case "removed":
		return ":red_circle:"
	case "changed", "replaced":
		return ":large_yellow_circle:"
	default:
		return ":white_circle:"
	}
}

// ValueChange represents a before/after value pair for Slack formatting.
type ValueChange struct {
	Before interface{}
	After  interface{}
}

// extractChanges compares before and after attribute maps and returns changed attributes.
func extractChanges(before, after map[string]interface{}) map[string]ValueChange {
	changes := make(map[string]ValueChange)

	// Collect all unique attribute keys
	allKeys := make(map[string]bool)
	for key := range before {
		allKeys[key] = true
	}
	for key := range after {
		allKeys[key] = true
	}

	// Compare each attribute
	for key := range allKeys {
		beforeVal := before[key]
		afterVal := after[key]

		// Check if values are different (simple comparison)
		if fmt.Sprintf("%v", beforeVal) != fmt.Sprintf("%v", afterVal) {
			changes[key] = ValueChange{
				Before: beforeVal,
				After:  afterVal,
			}
		}
	}

	return changes
}
