package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"
)

type Snapshot struct {
	// bytes allocated and still in use
	Alloc uint64

	// total number of allocated objects
	HeapObjects uint64

	// total number of GC cycles
	NumGC uint32
}

type Diff struct {
	Before Snapshot
	After  Snapshot

	Change
}

type Change struct {
	Alloc       uint64
	HeapObjects uint64
}

func main() {
	var stats runtime.MemStats
	tick := time.Tick(1 * time.Second)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-tick:
			runtime.ReadMemStats(&stats)
			fmt.Printf("%+v\n\n\n", stats)
		case <-interrupt:
			return
		}
	}
}
