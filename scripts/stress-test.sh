#!/bin/bash

# Containr Stress Test Script
# This script tests the container runtime under heavy load by creating and managing 100+ containers

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NUM_CONTAINERS=${1:-100}
CONTAINR_BIN=${CONTAINR_BIN:-./bin/containr}
TEST_IMAGE=${TEST_IMAGE:-alpine}
LOG_FILE="stress-test-$(date +%Y%m%d-%H%M%S).log"
RESULTS_FILE="stress-test-results-$(date +%Y%m%d-%H%M%S).json"

# Metrics
TOTAL_CREATED=0
TOTAL_FAILED=0
TOTAL_RUNNING=0
START_TIME=$(date +%s)

echo -e "${GREEN}=== Containr Stress Test ===${NC}"
echo "Testing with $NUM_CONTAINERS containers"
echo "Log file: $LOG_FILE"
echo

# Check if containr binary exists
if [ ! -f "$CONTAINR_BIN" ]; then
    echo -e "${RED}Error: containr binary not found at $CONTAINR_BIN${NC}"
    echo "Please build containr first with 'make build'"
    exit 1
fi

# Check for root privileges
if [ "$EUID" -ne 0 ]; then
    echo -e "${YELLOW}Warning: This script requires root privileges${NC}"
    echo "Please run with sudo"
    exit 1
fi

# Initialize log
echo "Stress Test Started: $(date)" > "$LOG_FILE"
echo "Configuration: NUM_CONTAINERS=$NUM_CONTAINERS, IMAGE=$TEST_IMAGE" >> "$LOG_FILE"
echo "---" >> "$LOG_FILE"

# Function to create a container
create_container() {
    local id=$1
    local name="stress-test-$id"

    echo -n "Creating container $id/$NUM_CONTAINERS... "

    if timeout 30s "$CONTAINR_BIN" run --name "$name" --rm -d "$TEST_IMAGE" sleep 60 >> "$LOG_FILE" 2>&1; then
        echo -e "${GREEN}OK${NC}"
        ((TOTAL_CREATED++))
        return 0
    else
        echo -e "${RED}FAILED${NC}"
        ((TOTAL_FAILED++))
        return 1
    fi
}

# Function to monitor system resources
monitor_resources() {
    echo "=== System Resource Usage ===" >> "$LOG_FILE"
    echo "Timestamp: $(date)" >> "$LOG_FILE"

    # CPU usage
    echo "CPU Usage:" >> "$LOG_FILE"
    top -bn1 | grep "Cpu(s)" >> "$LOG_FILE"

    # Memory usage
    echo "Memory Usage:" >> "$LOG_FILE"
    free -h >> "$LOG_FILE"

    # Disk usage
    echo "Disk Usage:" >> "$LOG_FILE"
    df -h / >> "$LOG_FILE"

    # Container count
    RUNNING_COUNT=$(ps aux | grep -c "[c]ontainr" || true)
    echo "Running containr processes: $RUNNING_COUNT" >> "$LOG_FILE"

    # cgroup usage
    if [ -d "/sys/fs/cgroup" ]; then
        echo "Cgroup subdirs: $(find /sys/fs/cgroup -mindepth 1 -maxdepth 2 -type d 2>/dev/null | wc -l)" >> "$LOG_FILE"
    fi

    echo "---" >> "$LOG_FILE"
}

# Phase 1: Sequential Creation
echo -e "\n${YELLOW}Phase 1: Creating $NUM_CONTAINERS containers sequentially${NC}"
PHASE1_START=$(date +%s)

for i in $(seq 1 $NUM_CONTAINERS); do
    create_container $i

    # Monitor every 10 containers
    if [ $((i % 10)) -eq 0 ]; then
        monitor_resources
        echo "Progress: $i/$NUM_CONTAINERS created"
    fi

    # Small delay to avoid overwhelming the system
    sleep 0.1
done

PHASE1_END=$(date +%s)
PHASE1_DURATION=$((PHASE1_END - PHASE1_START))

echo -e "\nPhase 1 Complete:"
echo "  Created: $TOTAL_CREATED"
echo "  Failed: $TOTAL_FAILED"
echo "  Duration: ${PHASE1_DURATION}s"
echo "  Rate: $(echo "scale=2; $TOTAL_CREATED / $PHASE1_DURATION" | bc) containers/sec"

# Phase 2: Resource Monitoring
echo -e "\n${YELLOW}Phase 2: Monitoring resource usage under load${NC}"

for i in {1..10}; do
    echo "Monitoring iteration $i/10..."
    monitor_resources
    sleep 5
done

# Phase 3: Container Cleanup Test
echo -e "\n${YELLOW}Phase 3: Testing cleanup (waiting for containers to exit)${NC}"

echo "Waiting for containers to exit naturally..."
sleep 65  # Containers were started with sleep 60

