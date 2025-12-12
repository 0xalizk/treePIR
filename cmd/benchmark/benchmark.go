package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"checklist/driver"
	"checklist/pir"

	"gotest.tools/assert"
)

// ParseSize converts human-readable size strings to bytes
func ParseSize(sizeStr string) (int64, error) {
	// Trim whitespace and convert to uppercase
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))

	// Regex to extract number and optional unit
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGT]?B?)$`)
	matches := re.FindStringSubmatch(sizeStr)

	if matches == nil {
		return 0, fmt.Errorf("invalid size format: %s (use formats like 1GB, 500MB, 100KB)", sizeStr)
	}

	// Parse the number
	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", matches[1])
	}

	// Parse the unit
	unit := matches[2]
	var multiplier int64 = 1

	switch unit {
	case "", "B":
		multiplier = 1
	case "K", "KB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	case "T", "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}

	return int64(num * float64(multiplier)), nil
}

// CalculateParameters determines optimal numRows and rowLen for a given database size
func CalculateParameters(dbSize int64, rowLenOverride int) (numRows, rowLen int) {
	if rowLenOverride > 0 {
		rowLen = rowLenOverride
	} else {
		// Heuristic selection based on dbSize
		switch {
		case dbSize < 10*1024*1024: // < 10MB
			rowLen = 16
		case dbSize < 100*1024*1024: // < 100MB
			rowLen = 32
		case dbSize < 1024*1024*1024: // < 1GB
			rowLen = 64
		default:
			rowLen = 256
		}
	}

	numRows = int(dbSize / int64(rowLen))

	// Ensure at least 1 row
	if numRows < 1 {
		numRows = 1
	}

	return numRows, rowLen
}

// BenchmarkResult holds all metrics from the benchmark
type BenchmarkResult struct {
	Config struct {
		DbSize   int64
		NumRows  int
		RowLen   int
		SetSize  int
		PIRType  string
	}
	Offline struct {
		ServerTimeUs int64
		ClientTimeUs int64
		CommBytes    int
		StorageBytes int
	}
	Online struct {
		AvgQueryGenUs    int64
		AvgServerTimeUs  int64
		AvgReconstructUs int64
		AvgTotalUs       int64
		AvgOnlineBytes   int
		NumQueries       int
		MinTimeUs        int64
		MaxTimeUs        int64
		StdDevUs         float64
	}
}

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

// FormatTime converts microseconds to human-readable format
func FormatTime(us int64) string {
	if us < 1000 {
		return fmt.Sprintf("%.2f us", float64(us))
	} else if us < 1000000 {
		return fmt.Sprintf("%.2f ms", float64(us)/1000.0)
	} else {
		return fmt.Sprintf("%.2f s", float64(us)/1000000.0)
	}
}

// RunBenchmark executes the full benchmark
func RunBenchmark(numRows, rowLen int, numQueries int, pirType pir.PirType) (*BenchmarkResult, error) {
	result := &BenchmarkResult{}
	result.Config.NumRows = numRows
	result.Config.RowLen = rowLen
	result.Config.DbSize = int64(numRows) * int64(rowLen)
	result.Config.SetSize = int(math.Sqrt(float64(numRows)))
	result.Config.PIRType = pirType.String()

	fmt.Printf("\nInitializing benchmark...\n")

	// Setup driver
	dr, err := driver.NewServerDriver()
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %v", err)
	}

	config := driver.TestConfig{
		NumRows:          numRows,
		RowLen:           rowLen,
		Updatable:        false,
		MeasureBandwidth: true,
	}

	var none int
	if err := dr.Configure(config, &none); err != nil {
		return nil, fmt.Errorf("failed to configure driver: %v", err)
	}

	// Offline Phase Benchmark
	fmt.Printf("Running offline phase (hint generation)...\n")
	randSource := pir.RandSource()
	var client pir.PIRReader
	var ep driver.ErrorPrinter

	benchResult := testing.Benchmark(func(b *testing.B) {
		assert.NilError(ep, dr.ResetMetrics(0, &none))
		var clientInitTime time.Duration

		for i := 0; i < b.N; i++ {
			start := time.Now()
			client = pir.NewPIRReader(randSource, dr, dr)
			err = client.Init(pirType)
			if err != nil {
				b.Fatal(err)
			}
			clientInitTime += time.Since(start)
		}

		var serverOfflineTime time.Duration
		assert.NilError(ep, dr.GetOfflineTimer(0, &serverOfflineTime))
		b.ReportMetric(float64(serverOfflineTime.Microseconds())/float64(b.N), "hint-us/op")
		b.ReportMetric(float64((clientInitTime-serverOfflineTime).Microseconds())/float64(b.N), "init-us/op")

		var offlineBytes int
		assert.NilError(ep, dr.GetOfflineBytes(0, &offlineBytes))
		b.ReportMetric(float64(offlineBytes)/float64(b.N), "hint-bytes/op")

		// Calculate client storage for TreePIR
		// NumHintsMultiplier = SecParam * ln(2) ≈ 128 * 0.693 ≈ 88
		// nHints = NumHintsMultiplier * numRows / setSize
		setSize := int(math.Sqrt(float64(numRows)))
		numHintsMultiplier := int(float64(128) * math.Log(2))
		nHints := numHintsMultiplier * numRows / setSize
		bitsPerKey := int(math.Log2(float64(nHints)))
		fixedBytes := nHints * rowLen
		storageBytes := (bitsPerKey*numRows+7)/8 + fixedBytes
		b.ReportMetric(float64(storageBytes), "client-bytes/op")
	})

	result.Offline.ServerTimeUs = int64(benchResult.Extra["hint-us/op"])
	result.Offline.ClientTimeUs = int64(benchResult.Extra["init-us/op"])
	result.Offline.CommBytes = int(benchResult.Extra["hint-bytes/op"])
	result.Offline.StorageBytes = int(benchResult.Extra["client-bytes/op"])

	// Online Phase Benchmark
	fmt.Printf("Running online phase (%d queries)...\n", numQueries)
	assert.NilError(ep, dr.ResetMetrics(0, &none))

	var queryTimes []int64
	var serverTimes []int64
	var totalOnlineBytes int

	for i := 0; i < numQueries; i++ {
		// Random query index
		queryIdx := randSource.Intn(numRows)

		// Get expected row for verification
		var rowIV driver.RowIndexVal
		assert.NilError(ep, dr.GetRow(queryIdx, &rowIV))

		// Reset metrics for this query
		assert.NilError(ep, dr.ResetMetrics(0, &none))

		// Time the query
		start := time.Now()
		row, err := client.Read(queryIdx)
		queryTime := time.Since(start)

		if err != nil {
			return nil, fmt.Errorf("query %d failed: %v", i, err)
		}

		// Verify correctness
		if row[0] != rowIV.Value[0] {
			return nil, fmt.Errorf("query %d returned wrong value", i)
		}

		queryTimes = append(queryTimes, queryTime.Microseconds())

		var serverOnlineTime time.Duration
		assert.NilError(ep, dr.GetOnlineTimer(0, &serverOnlineTime))
		serverTimes = append(serverTimes, serverOnlineTime.Microseconds())

		var onlineBytes int
		assert.NilError(ep, dr.GetOnlineBytes(0, &onlineBytes))
		totalOnlineBytes += onlineBytes
	}

	// Calculate statistics
	result.Online.NumQueries = numQueries
	result.Online.AvgTotalUs = average(queryTimes)
	result.Online.AvgServerTimeUs = average(serverTimes)
	result.Online.AvgQueryGenUs = result.Online.AvgTotalUs - result.Online.AvgServerTimeUs
	result.Online.MinTimeUs = min(queryTimes)
	result.Online.MaxTimeUs = max(queryTimes)
	result.Online.StdDevUs = stddev(queryTimes)
	result.Online.AvgOnlineBytes = totalOnlineBytes / numQueries

	return result, nil
}

// Helper functions for statistics
func average(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	sum := int64(0)
	for _, v := range values {
		sum += v
	}
	return sum / int64(len(values))
}

func min(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	minVal := values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}

func max(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

func stddev(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}
	avg := float64(average(values))
	variance := 0.0
	for _, v := range values {
		diff := float64(v) - avg
		variance += diff * diff
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}

// PrintReport displays formatted benchmark results
func PrintReport(result *BenchmarkResult) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TreePIR Benchmark Report")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("Database Configuration:")
	fmt.Printf("  Database Size:      %s (%d bytes)\n", FormatBytes(int(result.Config.DbSize)), result.Config.DbSize)
	fmt.Printf("  Number of Rows:     %d\n", result.Config.NumRows)
	fmt.Printf("  Row Length:         %d bytes\n", result.Config.RowLen)
	fmt.Printf("  PIR Type:           %s\n", result.Config.PIRType)
	fmt.Printf("  Set Size (√n):      %d\n", result.Config.SetSize)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("OFFLINE PHASE (One-time Setup)")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("  Server Hint Time:         %20s\n", FormatTime(result.Offline.ServerTimeUs))
	fmt.Printf("  Client Init Time:         %20s\n", FormatTime(result.Offline.ClientTimeUs))
	fmt.Printf("  Total Offline Time:       %20s\n", FormatTime(result.Offline.ServerTimeUs+result.Offline.ClientTimeUs))
	fmt.Println()
	fmt.Println("  Communication Cost:")
	fmt.Printf("    Offline Bytes:          %20s\n", FormatBytes(result.Offline.CommBytes))
	fmt.Printf("    Client Storage:         %20s\n", FormatBytes(result.Offline.StorageBytes))

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("ONLINE PHASE (Per Query, averaged over %d queries)\n", result.Online.NumQueries)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("  Query Generation:         %20s\n", FormatTime(result.Online.AvgQueryGenUs))
	fmt.Printf("  Server Answer Time:       %20s\n", FormatTime(result.Online.AvgServerTimeUs))
	fmt.Printf("  Total Online Time:        %20s\n", FormatTime(result.Online.AvgTotalUs))
	fmt.Println()
	fmt.Println("  Communication Cost:")
	fmt.Printf("    Per Query:              %20s\n", FormatBytes(result.Online.AvgOnlineBytes))
	fmt.Println()
	fmt.Println("  Statistics:")
	fmt.Printf("    Min Query Time:         %20s\n", FormatTime(result.Online.MinTimeUs))
	fmt.Printf("    Max Query Time:         %20s\n", FormatTime(result.Online.MaxTimeUs))
	fmt.Printf("    Std Deviation:          %20s\n", FormatTime(int64(result.Online.StdDevUs)))

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
}

// SaveResultsJSON saves benchmark results to a JSON file
func SaveResultsJSON(result *BenchmarkResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("\nResults saved to: %s\n", filename)
	return nil
}

func main() {
	// Parse flags
	var (
		rowLenOverride int
		numQueries     int
		forceRun       bool
		pirTypeName    string
		outputFile     string
	)

	flag.IntVar(&rowLenOverride, "rowLen", 0, "Override automatic rowLen selection")
	flag.IntVar(&numQueries, "queries", 100, "Number of online queries to benchmark")
	flag.BoolVar(&forceRun, "force", false, "Force run even if database size exceeds recommended limit")
	flag.StringVar(&pirTypeName, "pirType", "TreePIR", "PIR type to use (TreePIR, NonPrivate, Matrix, Punc, DPF)")
	flag.StringVar(&outputFile, "output", "", "Save results to JSON file (optional)")
	flag.Parse()

	// Get size argument
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: benchmark <size> [options]\n")
		fmt.Fprintf(os.Stderr, "Example: benchmark 1GB --queries=100\n\n")
		fmt.Fprintf(os.Stderr, "Size formats: 1GB, 500MB, 100MB, 10KB, or raw bytes\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	sizeStr := flag.Arg(0)

	// Parse size
	dbSize, err := ParseSize(sizeStr)
	if err != nil {
		log.Fatalf("Invalid size '%s': %v\n", sizeStr, err)
	}

	// Parse PIR type
	var pirType pir.PirType
	switch pirTypeName {
	case "TreePIR":
		pirType = pir.TreePIR
	case "NonPrivate":
		pirType = pir.NonPrivate
	case "Matrix":
		pirType = pir.Matrix
	case "Punc":
		pirType = pir.Punc
	case "DPF":
		pirType = pir.DPF
	default:
		log.Fatalf("Invalid PIR type '%s'. Valid options: TreePIR, NonPrivate, Matrix, Punc, DPF\n", pirTypeName)
	}

	// Calculate parameters
	numRows, rowLen := CalculateParameters(dbSize, rowLenOverride)
	actualSize := int64(numRows) * int64(rowLen)

	// Check for reasonable database size (in-memory limit ~8GB)
	const maxInMemorySize = 8 * 1024 * 1024 * 1024 // 8GB
	if actualSize > maxInMemorySize && !forceRun {
		log.Fatalf("Database size too large for in-memory benchmark (%s). Maximum recommended size is 8GB.\n"+
			"For larger databases, you can:\n"+
			"  - Use --force flag to bypass this check (may cause out-of-memory errors)\n"+
			"  - Benchmark smaller databases (1GB, 2GB, 4GB)\n"+
			"  - Use smaller row length with --rowLen flag (though total size stays the same)\n",
			FormatBytes(int(actualSize)))
	}

	if actualSize > maxInMemorySize && forceRun {
		fmt.Printf("\nWARNING: Database size (%s) exceeds recommended limit. This may cause memory errors.\n\n",
			FormatBytes(int(actualSize)))
	}

	// Print configuration
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("TreePIR Benchmark Configuration\n")
	fmt.Printf(strings.Repeat("=", 80) + "\n")
	fmt.Printf("  Target Size:        %s (%d bytes)\n", sizeStr, dbSize)
	fmt.Printf("  Actual Size:        %s (%d bytes)\n", FormatBytes(int(actualSize)), actualSize)
	fmt.Printf("  Num Rows:           %d\n", numRows)
	fmt.Printf("  Row Length:         %d bytes\n", rowLen)
	fmt.Printf("  PIR Type:           %s\n", pirTypeName)
	fmt.Printf("  Queries:            %d\n", numQueries)
	fmt.Printf(strings.Repeat("=", 80) + "\n")

	// Run benchmark
	result, err := RunBenchmark(numRows, rowLen, numQueries, pirType)
	if err != nil {
		log.Fatalf("Benchmark failed: %v\n", err)
	}

	// Display results
	PrintReport(result)

	// Save results to JSON if output file specified
	if outputFile != "" {
		if err := SaveResultsJSON(result, outputFile); err != nil {
			log.Fatalf("Failed to save results: %v\n", err)
		}
	}
}
