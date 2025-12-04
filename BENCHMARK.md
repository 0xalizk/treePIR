# Benchmarking PIR on a 1GB Database

This guide explains how to benchmark the PIR scheme on a 1GB database.

## Quick Start

Run the provided script:

```bash
./benchmark.sh [pirType] [rowLen]
```

### Parameters

- **pirType** (optional): PIR type to use. Options:
  - `TreePIR` (default) - TreePIR scheme
  - `Punc` - Checklist's PIR scheme
  - `Matrix` - Matrix-based PIR
  - `DPF` - DPF-based PIR
  - `Perm` - Permutation-based PIR
  - `NonPrivate` - No privacy (baseline)
  
- **rowLen** (optional): Row length in bytes (default: 1024)

### Examples

```bash
# Benchmark TreePIR with 1KB rows (default)
./benchmark.sh

# Benchmark TreePIR with 4KB rows
./benchmark.sh TreePIR 4096

# Benchmark Punc scheme with 32-byte rows
./benchmark.sh Punc 32

# Benchmark DPF with 1KB rows
./benchmark.sh DPF 1024
```

## Manual Execution

You can also run the benchmark manually:

```bash
# Build the benchmark
go build -o benchmark_initial ./cmd/benchmark_initial

# Calculate parameters for 1GB
# For 1GB = 1,073,741,824 bytes
# If rowLen = 1024, then numRows = 1,048,576

# Run benchmark
./benchmark_initial -numRows=1048576 -rowLen=1024 -pirType=TreePIR -updatable=true
```

## Database Size Calculations

For a 1GB database (1,073,741,824 bytes):

| Row Length | Number of Rows |
|------------|----------------|
| 32 bytes   | 33,554,432     |
| 1024 bytes | 1,048,576      |
| 4096 bytes | 262,144        |

## Resource Requirements

**Warning**: Benchmarking a 1GB database requires:
- Significant memory (several GB of RAM)
- Time to complete (may take minutes to hours depending on the PIR scheme)
- Disk space for any generated profiles

For testing, consider starting with smaller databases first:
```bash
# Test with 10MB first
./benchmark_initial -numRows=10000 -rowLen=1024 -pirType=TreePIR
```

## Output

The benchmark outputs the following metrics:

- **DB Size**: Total database size (`numRows × rowLen`) shown in KB/MB/GB
- **OfflineServerTime[us]**: Server time for offline phase (hint generation)
- **OfflineClientTime[us]**: Client time for offline phase (hint processing)
- **OfflineBytes**: Communication bytes in offline phase
- **ClientBytes**: Client storage bytes
- **OnlineServerTime[us]**: Server time for online phase (query processing)
- **OnlineClientTime[us]**: Client time for online phase (query generation/decryption)
- **OnlineBytes**: Communication bytes in online phase

## Other Benchmark Tools

The codebase includes other benchmark tools:

- `cmd/benchmark_initial/` - Initial setup and query benchmarks
- `cmd/benchmark_updates/` - Database update benchmarks
- `cmd/benchmark_incremental/` - Incremental update benchmarks
- `cmd/benchmark_trace/` - Trace-based benchmarks

You can adapt the same parameters for these tools as well.

