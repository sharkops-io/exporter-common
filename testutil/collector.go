// Package testutil provides test helpers shared across this repo's packages.
// These helpers are exported for use by exporters built on this module; they
// are test-only and not part of any runtime collection path.
package testutil

import "github.com/prometheus/client_golang/prometheus"

// MetricSlice renders a fixed slice of already-built metrics, so exposition
// output can be asserted without any collector machinery.
type MetricSlice []prometheus.Metric

func (c MetricSlice) Describe(chan<- *prometheus.Desc) {}

func (c MetricSlice) Collect(ch chan<- prometheus.Metric) {
	for _, m := range c {
		ch <- m
	}
}
