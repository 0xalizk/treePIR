#!/bin/bash

# Script to benchmark PIR scheme on a 1GB database
# Usage: ./benchmark.sh [pirType] [rowLen]
#   pirType: PIR type to use (default: TreePIR)
#   rowLen:  Row length in bytes (default: 1024)
#
# Examples:
#   ./benchmark.sh TreePIR 1024
#   ./benchmark.sh Punc 4096
#   ./benchmark.sh DPF 32

set -e

# Configuration
GB_SIZE=1073741824  # 1GB in bytes
PIR_TYPE=${1:-TreePIR}
ROW_LEN=${2:-1024}

# Calculate number of rows for 1GB database
NUM_ROWS=$((GB_SIZE / ROW_LEN))

# Verify the calculation
TOTAL_SIZE=$((NUM_ROWS * ROW_LEN))

# Check for unusual configurations
if [ $NUM_ROWS -lt 16 ]; then
    echo "WARNING: Only $NUM_ROWS rows with rowLen=$ROW_LEN bytes"
    echo "TreePIR requires at least 16 rows. Consider using a smaller row length."
    echo "Suggested row lengths: 32, 1024, 4096, 16384"
    echo ""
fi

echo "=========================================="
echo "Benchmark Configuration:"
echo "  PIR Type:     $PIR_TYPE"
echo "  Row Length:   $ROW_LEN bytes ($(echo "scale=2; $ROW_LEN/1024/1024" | bc)MB)"
echo "  Number Rows:  $NUM_ROWS"
echo "  Total Size:   $TOTAL_SIZE bytes (~1GB)"
echo "=========================================="
echo ""

# Build the benchmark if needed
BENCHMARK_BIN="./benchmark_initial"
if [ ! -f "$BENCHMARK_BIN" ]; then
    echo "Building benchmark_initial..."
    go build -o "$BENCHMARK_BIN" ./cmd/benchmark_initial
fi

# Run the benchmark
echo "Running benchmark..."
"$BENCHMARK_BIN" \
    -numRows=$NUM_ROWS \
    -rowLen=$ROW_LEN \
    -pirType=$PIR_TYPE \
    -updatable=true

echo ""
echo "Benchmark completed!"

