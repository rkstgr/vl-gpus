package types

import "time"

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