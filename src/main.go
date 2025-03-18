package main

import (
	"flag"
	"fmt"
	"infralog/backend/s3"
	"infralog/config"
	"infralog/target"
	"infralog/target/webhook"
	"infralog/tfstate"
	"infralog/ticker"
	"os"
)

func main() {
	configFile := flag.String("config-file", "", "Path to configuration file")
	flag.Parse()

	if *configFile == "" {
		fmt.Println("Error: --config-file parameter is required")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	targets := []target.Target{}
	if cfg.Target.Webhook.URL != "" {
		targets = append(targets, webhook.New(cfg.Target.Webhook.URL))
	}

	initialStateData, err := s3.GetObject(cfg.TFState.S3.Bucket, cfg.TFState.S3.Key, cfg.TFState.S3.Region)
	if err != nil {
		fmt.Printf("Error getting initial state: %v\n", err)
		os.Exit(1)
	}

	tfstate.LastState, err = tfstate.ParseState(string(initialStateData))
	if err != nil {
		fmt.Printf("Error parsing initial state: %v\n", err)
		os.Exit(1)
	}

	t := ticker.NewTicker(cfg.Polling.Interval)
	t.Start(func() {
		diff, err := compareStates(cfg.TFState.S3.Bucket, cfg.TFState.S3.Key, cfg.TFState.S3.Region)
		if err != nil {
			fmt.Printf("Error comparing states: %v\n", err)
			return
		}

		// TODO: Only write to targets if there are changes
		for _, t := range targets {
			if err := t.Write(diff); err != nil {
				fmt.Printf("Error writing to target: %v\n", err)
			}
		}

		fmt.Println(diff)
	})
}

func compareStates(bucket, key, region string) (*tfstate.StateDiff, error) {
	currentStateData, err := s3.GetObject(bucket, key, region)
	if err != nil {
		return nil, fmt.Errorf("failed to get current state: %w", err)
	}

	currentState, err := tfstate.ParseState(string(currentStateData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse current state: %w", err)
	}

	diff, err := tfstate.Compare(tfstate.LastState, currentState)
	if err != nil {
		return nil, fmt.Errorf("failed to compare states: %w", err)
	}

	return diff, nil
}
