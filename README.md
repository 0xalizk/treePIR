** Buggy for DB > 2GB **


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

- **DB Size**: Total logical database size (number of rows × row length), shown in a human‑readable unit.
- **OfflineServerTime**: Average server time per operation in the **offline phase** (hint generation).
- **OfflineClientTime**: Average client time per operation in the **offline phase** (processing hints / initialization).
- **OfflineBytes**: Average communication volume in the offline phase, per operation.
- **ClientBytes**: Client storage required to hold TreePIR state (keys, hints, etc.).
- **OnlineServerTime**: Average server time per query in the **online phase**.
- **OnlineClientTime**: Average client time per query in the online phase (query generation + decryption).
- **OnlineBytes**: Average online communication volume per query.


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


