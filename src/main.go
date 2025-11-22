package main

import (
	"context"
	"flag"
	"fmt"
	"infralog/backend"
	"infralog/backend/local"
	"infralog/backend/s3"
	"infralog/config"
	"infralog/metrics"
	"infralog/persistence"
	"infralog/target"
	"infralog/target/slack"
	"infralog/target/stdout"
	"infralog/target/webhook"
	"infralog/tfstate"
	"infralog/ticker"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := loadConfig()

	targets := initTargets(cfg)
	store := initPersistence(cfg)
	stateBackend := initBackend(cfg)

	initState(stateBackend, store)

	metricsServer := startMetricsServer(cfg)
	ctx := setupSignalHandler()

	runPollingLoop(ctx, cfg, stateBackend, targets, store)

	shutdownMetricsServer(metricsServer)
	fmt.Println("Shutdown complete")
}

// loadConfig parses CLI flags and loads the configuration file.
func loadConfig() *config.Config {
	configFile := flag.String("config-file", "", "Path to configuration file")
	flag.Parse()

	configPath := os.Getenv("INFRALOG_CONFIG_FILE")
	if *configFile != "" {
		configPath = *configFile
	}

	if configPath == "" {
		fmt.Println("Error: config file must be provided via --config-file or INFRALOG_CONFIG_FILE")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	return cfg
}

// initTargets creates notification targets based on configuration.
// Falls back to stdout if no targets are configured.
func initTargets(cfg *config.Config) []target.Target {
	var targets []target.Target

	if cfg.Target.Webhook.URL != "" {
		t, err := webhook.New(cfg.Target.Webhook)
		if err != nil {
			fmt.Printf("Error creating webhook target: %v\n", err)
			os.Exit(1)
		}
		targets = append(targets, t)
	}

	if cfg.Target.Slack.WebhookURL != "" {
		t, err := slack.New(cfg.Target.Slack)
		if err != nil {
			fmt.Printf("Error creating slack target: %v\n", err)
			os.Exit(1)
		}
		targets = append(targets, t)
	}

	if cfg.Target.Stdout.Enabled {
		targets = append(targets, stdout.New(cfg.Target.Stdout))
	}

	if len(targets) == 0 {
		fmt.Println("No targets configured, using stdout as default")
		targets = append(targets, stdout.New(config.StdoutConfig{Enabled: true, Format: "text"}))
	}

	return targets
}

// initPersistence sets up the state persistence store if configured.
func initPersistence(cfg *config.Config) persistence.Store {
	if cfg.Persistence.StateFile == "" {
		return nil
	}

	store, err := persistence.NewFileStore(cfg.Persistence.StateFile)
	if err != nil {
		fmt.Printf("Error creating persistence store: %v\n", err)
		os.Exit(1)
	}

	tfstate.LastState, err = store.Load()
	if err != nil {
		fmt.Printf("Error loading persisted state: %v\n", err)
		os.Exit(1)
	}

	if tfstate.LastState != nil {
		fmt.Println("Loaded persisted state")
	}

	return store
}

// initBackend creates the Terraform state backend (S3 or local).
func initBackend(cfg *config.Config) backend.Backend {
	switch {
	case cfg.TFState.Local.Path != "":
		b := local.New(cfg.TFState.Local)
		fmt.Printf("Using %s backend\n", b.Name())
		return b
	case cfg.TFState.S3.Bucket != "":
		b := s3.New(cfg.TFState.S3)
		fmt.Printf("Using %s backend\n", b.Name())
		return b
	default:
		fmt.Println("Error: no backend configured. Configure either tfstate.s3 or tfstate.local")
		os.Exit(1)
		return nil
	}
}

// initState loads the initial Terraform state if not already loaded from persistence.
func initState(stateBackend backend.Backend, store persistence.Store) {
	if tfstate.LastState != nil {
		return
	}

	stateData, err := stateBackend.GetState()
	if err != nil {
		fmt.Printf("Error getting initial state: %v\n", err)
		os.Exit(1)
	}

	tfstate.LastState, err = tfstate.ParseState(string(stateData))
	if err != nil {
		fmt.Printf("Error parsing initial state: %v\n", err)
		os.Exit(1)
	}

	if store != nil {
		if err := store.Save(tfstate.LastState); err != nil {
			fmt.Printf("Warning: failed to persist initial state: %v\n", err)
		}
	}
}

// startMetricsServer starts the Prometheus metrics server if enabled.
func startMetricsServer(cfg *config.Config) *metrics.Server {
	if !cfg.Metrics.Enabled {
		return nil
	}

	metricsCfg := cfg.Metrics.WithDefaults()
	server := metrics.NewServer(metricsCfg.Address)

	if err := server.Start(); err != nil {
		fmt.Printf("Error starting metrics server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Metrics server started on %s\n", metricsCfg.Address)
	return server
}

// setupSignalHandler creates a context that cancels on SIGINT or SIGTERM.
func setupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
		cancel()
	}()

	return ctx
}

