# VL GPUs - GPU Metrics Collection System

A distributed system for collecting and storing GPU metrics from compute instances.

## Architecture

- **Collector** (`cmd/collector/`): Runs on GPU instances to collect metrics via nvidia-smi
- **Server** (`cmd/server/`): Receives metrics from collectors and stores in ClickHouse
- **Shared Types** (`pkg/types/`): Common data structures
- **Configuration** (`pkg/config/`): Configuration management
- **Provisioning** (`provision.sh`): Script to provision new instances
- **Deployment** (`deploy/`): Containing ansible deployment playbook

## Run provisioning
Build collector:
```sh
GOOS=linux GOARCH=amd64 go build -o deploy/gpu-metrics-collector cmd/collector/main.go
```

Create env variables (.envrc):
```sh
export CLICKHOUSE_HOST=your_host
export CLICKHOUSE_USER=your_user
export CLICKHOUSE_PASSWORD=your_password

export ADMIN_SSH_KEY_FILE=~/.ssh/admin-key
export METRICS_SERVER_URL=https://metrics-server.com/metrics
export METRICS_COLLECT_INTERVAL_SEC=60
```

Make sure you have the clickhouse client installed, download it with `curl https://clickhouse.com/ | sh`

Finally, run the provisioning script:
```sh
./provision.sh instance_id ip_address startup_ssh_key
# example
./provision.sh radiant-pasteur 38.128.233.116 "ssh-ed25519 AAAAC3N..."
```

## Server
build
```bash
go build -o server cmd/server/main.go
```

set environment variables:
```bash
export CLICKHOUSE_USER=your_user
export CLICKHOUSE_PASS=your_password
```

Then run:
```bash
./server
```

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
