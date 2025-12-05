package main

import (
	"flag"
	"fmt"
	"infralog/config"
	"infralog/target"
	"infralog/target/slack"
	"infralog/target/webhook"
	"infralog/tfplan"
	"os"
	"sort"
)

func main() {
	// Parse CLI flags
	planFile := flag.String("plan-file", "", "Path to Terraform plan JSON file (required)")
	planFileShort := flag.String("f", "", "Path to Terraform plan JSON file (shorthand)")
	configFile := flag.String("config-file", "", "Path to configuration file (optional)")
	flag.Parse()

	// Determine plan file (prefer -f, then --plan-file)
	plan := *planFileShort
	if plan == "" {
		plan = *planFile
	}

	if plan == "" {
		fmt.Println("Error: --plan-file or -f is required")
		fmt.Println("\nUsage: infralog -f <plan.json> [--config-file <config.yml>]")
		fmt.Println("\nExample:")
		fmt.Println("  terraform show -json plan.tfplan > plan.json")
		fmt.Println("  infralog -f plan.json --config-file config.yml")
		os.Exit(1)
	}

	// Load configuration (optional)
	cfg := loadConfig(*configFile)

	// Initialize targets
	targets := initTargets(cfg)

	// Parse plan file
	terraformPlan, err := tfplan.ParsePlanFile(plan)
	if err != nil {
		fmt.Printf("Error parsing plan file: %v\n", err)
		os.Exit(1)
	}

	// Apply filters to plan
	filteredPlan := tfplan.ApplyFilter(terraformPlan, cfg.Filter)

	// Exit early if no changes
	if !filteredPlan.HasChanges() {
		fmt.Println("No changes detected in plan")
		os.Exit(0)
	}

	// Notify targets
	hasNotificationTargets := len(targets) > 0
	if err := notifyTargets(targets, filteredPlan, plan); err != nil {
		fmt.Fprintf(os.Stderr, "Error notifying targets: %v\n", err)
		os.Exit(1)
	}

	// Print output based on whether notification targets exist
	if hasNotificationTargets {
		printNotificationSummary(filteredPlan, targets)
	} else {
		printDetailedSummary(filteredPlan, plan)
	}

	os.Exit(0)
}

// loadConfig loads the configuration file if provided, otherwise returns empty config.
func loadConfig(configPath string) *config.Config {
	if configPath == "" {
		// No config provided - use defaults (empty filter matches all)
		return &config.Config{}
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	return cfg
}

// initTargets creates notification targets based on configuration.
func initTargets(cfg *config.Config) []target.Target {
	var targets []target.Target

	if cfg.Target.Webhook.URL != "" {
		t, err := webhook.New(cfg.Target.Webhook)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating webhook target: %v\n", err)
			os.Exit(1)
		}
		targets = append(targets, t)
	}

	if cfg.Target.Slack.WebhookURL != "" {
		t, err := slack.New(cfg.Target.Slack)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating slack target: %v\n", err)
			os.Exit(1)
		}
		targets = append(targets, t)
	}

	return targets
}

// notifyTargets sends the plan to all configured targets.
func notifyTargets(targets []target.Target, plan *tfplan.Plan, planFile string) error {
	payload := target.NewPayload(plan)

	var hasError bool
	for _, t := range targets {
		if err := t.Write(payload); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to target: %v\n", err)
			hasError = true
		}
	}

	if hasError {
		return fmt.Errorf("one or more targets failed")
	}

	return nil
}

// printDetailedSummary prints a detailed summary for local usage (no notification targets).
func printDetailedSummary(plan *tfplan.Plan, planFile string) {
	resourceCount := len(plan.ResourceChanges)
	outputCount := len(plan.OutputChanges)

	fmt.Printf("✓ Plan analyzed: %d resource(s) changed, %d output(s) changed\n",
		resourceCount, outputCount)

	if resourceCount > 0 {
		for _, rc := range plan.ResourceChanges {
			symbol := actionSymbol(rc.Change.Actions)
			fmt.Printf("  %s %s.%s\n", symbol, rc.Type, rc.Name)
		}
	}

	if outputCount > 0 {
		// Sort output names for consistent display
		names := make([]string, 0, len(plan.OutputChanges))
		for name := range plan.OutputChanges {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			oc := plan.OutputChanges[name]
			symbol := actionSymbol(oc.Change.Actions)
			fmt.Printf("  %s output.%s\n", symbol, name)
		}
	}
}

// printNotificationSummary prints a minimal summary when notification targets exist.
func printNotificationSummary(plan *tfplan.Plan, targets []target.Target) {
	resourceCount := len(plan.ResourceChanges)
	outputCount := len(plan.OutputChanges)

	fmt.Printf("✓ Plan analyzed: %d resource(s) changed, %d output(s) changed\n",
		resourceCount, outputCount)

	for _, t := range targets {
		fmt.Printf("✓ %s notification sent\n", targetName(t))
	}
}

// actionSymbol returns a symbol for the given action list.
func actionSymbol(actions []string) string {
	if len(actions) == 0 {
		return "[?]"
	}

	// Sort for consistent comparison
	sorted := make([]string, len(actions))
	copy(sorted, actions)
	sort.Strings(sorted)

	if len(sorted) == 1 {
		switch sorted[0] {
		case "create":
			return "[+]"
		case "delete":
			return "[-]"
		case "update":
			return "[~]"
		default:
			return "[?]"
		}
	}

	// Multiple actions (e.g., replace: create + delete)
	if len(sorted) == 2 && sorted[0] == "create" && sorted[1] == "delete" {
		return "[~]"
	}

	return "[~]"
}

// targetName returns a human-readable name for the target.
func targetName(t target.Target) string {
	switch t.(type) {
	case *webhook.WebhookTarget:
		return "Webhook"
	case *slack.SlackTarget:
		return "Slack"
	default:
		return "Target"
	}
}