// runPollingLoop starts the ticker and handles state polling.
func runPollingLoop(ctx context.Context, cfg *config.Config, stateBackend backend.Backend, targets []target.Target, store persistence.Store) {
	t := ticker.NewTicker(cfg.Polling.Interval)
	t.Start(ctx, func() {
		handlePoll(cfg, stateBackend, targets, store)
	})
}

// handlePoll fetches current state, compares with last state, and notifies targets.
func handlePoll(cfg *config.Config, stateBackend backend.Backend, targets []target.Target, store persistence.Store) {
	fmt.Println("Polling...")

	currentState, err := fetchAndParseState(stateBackend)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	diff, err := tfstate.Compare(tfstate.LastState, currentState, cfg.Filter)
	if err != nil {
		fmt.Printf("Failed to compare states: %v\n", err)
		metrics.RecordPollError("compare")
		return
	}

	metrics.RecordPollSuccess()

	if !diff.HasChanges() {
		return
	}

	recordChangeMetrics(diff)
	notifyTargets(targets, diff, cfg)

	tfstate.LastState = currentState
	persistState(store, currentState)
}

// fetchAndParseState retrieves and parses the current Terraform state.
func fetchAndParseState(stateBackend backend.Backend) (*tfstate.State, error) {
	stateData, err := stateBackend.GetState()
	if err != nil {
		metrics.RecordPollError("fetch")
		return nil, fmt.Errorf("getting state: %w", err)
	}

	state, err := tfstate.ParseState(string(stateData))
	if err != nil {
		metrics.RecordPollError("parse")
		return nil, fmt.Errorf("parsing state: %w", err)
	}

	return state, nil
}

// recordChangeMetrics records metrics for each resource change.
func recordChangeMetrics(diff *tfstate.StateDiff) {
	for _, rd := range diff.ResourceDiffs {
		metrics.RecordChange(string(rd.Status), rd.ResourceType)
	}
}

// notifyTargets sends the diff to all configured targets.
func notifyTargets(targets []target.Target, diff *tfstate.StateDiff, cfg *config.Config) {
	payload := target.NewPayload(diff, cfg.TFState)

	for _, t := range targets {
		name := getTargetName(t)
		if err := t.Write(payload); err != nil {
			fmt.Printf("Error writing to target: %v\n", err)
			metrics.RecordNotificationError(name)
		} else {
			metrics.RecordNotificationSuccess(name)
		}
	}
}

// persistState saves the current state to disk if persistence is configured.
func persistState(store persistence.Store, state *tfstate.State) {
	if store == nil {
		return
	}
	if err := store.Save(state); err != nil {
		fmt.Printf("Warning: failed to persist state: %v\n", err)
	}
}

// shutdownMetricsServer gracefully stops the metrics server.
func shutdownMetricsServer(server *metrics.Server) {
	if server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Error shutting down metrics server: %v\n", err)
	}
}

// getTargetName returns a string identifier for the target type.
func getTargetName(t target.Target) string {
	switch t.(type) {
	case *webhook.WebhookTarget:
		return "webhook"
	case *slack.SlackTarget:
		return "slack"
	case *stdout.StdoutTarget:
		return "stdout"
	default:
		return "unknown"
	}
}
