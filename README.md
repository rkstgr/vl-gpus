# VL GPUs - GPU Metrics Collection System

A distributed system for collecting and storing GPU metrics from compute instances.

## Architecture

- **Collector** (`cmd/collector/`): Runs on GPU instances to collect metrics via nvidia-smi
- **Server** (`cmd/server/`): Receives metrics from collectors and stores in ClickHouse
- **Shared Types** (`pkg/types/`): Common data structures
- **Configuration** (`pkg/config/`): Configuration management

## Quick Start

### Build Collector
```bash
# For local testing (macOS)
go build -o collector cmd/collector/main.go

# For production (Linux)
GOOS=linux GOARCH=amd64 go build -o collector cmd/collector/main.go
```

### Build Server
```bash
go build -o server cmd/server/main.go
```

### Run Server
Set environment variables:
```bash
export CLICKHOUSE_USER=your_user
export CLICKHOUSE_PASS=your_password
```

Then run:
```bash
./server
```

### Deploy Collector
Use the deployment configuration in `deploy/` directory.

## Testing

For local testing without nvidia-smi, use the fake script:
```bash
chmod +x scripts/fake-nvidia-smi
cp scripts/fake-nvidia-smi ~/.local/bin/nvidia-smi
```

## Database Setup

Run the SQL setup script:
```bash
clickhouse-client < setup-clickhouse.sql
```