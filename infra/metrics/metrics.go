/*
Package metrics provides utilities for capturing counters gauges and histograms
from various services written with bingo
*/
package metrics

import (
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
)

// Init initializes the metrics lib . It registers the runtime memory stats
//TODO: make it configurable to accept metrics publishers other than exp
func Init() {
	metrics.RegisterRuntimeMemStats(metrics.DefaultRegistry)
	go metrics.CaptureRuntimeMemStats(metrics.DefaultRegistry, 60*time.Second)
	exp.Exp(metrics.DefaultRegistry)
}

// AddCounter increments a counter by 1
func AddCounter(key string) {
	metrics.GetOrRegisterCounter(key, metrics.DefaultRegistry).Inc(1)
}

// AddGauge updates the gauge with the value v
func AddGauge(key string, v int64) {
	metrics.GetOrRegisterGauge(key, metrics.DefaultRegistry).Update(v)

}

// UpdateHistogram registers a new value for the histogram
func UpdateHistogram(name string, v int64) {
	metrics.GetOrRegisterHistogram(name, metrics.DefaultRegistry, metrics.NewUniformSample(1000)).Update(v)
}

// UpdateTimer registers a new value for a timer
func UpdateTimer(name string, t time.Duration) {
	metrics.GetOrRegisterTimer(name, metrics.DefaultRegistry).Update(t)
}

// UpdateTimerSince registers a new value for a timer
func UpdateTimerSince(name string, since time.Time) {
	metrics.GetOrRegisterTimer(name, metrics.DefaultRegistry).UpdateSince(since)
}
