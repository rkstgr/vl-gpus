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

CREATE TABLE IF NOT EXISTS instances
(
    `instance_id` String,
    `instance_ipv4` IPv4,
    `startup_id` UInt32,
    `assigned_date` DateTime('UTC'),
    `unassigned_date` Nullable(DateTime('UTC')),
    `num_gpus_assigned` UInt8,
    `instance_status` Enum8('Requested' = 1, 'Provisioning' = 2, 'Active' = 3, 'Decommissioning' = 4, 'Idle' = 5, 'Error' = 6),
    `is_provisioned` BOOLEAN DEFAULT FALSE,
    `last_status_update` DateTime('UTC') DEFAULT now()
)
ENGINE = ReplacingMergeTree(last_status_update)
ORDER BY instance_id;

INSERT INTO startups (id, startup_name, contact_email, onboarding_date, public_ssh_key, notes) VALUES
(101, 'AI Innovators GmbH', 'contact@aiinnovators.com', '2024-03-15', 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC+... ai_innovators_key', 'Focus on generative AI for content creation.'),
(102, 'DeepMind Solutions', 'info@deepmindsolutions.co', '2024-04-01', 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD+... deep_mind_key', 'Specializing in reinforcement learning for robotics.'),
(103, 'NeuroFlow Analytics', 'hello@neuroflow.ai', '2024-05-10', 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD+... neuro_flow_key', 'Developing neural networks for financial market prediction.');
