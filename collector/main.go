package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	InstanceID      string        `json:"instance_id"`
	APIKey          string        `json:"api_key"`
	MetricsURL      string        `json:"metrics_url"`
	CollectInterval time.Duration `json:"collect_interval_seconds"`
}

type GPUMetric struct {
	Index              int `json:"gpu_index"`
	UtilizationPercent int `json:"gpu_utilization_percent"`
	MemoryUsedMB       int `json:"gpu_memory_used_mb"`
	MemoryTotalMB      int `json:"gpu_memory_total_mb"`
	TemperatureCelsius int `json:"temperature_celsius,omitempty"`
	PowerDrawWatts     int `json:"power_draw_watts,omitempty"`
}

type MetricsPayload struct {
	InstanceID string      `json:"instance_id"`
	Timestamp  time.Time   `json:"timestamp"`
	GPUs       []GPUMetric `json:"gpus"`
}

type Collector struct {
	config Config
	client *http.Client
}

func NewCollector(config Config) *Collector {
	return &Collector{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Collector) collectGPUMetrics() ([]GPUMetric, error) {
	// Query nvidia-smi for GPU metrics
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw", "--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run nvidia-smi: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	metrics := make([]GPUMetric, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ", ")
		if len(fields) < 4 {
			log.Printf("Warning: unexpected nvidia-smi output format: %s", line)
			continue
		}

		metric := GPUMetric{}

		// Parse required fields
		if index, err := strconv.Atoi(strings.TrimSpace(fields[0])); err == nil {
			metric.Index = index
		} else {
			log.Printf("Warning: invalid GPU index: %s", fields[0])
			continue
		}

		if util, err := strconv.Atoi(strings.TrimSpace(fields[1])); err == nil {
			metric.UtilizationPercent = util
		}

		if memUsed, err := strconv.Atoi(strings.TrimSpace(fields[2])); err == nil {
			metric.MemoryUsedMB = memUsed
		}

		if memTotal, err := strconv.Atoi(strings.TrimSpace(fields[3])); err == nil {
			metric.MemoryTotalMB = memTotal
		}

		// Parse optional fields (temperature and power might not be available on all GPUs)
		if len(fields) > 4 && fields[4] != "N/A" && strings.TrimSpace(fields[4]) != "" {
			if temp, err := strconv.Atoi(strings.TrimSpace(fields[4])); err == nil {
				metric.TemperatureCelsius = temp
			}
		}

		if len(fields) > 5 && fields[5] != "N/A" && strings.TrimSpace(fields[5]) != "" {
			// Power might come with decimals, so parse as float first
			if powerStr := strings.TrimSpace(fields[5]); powerStr != "" {
				if power, err := strconv.ParseFloat(powerStr, 64); err == nil {
					metric.PowerDrawWatts = int(power)
				}
			}
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (c *Collector) sendMetrics(metrics []GPUMetric) error {
	payload := MetricsPayload{
		InstanceID: c.config.InstanceID,
		Timestamp:  time.Now().UTC(),
		GPUs:       metrics,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.MetricsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("metrics server returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *Collector) Run() {
	log.Printf("Starting GPU metrics collector for instance %s", c.config.InstanceID)
	log.Printf("Collecting metrics every %v", c.config.CollectInterval)
	log.Printf("Sending to: %s", c.config.MetricsURL)

	ticker := time.NewTicker(c.config.CollectInterval)
	defer ticker.Stop()

	// Collect metrics immediately on start
	c.collectAndSend()

	// Then collect on the scheduled interval
	for range ticker.C {
		c.collectAndSend()
	}
}

func (c *Collector) collectAndSend() {
	metrics, err := c.collectGPUMetrics()
	if err != nil {
		log.Printf("Error collecting GPU metrics: %v", err)
		return
	}

	log.Printf("Collected metrics for %d GPUs", len(metrics))

	if err := c.sendMetrics(metrics); err != nil {
		log.Printf("Error sending metrics: %v", err)
		return
	}

	log.Printf("Successfully sent metrics for %d GPUs", len(metrics))
}

func loadConfig() (Config, error) {
	configPath := "/etc/gpu-metrics/config.json"

	// Allow override via environment variable for testing
	if path := os.Getenv("GPU_METRICS_CONFIG"); path != "" {
		configPath = path
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}

	// Convert seconds to duration
	if config.CollectInterval > 0 {
		config.CollectInterval = time.Duration(config.CollectInterval) * time.Second
	} else {
		config.CollectInterval = 60 * time.Second // Default to 60 seconds
	}

	// Validate required fields
	if config.InstanceID == "" {
		return Config{}, fmt.Errorf("instance_id is required")
	}
	if config.APIKey == "" {
		return Config{}, fmt.Errorf("api_key is required")
	}
	if config.MetricsURL == "" {
		return Config{}, fmt.Errorf("metrics_url is required")
	}

	return config, nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	collector := NewCollector(config)
	collector.Run()
}
