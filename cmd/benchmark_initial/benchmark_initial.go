package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"checklist/driver"
	"checklist/pir"
	"checklist/updatable"

	"gotest.tools/assert"
)

// formatTime converts microseconds to a human-readable string (ms or s)
func formatTime(us float64) string {
	if us >= 1000000 {
		return fmt.Sprintf("%.3fs", us/1000000.0)
	} else if us >= 1000 {
		return fmt.Sprintf("%.1fms", us/1000.0)
	} else {
		return fmt.Sprintf("%.0fus", us)
	}
}

// formatBytes converts bytes to a human-readable string (KB or MB)
func formatBytes(bytes float64) string {
	if bytes >= 1048576 {
		return fmt.Sprintf("%.2fMB", bytes/1048576.0)
	} else if bytes >= 1024 {
		return fmt.Sprintf("%.2fKB", bytes/1024.0)
	} else {
		return fmt.Sprintf("%.0fB", bytes)
	}
}

// formatDBSize converts bytes to a human-readable string (KB, MB, or GB)
func formatDBSize(bytes float64) string {
	if bytes >= 1073741824 {
		return fmt.Sprintf("%.2fGB", bytes/1073741824.0)
	} else if bytes >= 1048576 {
		return fmt.Sprintf("%.2fMB", bytes/1048576.0)
	} else if bytes >= 1024 {
		return fmt.Sprintf("%.2fKB", bytes/1024.0)
	} else {
		return fmt.Sprintf("%.0fB", bytes)
	}
}

func main() {
	config := new(driver.Config).AddPirFlags().AddClientFlags().AddBenchmarkFlags().Parse()

	var ep driver.ErrorPrinter

	prof := driver.NewProfiler(config.CpuProfile)
	defer prof.Close()

	fmt.Printf("# %s %s\n", path.Base(os.Args[0]), strings.Join(os.Args[1:], " "))
	fmt.Printf("%10s%20s%20s%15s%15s%20s%20s%15s\n",
		"DB Size", "OfflineServerTime", "OfflineClientTime", "OfflineBytes", "ClientBytes",
		"OnlineServerTime", "OnlineClientTime", "OnlineBytes")

	dr, err := config.ServerDriver()
	if err != nil {
		log.Fatalf("Failed to create driver: %s\n", err)
	}

	rand := pir.RandSource()

	var clientStatic pir.PIRReader
	var clientUpdatable *updatable.Client
	var none int
	if err := dr.Configure(config.TestConfig, &none); err != nil {
		log.Fatalf("Failed to configure driver: %s\n", err)
	}

	result := testing.Benchmark(func(b *testing.B) {
		assert.NilError(ep, dr.ResetMetrics(0, &none))
		var clientInitTime time.Duration
		var clientBytes int
		for i := 0; i < b.N; i++ {
			start := time.Now()
			if config.Updatable {
				clientUpdatable = updatable.NewClient(pir.RandSource(), config.PirType, [2]updatable.UpdatableServer{dr, dr})
				err = clientUpdatable.Init()
			} else {
				clientStatic = pir.NewPIRReader(rand, dr, dr)
				err = clientStatic.Init(config.PirType)

			}
			assert.NilError(ep, err)
			clientInitTime += time.Since(start)
			if config.Updatable {
				clientBytes += clientUpdatable.StorageNumBytes(driver.SerializedSizeOf)
			}
		}

		var serverOfflineTime time.Duration
		assert.NilError(ep, dr.GetOfflineTimer(0, &serverOfflineTime))
		b.ReportMetric(float64(serverOfflineTime.Microseconds())/float64(b.N), "hint-us/op")
		b.ReportMetric(float64((clientInitTime-serverOfflineTime).Microseconds())/float64(b.N), "init-us/op")

		var offlineBytes int
		assert.NilError(ep, dr.GetOfflineBytes(0, &offlineBytes))
		b.ReportMetric(float64(offlineBytes)/float64(b.N), "hint-bytes/op")
		b.ReportMetric(float64(clientBytes)/float64(b.N), "client-bytes/op")
	})
	dbSize := float64(config.NumRows * config.RowLen)
	fmt.Printf("%10s%20s%20s%15s%15s",
		formatDBSize(dbSize),
		formatTime(result.Extra["hint-us/op"]),
		formatTime(result.Extra["init-us/op"]),
		formatBytes(result.Extra["hint-bytes/op"]),
		formatBytes(result.Extra["client-bytes/op"]))

	result = testing.Benchmark(func(b *testing.B) {
		assert.NilError(ep, dr.ResetMetrics(0, &none))
		var clientReadTime time.Duration
		for i := 0; i < b.N; i++ {
			var rowIV driver.RowIndexVal
			var row pir.Row

			var numKeys int
			if config.Updatable {
				assert.NilError(ep, dr.NumKeys(0, &numKeys))
			} else {
				numKeys = config.NumRows
			}
			assert.NilError(ep, dr.GetRow(rand.Intn(numKeys), &rowIV))

			start := time.Now()
			if clientStatic != nil {
				row, err = clientStatic.Read(rowIV.Index)
			} else {
				row, err = clientUpdatable.Read(rowIV.Key)
			}
			clientReadTime += time.Since(start)
			assert.NilError(ep, err)
			if row[0] != rowIV.Value[0] {
				fmt.Printf("BAD: %d\n", i)
			}

			if i == b.N-2 {
				runtime.GC()
				if memProf, err := os.Create("mem.prof"); err != nil {
					log.Printf("Failed to create memory profile: %s", err)
				} else {
					pprof.WriteHeapProfile(memProf)
					memProf.Close()
				}
			}
		}
		var serverOnlineTime time.Duration
		assert.NilError(ep, dr.GetOnlineTimer(0, &serverOnlineTime))
		b.ReportMetric(float64(serverOnlineTime.Microseconds())/float64(b.N), "answer-us/op")
		b.ReportMetric(float64((clientReadTime-serverOnlineTime).Microseconds())/float64(b.N), "read-us/op")

		var onlineBytes int
		assert.NilError(ep, dr.GetOnlineBytes(0, &onlineBytes))
		b.ReportMetric(float64(onlineBytes)/float64(b.N), "answer-bytes/op")

	})
	fmt.Printf("%20s%20s%15s\n",
		formatTime(result.Extra["answer-us/op"]),
		formatTime(result.Extra["read-us/op"]),
		formatBytes(result.Extra["answer-bytes/op"]))

}
