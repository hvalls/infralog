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
)

const (
	FormatJSON = "json"
	FormatText = "text"
)

type StdoutTarget struct {
	format string
	writer io.Writer
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
	jsonData, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling output: %w", err)
	}

	fmt.Fprintln(t.writer, string(jsonData))
	return nil
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
						attr, valueDiff.OldValue, valueDiff.NewValue))
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
					diff.ValueDiff.OldValue, diff.ValueDiff.NewValue))
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
