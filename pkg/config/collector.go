package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type CollectorConfig struct {
	InstanceID      string        `json:"instance_id"`
	APIKey          string        `json:"api_key"`
	MetricsURL      string        `json:"metrics_url"`
	CollectInterval time.Duration `json:"collect_interval_seconds"`
}

func LoadCollectorConfig() (CollectorConfig, error) {
	configPath := "/etc/gpu-metrics/config.json"

	if path := os.Getenv("GPU_METRICS_CONFIG"); path != "" {
		configPath = path
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return CollectorConfig{}, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config CollectorConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return CollectorConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	if config.CollectInterval > 0 {
		config.CollectInterval = time.Duration(config.CollectInterval) * time.Second
	} else {
		config.CollectInterval = 60 * time.Second
	}

	if config.InstanceID == "" {
		return CollectorConfig{}, fmt.Errorf("instance_id is required")
	}
	if config.APIKey == "" {
		return CollectorConfig{}, fmt.Errorf("api_key is required")
	}
	if config.MetricsURL == "" {
		return CollectorConfig{}, fmt.Errorf("metrics_url is required")
	}

	return config, nil
}