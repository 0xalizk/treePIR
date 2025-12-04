## TreePIR Benchmarks with Claude


### 1GB Database Benchmark

- **DB size**: 1GB
- **Parameters**:
  - `rowLen = 1024` bytes
  - `numRows = 1048576`
  - `DB Size = numRows × rowLen = 1GB`

```bash
./benchmark_initial -numRows=1048576 -rowLen=1024 -pirType=TreePIR -updatable=true
```

### Result

```text
# benchmark_initial -numRows=1048576 -rowLen=1024 -pirType=TreePIR -updatable=true
   DB Size   OfflineServerTime   OfflineClientTime   OfflineBytes    ClientBytes    OnlineServerTime    OnlineClientTime    OnlineBytes
    1.00GB             15.088s             117.1ms        89.83MB         1.57MB               4.3ms               1.1ms         2.00MB
```

### Reproducing These Benchmarks

From the repository root:

```bash
# Build the benchmark binary
go build -o benchmark_initial ./cmd/benchmark_initial

# ~256MB database (1KB rows)
./benchmark_initial -numRows=262144 -rowLen=1024 -pirType=TreePIR -updatable=true

# 1GB database (1KB rows)
./benchmark_initial -numRows=1048576 -rowLen=1024 -pirType=TreePIR -updatable=true
```

You can change `numRows` and `rowLen` to explore other database sizes; the benchmark will always print
the **DB Size** column in a human‑readable format (KB / MB / GB).


