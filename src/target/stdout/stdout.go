package stdout

import (
	"encoding/json"
	"fmt"
	"infralog/config"
	"infralog/target"
	"infralog/tfstate"
	"io"
	"os"
	"strings"
	"time"
)

const (
	FormatJSON = "json"
	FormatText = "text"
)

type StdoutTarget struct {
	format string
	writer io.Writer
}

// LogEntry represents a single log line for JSON output.
type LogEntry struct {
	Timestamp    time.Time              `json:"timestamp"`
	Level        string                 `json:"level"`
	Msg          string                 `json:"msg"`
	EventType    string                 `json:"event_type"`
	Source       string                 `json:"source"`
	ResourceType string                 `json:"resource_type,omitempty"`
	ResourceName string                 `json:"resource_name,omitempty"`
	OutputName   string                 `json:"output_name,omitempty"`
	Status       string                 `json:"status"`
	Changes      map[string]ValueChange `json:"changes,omitempty"`
}

// ValueChange represents a before/after value pair.
type ValueChange struct {
	Before any `json:"before,omitempty"`
	After  any `json:"after,omitempty"`
}

func New(cfg config.StdoutConfig) *StdoutTarget {
	format := strings.ToLower(cfg.Format)
	if format != FormatJSON {
		format = FormatText
	}

	return &StdoutTarget{
		format: format,
		writer: os.Stdout,
	}
}

func (t *StdoutTarget) Write(p *target.Payload) error {
	if t.format == FormatJSON {
		return t.writeJSON(p)
	}
	return t.writeText(p)
}

func (t *StdoutTarget) writeJSON(p *target.Payload) error {
	source := buildSource(p.Metadata.TFState)
	ts := p.Metadata.Timestamp

	for _, diff := range p.Diffs.ResourceDiffs {
		entry := LogEntry{
			Timestamp:    ts,
			Level:        "info",
			Msg:          "resource " + string(diff.Status),
			EventType:    "resource_change",
			Source:       source,
			ResourceType: diff.ResourceType,
			ResourceName: diff.ResourceName,
			Status:       string(diff.Status),
		}

		if len(diff.AttributeDiffs) > 0 {
			entry.Changes = make(map[string]ValueChange)
			for attr, vd := range diff.AttributeDiffs {
				entry.Changes[attr] = ValueChange{
					Before: vd.Before,
					After:  vd.After,
				}
			}
		}

		if err := t.writeLogEntry(entry); err != nil {
			return err
		}
	}

	for _, diff := range p.Diffs.OutputDiffs {
		entry := LogEntry{
			Timestamp:  ts,
			Level:      "info",
			Msg:        "output " + string(diff.Status),
			EventType:  "output_change",
			Source:     source,
			OutputName: diff.OutputName,
			Status:     string(diff.Status),
		}

		if diff.Status == tfstate.DiffStatusChanged {
			entry.Changes = map[string]ValueChange{
				"value": {
					Before: diff.ValueDiff.Before,
					After:  diff.ValueDiff.After,
				},
			}
		}

		if err := t.writeLogEntry(entry); err != nil {
			return err
		}
	}

	return nil
}

func (t *StdoutTarget) writeLogEntry(entry LogEntry) error {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("error marshaling log entry: %w", err)
	}
	fmt.Fprintln(t.writer, string(jsonData))
	return nil
}

func buildSource(tfs config.TFState) string {
	if tfs.S3.Bucket != "" {
		return fmt.Sprintf("s3://%s/%s", tfs.S3.Bucket, tfs.S3.Key)
	}
	if tfs.Local.Path != "" {
		return "file://" + tfs.Local.Path
	}
	return ""
}

func (t *StdoutTarget) writeText(p *target.Payload) error {
	var sb strings.Builder

	tfs := p.Metadata.TFState
	timestamp := p.Metadata.Timestamp.Format("2006-01-02 15:04:05 UTC")

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("  TERRAFORM STATE CHANGES DETECTED\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("  Time:   %s\n", timestamp))
	sb.WriteString(fmt.Sprintf("  Bucket: %s\n", tfs.S3.Bucket))
	sb.WriteString(fmt.Sprintf("  Key:    %s\n", tfs.S3.Key))
	sb.WriteString(fmt.Sprintf("  Region: %s\n", tfs.S3.Region))
	sb.WriteString("──────────────────────────────────────────────────\n")

	if len(p.Diffs.ResourceDiffs) > 0 {
		sb.WriteString("\n  RESOURCE CHANGES\n\n")
		for _, diff := range p.Diffs.ResourceDiffs {
			symbol := statusSymbol(string(diff.Status))
			sb.WriteString(fmt.Sprintf("  %s %s.%s (%s)\n",
				symbol, diff.ResourceType, diff.ResourceName, diff.Status))

			if len(diff.AttributeDiffs) > 0 && diff.Status == tfstate.DiffStatusChanged {
				for attr, valueDiff := range diff.AttributeDiffs {
					sb.WriteString(fmt.Sprintf("      %s: %v → %v\n",
						attr, valueDiff.Before, valueDiff.After))
				}
			}
		}
	}

	if len(p.Diffs.OutputDiffs) > 0 {
		sb.WriteString("\n  OUTPUT CHANGES\n\n")
		for _, diff := range p.Diffs.OutputDiffs {
			symbol := statusSymbol(string(diff.Status))
			sb.WriteString(fmt.Sprintf("  %s %s (%s)\n",
				symbol, diff.OutputName, diff.Status))

			if diff.Status == tfstate.DiffStatusChanged {
				sb.WriteString(fmt.Sprintf("      %v → %v\n",
					diff.ValueDiff.Before, diff.ValueDiff.After))
			}
		}
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	fmt.Fprint(t.writer, sb.String())
	return nil
}

func statusSymbol(status string) string {
	switch status {
	case tfstate.DiffStatusAdded:
		return "[+]"
	case tfstate.DiffStatusRemoved:
		return "[-]"
	case tfstate.DiffStatusChanged:
		return "[~]"
	default:
		return "[?]"
	}
}
