#!/bin/bash

# Fake nvidia-smi for testing
# This simulates the output of: nvidia-smi --query-gpu=index,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw --format=csv,noheader,nounits

# Check if we're being called with the right arguments
if [[ "$*" == *"--query-gpu=index,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw"* ]]; then
    # Simulate 2 GPUs with varying metrics
    GPU0_UTIL=$((50 + RANDOM % 50))  # 50-99% utilization
    GPU1_UTIL=$((20 + RANDOM % 60))  # 20-79% utilization

    GPU0_MEM_USED=$((8000 + RANDOM % 4000))  # 8-12GB used
    GPU1_MEM_USED=$((6000 + RANDOM % 6000))  # 6-12GB used

    GPU0_TEMP=$((65 + RANDOM % 20))   # 65-84°C
    GPU1_TEMP=$((60 + RANDOM % 25))   # 60-84°C

    GPU0_POWER=$((250 + RANDOM % 100)) # 250-349W
    GPU1_POWER=$((200 + RANDOM % 150)) # 200-349W

    echo "0, $GPU0_UTIL, $GPU0_MEM_USED, 24576, $GPU0_TEMP, $GPU0_POWER"
    echo "1, $GPU1_UTIL, $GPU1_MEM_USED, 24576, $GPU1_TEMP, $GPU1_POWER"
else
    # For any other nvidia-smi call, just return some basic info
    echo "NVIDIA-SMI 535.104.12    Driver Version: 535.104.12    CUDA Version: 12.2"
fi
