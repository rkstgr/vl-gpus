package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gorilla/mux"

	"vl-gpus/pkg/config"
	"vl-gpus/pkg/types"
)

type Server struct {
	db clickhouse.Conn
}

func NewServer(cfg config.ServerConfig) (*Server, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.ClickHouseAddr},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouseDB,
			Username: cfg.ClickHouseUser,
			Password: cfg.ClickHousePass,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &Server{db: conn}, nil
}

func (s *Server) authenticateInstance(apiKey string) (string, error) {
	var instanceID string
	query := `SELECT instance_id FROM vlgpus.instances WHERE api_key = ? AND is_provisioned = true`

	err := s.db.QueryRow(context.Background(), query, apiKey).Scan(&instanceID)
	if err != nil {
		return "", fmt.Errorf("invalid or inactive instance: %w", err)
	}

	return instanceID, nil
}

func (s *Server) insertMetrics(ctx context.Context, instanceID string, timestamp time.Time, gpus []types.GPUMetric) error {
	batch, err := s.db.PrepareBatch(ctx, `
		INSERT INTO vlgpus.gpu_metrics (
			timestamp, instance_id, gpu_index, gpu_utilization_percent,
			gpu_memory_used_mb, gpu_memory_total_mb, temperature_celsius, power_draw_watts
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, gpu := range gpus {
		err := batch.Append(
			timestamp,
			instanceID,
			uint8(gpu.Index),
			uint8(gpu.UtilizationPercent),
			uint32(gpu.MemoryUsedMB),
			uint32(gpu.MemoryTotalMB),
			uint8(gpu.TemperatureCelsius),
			uint16(gpu.PowerDrawWatts),
		)
		if err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	return batch.Send()
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Extract API key from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Expected format: "Bearer <api_key>"
	var apiKey string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		apiKey = authHeader[7:]
	} else {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	// Authenticate the instance
	instanceID, err := s.authenticateInstance(apiKey)
	if err != nil {
		log.Printf("Authentication failed: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the request body
	var payload types.MetricsPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate that the instance_id in payload matches authenticated instance
	if payload.InstanceID != instanceID {
		log.Printf("Instance ID mismatch: authenticated=%s, payload=%s", instanceID, payload.InstanceID)
		http.Error(w, "Instance ID mismatch", http.StatusForbidden)
		return
	}

	// Use current time if timestamp is zero
	if payload.Timestamp.IsZero() {
		payload.Timestamp = time.Now().UTC()
	}

	// Insert metrics into database
	ctx := context.Background()
	if err := s.insertMetrics(ctx, instanceID, payload.Timestamp, payload.GPUs); err != nil {
		log.Printf("Failed to insert metrics: %v", err)
		http.Error(w, "Failed to store metrics", http.StatusInternalServerError)
		return
	}

	log.Printf("Stored %d GPU metrics for instance %s", len(payload.GPUs), instanceID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Simple health check - ping ClickHouse
	if err := s.db.Ping(context.Background()); err != nil {
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}


func main() {
	cfg := config.LoadServerConfig()

	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer server.db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/metrics", server.handleMetrics).Methods("POST")
	r.HandleFunc("/health", server.handleHealth).Methods("GET")

	log.Printf("Starting metrics server on port %s", cfg.Port)
	log.Printf("ClickHouse connection: %s/%s", cfg.ClickHouseAddr, cfg.ClickHouseDB)

	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
