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

	configPath := os.Getenv("INFRALOG_CONFIG_FILE")
	if *configFile != "" {
		configPath = *configFile
	}

	if configPath == "" {
		fmt.Println("Error: config file must be provided either via --config-file parameter or INFRALOG_CONFIG_FILE environment variable")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	targets := []target.Target{}
	if cfg.Target.Webhook.URL != "" {
		webhookTarget, err := webhook.New(cfg.Target.Webhook.URL, cfg.Target.Webhook.Method)
		if err != nil {
			fmt.Printf("Error creating webhook target: %v\n", err)
			os.Exit(1)
		}
		targets = append(targets, webhookTarget)
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
		fmt.Println("Polling...")

		currentStateData, err := s3.GetObject(cfg.TFState.S3.Bucket, cfg.TFState.S3.Key, cfg.TFState.S3.Region)
		if err != nil {
			fmt.Printf("Error getting state: %v\n", err)
			return
		}

		currentState, err := tfstate.ParseState(string(currentStateData))
		if err != nil {
			fmt.Printf("Error parsing state: %v\n", err)
			return
		}

		diff, err := tfstate.Compare(tfstate.LastState, currentState, cfg.Filter)
		if err != nil {
			fmt.Printf("failed to compare states: %v", err)
			return
		}

		if !diff.HasChanges() {
			return
		}

		for _, t := range targets {
			if err := t.Write(diff, cfg.TFState); err != nil {
				fmt.Printf("Error writing to target: %v\n", err)
			}
		}

		tfstate.LastState = currentState
	})
}
