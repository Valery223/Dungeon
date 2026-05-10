package usecase_test

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/Valery223/Dungeon/internal/domain"
	infra_io "github.com/Valery223/Dungeon/internal/infrastructure/io"
	"github.com/Valery223/Dungeon/internal/infrastructure/memory"
	"github.com/Valery223/Dungeon/internal/usecase"
)

// TestGameRunner_Load1MillionEvents - load test for processing 1 million events (100k players with 10 events each)
func TestGameRunner_Load1MillionEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode")
	}

	// 1 Create temp file
	tmpFile, err := os.CreateTemp("", "load_test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name()) // Will be deleted automatically after the test
	}()

	t.Log("Generating 1,000,000 lines (100,000 players)...")

	// Generate 1,000,000 lines (100,000 players, 10 events each)
	// Use a simple loop to create a large log
	for i := 1; i <= 100000; i++ {
		// Registration, entry, kill, next floor, damage, healing and etc.
		_, _ = fmt.Fprintf(tmpFile, "[14:00:00] %d 1\n", i)
		_, _ = fmt.Fprintf(tmpFile, "[14:05:00] %d 2\n", i)
		_, _ = fmt.Fprintf(tmpFile, "[14:06:00] %d 3\n", i)
		_, _ = fmt.Fprintf(tmpFile, "[14:07:00] %d 4\n", i)
		_, _ = fmt.Fprintf(tmpFile, "[14:08:00] %d 11 20\n", i) // Damage
		_, _ = fmt.Fprintf(tmpFile, "[14:09:00] %d 10 10\n", i) // Healing
		_, _ = fmt.Fprintf(tmpFile, "[14:10:00] %d 3\n", i)
		_, _ = fmt.Fprintf(tmpFile, "[14:11:00] %d 4\n", i)
		_, _ = fmt.Fprintf(tmpFile, "[14:12:00] %d 6\n", i)
		_, _ = fmt.Fprintf(tmpFile, "[14:15:00] %d 7\n", i)
	}
	_ = tmpFile.Close()

	// 2 Open the file for reading
	f, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open temp file: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	cfg := &domain.DungeonConfig{
		Floors:      2,
		Monsters:    2,
		OpenAtSec:   14 * 3600,
		DurationSec: 2 * 3600,
		CloseAtSec:  16 * 3600,
	}

	reader := infra_io.NewStreamReader(f)
	// Use a no-op writer to avoid console output during load test
	writer := infra_io.NewStreamWriter(io.Discard)
	repo := memory.NewInMemoryPlayerRepo()
	processor := usecase.NewEventProcessor(cfg, repo)
	runner := usecase.NewGameRunner(reader, writer, processor)

	// 3 Force GC before starting to get more accurate memory usage
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	start := time.Now()

	// 4 Run the game runner
	err = runner.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// 5 Measure time and memory usage
	duration := time.Since(start)
	runtime.ReadMemStats(&m2)

	// Calculate memory usage in MB
	allocMB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
	heapMB := float64(m2.HeapAlloc) / 1024 / 1024

	t.Logf("Processed 1,000,000 events in: %v", duration)
	t.Logf("Memory allocated (TotalAlloc): %.2f MB", allocMB)
	t.Logf("Memory in use (HeapAlloc): %.2f MB", heapMB)
}
