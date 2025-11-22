#!/bin/bash

# Terraform State Change Simulator
# This script cycles through different state file versions to simulate infrastructure changes.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
STATE_FILE="$SCRIPT_DIR/terraform.tfstate"
VERSIONS=(state_v1.json state_v2.json state_v3.json state_v4.json)
CURRENT_INDEX=0

echo "Terraform State Change Simulator"
echo "================================="
echo "State file: $STATE_FILE"
echo "Press Enter to cycle to next version, or Ctrl+C to exit"
echo ""

# Initialize with version 1
cp "$SCRIPT_DIR/${VERSIONS[$CURRENT_INDEX]}" "$STATE_FILE"
echo "[$(date '+%H:%M:%S')] Initialized with ${VERSIONS[$CURRENT_INDEX]}"

while true; do
    read -r -p "Press Enter for next change... "

    # Move to next version (cycle back to 0 after reaching the end)
    CURRENT_INDEX=$(( (CURRENT_INDEX + 1) % ${#VERSIONS[@]} ))

    cp "$SCRIPT_DIR/${VERSIONS[$CURRENT_INDEX]}" "$STATE_FILE"
    echo "[$(date '+%H:%M:%S')] Changed to ${VERSIONS[$CURRENT_INDEX]}"

    # Show what changed
    case $CURRENT_INDEX in
        0)
            echo "  -> Reset to initial state (v1)"
            ;;
        1)
            echo "  -> aws_instance.web_server: instance_type t2.micro -> t2.small"
            ;;
        2)
            echo "  -> output.api_endpoint changed"
            echo "  -> aws_lambda_function.api_handler added"
            ;;
        3)
            echo "  -> aws_s3_bucket.data_bucket removed"
            echo "  -> aws_instance.web_server: instance_type t2.small -> t2.medium, ami changed"
            echo "  -> aws_rds_cluster.database: engine_version 14.6 -> 15.2"
            echo "  -> aws_lambda_function.api_handler: memory_size 256 -> 512"
            echo "  -> output.instance_ip changed"
            echo "  -> output.lambda_arn added"
            ;;
    esac
    echo ""
done
