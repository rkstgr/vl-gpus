#!/bin/bash

set -e

INSTANCE_ID="$1"
INSTANCE_IP="$2"

if [ -z "$INSTANCE_ID" ] || [ -z "$INSTANCE_IP" ]; then
    echo "Usage: $0 <instance_id> <instance_ip>"
    echo "Example: $0 gpu-node-001 192.168.1.100"
    exit 1
fi

# Generate API key (UUID)
API_KEY=$(uuidgen | tr '[:upper:]' '[:lower:]')

echo "Provisioning instance: $INSTANCE_ID"
echo "IP Address: $INSTANCE_IP"
echo "Generated API Key: $API_KEY"

# Check if instance already exists and get existing API key if available
echo "Checking if instance exists in database..."
EXISTING_API_KEY=$(./clickhouse client --host axe --user=admin --password=$CLICKHOUSE_PASSWORD --query="
SELECT api_key FROM vlgpus.instances WHERE instance_id = '$INSTANCE_ID' LIMIT 1
" 2>/dev/null || echo "")

if [ -n "$EXISTING_API_KEY" ] && [ "$EXISTING_API_KEY" != "" ]; then
    echo "Instance $INSTANCE_ID already exists in database"
    echo "Using existing API key: $EXISTING_API_KEY"
    API_KEY="$EXISTING_API_KEY"

    # Update existing instance
    echo "Updating existing database entry..."
    ./clickhouse client --host axe --user=admin --password=$CLICKHOUSE_PASSWORD --query="
    ALTER TABLE vlgpus.instances UPDATE
        instance_ipv4 = '$INSTANCE_IP',
        configuration = 'standard-gpu',
        is_provisioned = true
    WHERE instance_id = '$INSTANCE_ID'
    "
else
    echo "Creating new database entry..."
    ./clickhouse client --host axe --user=admin --password=$CLICKHOUSE_PASSWORD --query="
    INSERT INTO vlgpus.instances
    (instance_id, instance_ipv4, configuration, api_key, startup_id, is_provisioned, updated_at)
    VALUES
    ('$INSTANCE_ID', '$INSTANCE_IP', 'standard-gpu', '$API_KEY', NULL, true, now())
    "
fi

if [ $? -ne 0 ]; then
    echo "Error: Failed to create database entry"
    exit 1
fi

echo "Database entry created successfully"

# Create temporary inventory for this instance
TEMP_INVENTORY=$(mktemp).yml
cat > "$TEMP_INVENTORY" << EOF
gpu_instances:
  hosts:
    $INSTANCE_ID:
      ansible_host: $INSTANCE_IP
      ansible_user: ubuntu
      ansible_ssh_private_key_file: $ADMIN_SSH_KEY_FILE
  vars:
    instance_id: "$INSTANCE_ID"
    api_key: "$API_KEY"
    metrics_server_url: $METRICS_SERVER_URL
    collect_interval: $METRICS_COLLECT_INTERVAL_SEC
EOF

echo "Deploying collector using Ansible..."
ansible-playbook -i "$TEMP_INVENTORY" deploy/deploy-gpu-collector.yml --limit "$INSTANCE_ID"

if [ $? -eq 0 ]; then
    echo "✅ Provisioning completed successfully!"
    echo "Instance ID: $INSTANCE_ID"
    echo "IP Address: $INSTANCE_IP"
    echo "API Key: $API_KEY"
    echo "Collector deployed and running"
else
    echo "❌ Ansible deployment failed"
    echo "Rolling back database entry..."
    ./clickhouse client --host axe --user=admin --password=$CLICKHOUSE_PASSWORD --query="
    ALTER TABLE vlgpus.instances UPDATE is_provisioned = false WHERE instance_id = '$INSTANCE_ID'
    "
    exit 1
fi

# Cleanup
rm -f "$TEMP_INVENTORY"

echo "Provisioning complete for $INSTANCE_ID"
