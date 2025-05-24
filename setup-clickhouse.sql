CREATE DATABASE IF NOT EXISTS "vlgpus";
USE `vlgpus`;

CREATE TABLE IF NOT EXISTS gpu_metrics
(
    `timestamp` DateTime('UTC'),
    `instance_id` String,
    `gpu_index` UInt8,
    `gpu_utilization_percent` UInt8,
    `gpu_memory_used_mb` UInt32,
    `gpu_memory_total_mb` UInt32,
    `temperature_celsius` UInt8 DEFAULT 0,  -- Optional
    `power_draw_watts` UInt16 DEFAULT 0     -- Optional
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)  -- Good for querying data by month
ORDER BY (instance_id, gpu_index, timestamp) -- Optimal for queries filtering by instance/GPU and time range
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS startups
(
    `id` UInt32,
    `startup_name` String,
    `contact_email` String,
    `onboarding_date` Date,
    `offboarding_date` Nullable(Date),
    `public_ssh_key` String,
    `notes` String DEFAULT '',
    `created_at` DateTime('UTC') DEFAULT now(),
    `updated_at` DateTime('UTC') DEFAULT now()
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY id;

CREATE TABLE IF NOT EXISTS vlgpus.instances
(
    `instance_id` String,
    `instance_ipv4` IPv4,
    `configuration` String,
    `api_key` String DEFAULT '',
    `startup_id` Nullable(UInt32),
    `assigned_at` Nullable(DateTime('UTC')),
    `is_provisioned` BOOLEAN DEFAULT FALSE,
    `updated_at` DateTime('UTC') DEFAULT now()
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY instance_id;

INSERT INTO vlgpus.instances
(instance_id, instance_ipv4, configuration, api_key, startup_id, is_provisioned)
VALUES
('test-vm-001', '127.0.0.1', 'test-config', 'test-api-key-12345', NULL, true);