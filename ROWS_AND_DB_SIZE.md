# Rows and Database Size Relationship

## Basic Formula

The relationship between rows and database size is straightforward:

```
Database Size (bytes) = Number of Rows × Row Length (bytes)
```

Or rearranged:

```
Number of Rows = Database Size (bytes) ÷ Row Length (bytes)
Row Length (bytes) = Database Size (bytes) ÷ Number of Rows
```

## Examples for 1GB Database

For a **1GB database** (1,073,741,824 bytes), here are common configurations:

| Row Length | Number of Rows | Description |
|------------|----------------|-------------|
| 32 bytes   | 33,554,432     | Many small rows (default for some schemes) |
| 256 bytes  | 4,194,304      | Small rows |
| 1,024 bytes (1KB) | 1,048,576 | **Recommended** - balanced |
| 4,096 bytes (4KB) | 262,144 | Medium rows |
| 16,384 bytes (16KB) | 65,536 | Large rows |
| 65,536 bytes (64KB) | 16,384 | Very large rows |
| 104,857,600 bytes (100MB) | 10 | **Too few rows** - TreePIR requires ≥16 |

## Trade-offs

### Many Small Rows (e.g., 32 bytes, 33M rows)
- ✅ Better for schemes that benefit from many rows
- ✅ More granular queries
- ❌ Higher overhead per row
- ❌ More complex indexing

### Few Large Rows (e.g., 100MB, 10 rows)
- ✅ Simple structure
- ❌ **TreePIR requires at least 16 rows** (setSize = √n)
- ❌ Less flexible for partial data retrieval
- ❌ May hit memory limits

### Balanced (e.g., 1KB-16KB rows)
- ✅ Good balance for most PIR schemes
- ✅ Reasonable number of rows
- ✅ Efficient for typical use cases
- ✅ **Recommended for TreePIR**

## TreePIR Specific Requirements

TreePIR has a minimum requirement:
- **Minimum 16 rows** (setSize = √n, minimum setSize = 4, so n ≥ 16)
- For optimal performance, use row lengths between **1KB and 64KB**
- Number of rows should be a power of 2 or close to it for best performance

## Calculating Your Configuration

### Given: Target Database Size
If you want a specific database size:

```bash
# For 1GB database with 1KB rows:
numRows = 1,073,741,824 ÷ 1,024 = 1,048,576 rows

# For 1GB database with 4KB rows:
numRows = 1,073,741,824 ÷ 4,096 = 262,144 rows
```

### Given: Number of Rows
If you have a specific number of rows:

```bash
# For 1,048,576 rows to make 1GB:
rowLen = 1,073,741,824 ÷ 1,048,576 = 1,024 bytes (1KB)

# For 262,144 rows to make 1GB:
rowLen = 1,073,741,824 ÷ 262,144 = 4,096 bytes (4KB)
```

## Quick Reference Table

| Database Size | Row Length | Number of Rows |
|---------------|------------|----------------|
| 1 MB          | 1 KB       | 1,024          |
| 10 MB         | 1 KB       | 10,240         |
| 100 MB        | 1 KB       | 102,400        |
| 1 GB          | 1 KB       | 1,048,576      |
| 1 GB          | 4 KB       | 262,144        |
| 1 GB          | 16 KB      | 65,536         |
| 10 GB         | 1 KB       | 10,485,760     |
| 10 GB         | 4 KB       | 2,621,440      |

## In the Benchmark Script

The `benchmark.sh` script automatically calculates:

```bash
NUM_ROWS = 1,073,741,824 ÷ ROW_LEN
```

So when you run:
- `./benchmark.sh TreePIR 1024` → 1,048,576 rows
- `./benchmark.sh TreePIR 4096` → 262,144 rows
- `./benchmark.sh TreePIR 104857600` → 10 rows (⚠️ too few!)

## Performance Implications

The choice of row length affects:

1. **Offline Phase**: 
   - More rows = more hints to generate
   - Larger rows = larger hints

2. **Online Phase**:
   - More rows = larger setSize = more computation
   - Larger rows = more data to transfer per query

3. **Memory**:
   - More rows = more memory for indexing
   - Larger rows = more memory per row

**Recommendation**: Use **1KB to 16KB** row lengths for best balance.