# Check for any remaining processes
REMAINING=$(ps aux | grep -c "[s]tress-test-" || true)
echo "Remaining container processes: $REMAINING"

if [ $REMAINING -gt 0 ]; then
    echo -e "${YELLOW}Warning: Some containers still running. Attempting cleanup...${NC}"
    pkill -f "stress-test-" || true
    sleep 5
fi

# Phase 4: Rapid Creation/Deletion
echo -e "\n${YELLOW}Phase 4: Rapid creation and deletion test${NC}"

RAPID_TEST_COUNT=20
echo "Creating and immediately stopping $RAPID_TEST_COUNT containers..."

RAPID_START=$(date +%s)
for i in $(seq 1 $RAPID_TEST_COUNT); do
    local name="rapid-test-$i"
    "$CONTAINR_BIN" run --name "$name" "$TEST_IMAGE" echo "test" >> "$LOG_FILE" 2>&1 || true
done
RAPID_END=$(date +%s)
RAPID_DURATION=$((RAPID_END - RAPID_START))

echo "Rapid test complete: $RAPID_TEST_COUNT containers in ${RAPID_DURATION}s"

# Final resource check
echo -e "\n${YELLOW}Final System State:${NC}"
monitor_resources

# Check for resource leaks
echo -e "\n${YELLOW}Checking for resource leaks...${NC}"

# Check for leftover cgroups
LEFTOVER_CGROUPS=$(find /sys/fs/cgroup -name "*stress-test*" -o -name "*rapid-test*" 2>/dev/null | wc -l)
echo "Leftover cgroups: $LEFTOVER_CGROUPS"

# Check for leftover namespaces
LEFTOVER_NS=$(find /proc -maxdepth 1 -name "[0-9]*" -exec ls {}/ns 2>/dev/null \; | wc -l)
echo "Total namespace count: $LEFTOVER_NS"

# Check for leftover network interfaces
LEFTOVER_VETH=$(ip link | grep -c "veth" || true)
echo "Leftover veth interfaces: $LEFTOVER_VETH"

# Generate Results
END_TIME=$(date +%s)
TOTAL_DURATION=$((END_TIME - START_TIME))

echo -e "\n${GREEN}=== Test Complete ===${NC}"
echo "Total Duration: ${TOTAL_DURATION}s"
echo "Containers Created: $TOTAL_CREATED"
echo "Containers Failed: $TOTAL_FAILED"
echo "Success Rate: $(echo "scale=2; $TOTAL_CREATED * 100 / $NUM_CONTAINERS" | bc)%"

# Generate JSON results
cat > "$RESULTS_FILE" <<EOF
{
  "test_name": "Containr Stress Test",
  "timestamp": "$(date -Iseconds)",
  "configuration": {
    "num_containers": $NUM_CONTAINERS,
    "test_image": "$TEST_IMAGE",
    "containr_bin": "$CONTAINR_BIN"
  },
  "results": {
    "total_duration_seconds": $TOTAL_DURATION,
    "containers_created": $TOTAL_CREATED,
    "containers_failed": $TOTAL_FAILED,
    "success_rate_percent": $(echo "scale=2; $TOTAL_CREATED * 100 / $NUM_CONTAINERS" | bc),
    "creation_rate_per_second": $(echo "scale=2; $TOTAL_CREATED / $PHASE1_DURATION" | bc)
  },
  "phases": {
    "sequential_creation": {
      "duration_seconds": $PHASE1_DURATION,
      "containers_created": $TOTAL_CREATED,
      "rate_per_second": $(echo "scale=2; $TOTAL_CREATED / $PHASE1_DURATION" | bc)
    },
    "rapid_test": {
      "duration_seconds": $RAPID_DURATION,
      "containers_tested": $RAPID_TEST_COUNT
    }
  },
  "resource_leaks": {
    "leftover_cgroups": $LEFTOVER_CGROUPS,
    "leftover_veth_interfaces": $LEFTOVER_VETH
  },
  "log_file": "$LOG_FILE"
}
EOF

echo
echo "Results saved to: $RESULTS_FILE"
echo "Detailed logs saved to: $LOG_FILE"

# Summary
if [ $LEFTOVER_CGROUPS -gt 10 ] || [ $LEFTOVER_VETH -gt 10 ]; then
    echo -e "\n${RED}WARNING: Potential resource leaks detected!${NC}"
    echo "Please review the logs and system state."
    exit 1
fi

if [ $TOTAL_FAILED -gt $((NUM_CONTAINERS / 10)) ]; then
    echo -e "\n${RED}WARNING: High failure rate detected!${NC}"
    echo "More than 10% of containers failed to create."
    exit 1
fi

echo -e "\n${GREEN}Stress test passed successfully!${NC}"
exit 0
