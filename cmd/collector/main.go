package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"vl-gpus/pkg/config"
	"vl-gpus/pkg/types"
)

type Collector struct {
	config config.CollectorConfig
	client *http.Client
}

func NewCollector(cfg config.CollectorConfig) *Collector {
	return &Collector{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Collector) collectGPUMetrics() ([]types.GPUMetric, error) {
	// Query nvidia-smi for GPU metrics
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw", "--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run nvidia-smi: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	metrics := make([]types.GPUMetric, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ", ")
		if len(fields) < 4 {
			log.Printf("Warning: unexpected nvidia-smi output format: %s", line)
			continue
		}

		metric := types.GPUMetric{}

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

func (c *Collector) sendMetrics(metrics []types.GPUMetric) error {
	payload := types.MetricsPayload{
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


func main() {
	cfg, err := config.LoadCollectorConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	collector := NewCollector(cfg)
	collector.Run()
}
