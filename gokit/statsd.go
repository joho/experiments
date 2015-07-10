package main

import (
	"net"
	"os"
	"runtime"
	"time"

	"github.com/go-kit/kit/metrics/statsd"
)

func main() {
	statsdWriter, err := net.Dial("udp", "127.0.0.1:8126")
	if err != nil {
		os.Exit(1)
	}

	reportingDuration := 5 * time.Second
	goroutines := statsd.NewGauge(statsdWriter, "total_goroutines", reportingDuration)
	for range time.Tick(reportingDuration) {
		goroutines.Set(float64(runtime.NumGoroutine()))
	}
}
