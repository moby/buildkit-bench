package testutil

import (
	"testing"
)

type MetricUnit string

const (
	MetricUnitBytesPerOp  MetricUnit = "B/op"
	MetricUnitNsPerOp     MetricUnit = "ns/op"
	MetricUnitBytes       MetricUnit = "B"
	MetricUnitAllocations MetricUnit = "allocs/op"
)

func ReportMetric(b *testing.B, value float64, unit MetricUnit) {
	b.ReportMetric(value, string(unit))
}
