package config

import (
	"os"

	"slices"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Polling struct {
		Interval int `yaml:"interval"`
	} `yaml:"polling"`
	TFState struct {
		S3 struct {
			Bucket string `yaml:"bucket"`
			Key    string `yaml:"key"`
			Region string `yaml:"region"`
		} `yaml:"s3"`
	} `yaml:"tfstate"`
	Target struct {
		Webhook struct {
			URL string `yaml:"url"`
		} `yaml:"webhook"`
	} `yaml:"target"`
	Filter Filter `yaml:"filter"`
}

type Filter struct {
	ResourceTypes []string `yaml:"resource_types"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (f *Filter) MatchesResourceType(resourceType string) bool {
	if f.ResourceTypes == nil {
		return true
	}
	if len(f.ResourceTypes) == 0 {
		return false
	}
	return slices.Contains(f.ResourceTypes, resourceType)
}
