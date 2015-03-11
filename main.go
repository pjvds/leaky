package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"
)

func main() {
	var stats runtime.MemStats
	timer := time.NewTimer(1 * time.Second)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-timer.C:
			runtime.ReadMemStats(&stats)
			fmt.Printf("%+v\n\n\n", stats)
		case <-interrupt:
			return
		}
	}
}
