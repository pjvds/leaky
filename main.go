package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/pjvds/tidy"
)

var log = tidy.GetLogger()

type Monitor struct {
	closed  chan struct{}
	closing chan struct{}
}

type Leaky struct {
	Start         time.Time
	End           time.Time
	Collections   uint32
	Growth        uint32
	GrowthPerHour uint32
	Reason        string
}

func NewMonitor() Monitor {
	monitor := Monitor{
		closed:  make(chan struct{}),
		closing: make(chan struct{}),
	}

	go monitor.do()
	return monitor
}

// Reads the memory stats as soon as a GC cycle happened.
// The latency is between 100µs-500µs (0.1ms-0.5ms) after
// a garbage collection.
func trapGc() runtime.MemStats {
	read := make(chan runtime.MemStats)

	go func() *byte {
		ref := new(byte)
		runtime.SetFinalizer(ref, func(_ *byte) {
			stats := new(runtime.MemStats)
			runtime.ReadMemStats(stats)

			read <- *stats
		})

		return ref
	}()

	return <-read
}

func (this Monitor) do() {
	defer close(this.closed)
	var lastNumGC uint32

	for {
		stats := trapGc()

		if lastNumGC == stats.NumGC {
			log.Withs(tidy.Fields{
				"LastNumGC": lastNumGC,
				"NumGC":     stats.NumGC,
			}).Error("unexpected numgc value")
		} else {
			lastNumGC = stats.NumGC
		}

		log.Withs(tidy.Fields{
			"NumGC":       stats.NumGC,
			"Alloc":       stats.Alloc,
			"HeapAlloc":   stats.HeapAlloc,
			"TotalAlloc":  stats.TotalAlloc,
			"Mallocs":     stats.Mallocs,
			"TimeSinceGC": time.Now().Sub(time.Unix(0, int64(stats.LastGC))),
		}).Debug("GC ran")
	}
}

func snapshotFromStats(stats runtime.MemStats) Snapshot {
	return Snapshot{
		Alloc:       stats.Alloc,
		HeapAlloc:   stats.HeapAlloc,
		HeapObjects: stats.HeapObjects,
		NumGC:       stats.NumGC,
		TakenAt:     time.Now(),
	}
}

type Snapshot struct {
	// bytes allocated and still in use
	Alloc uint64

	// Active heap memory
	HeapAlloc uint64

	// The total amount of memory (address space) requested from the OS
	sys uint64

	// total number of allocated objects
	HeapObjects uint64

	// total number of GC cycles
	NumGC uint32

	TakenAt time.Time
}

func main() {
	monitor := NewMonitor()
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(1 * time.Second)

	junk := "x"

	for {
		select {
		case <-interrupt:
			return
		case <-monitor.closed:
			return
		case <-ticker.C:
			junk += junk
			fmt.Printf("tick, %v\n", len(junk))
		}
	}
}
