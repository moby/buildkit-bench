package benchmark

// Forked from https://cs.opensource.google/go/x/tools/+/master:benchmark/parse/parse.go
// with changes to handle additional metrics reported by ReportMetric.

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Flags used by Benchmark.Measured to indicate
// which measurements a Benchmark contains.
const (
	NsPerOp = 1 << iota
	MBPerS
	AllocedBytesPerOp
	AllocsPerOp
)

// Benchmark is one run of a single benchmark.
type Benchmark struct {
	Name              string             // benchmark name
	N                 int                // number of iterations
	NsPerOp           float64            // nanoseconds per iteration
	AllocedBytesPerOp uint64             // bytes allocated per iteration
	AllocsPerOp       uint64             // allocs per iteration
	MBPerS            float64            // MB processed per second
	Measured          int                // which measurements were recorded
	Ord               int                // ordinal position within a benchmark run
	Extra             map[string]float64 // Extra records additional metrics reported by ReportMetric.
}

// ParseLine extracts a Benchmark from a single line of testing.B
// output.
func ParseLine(line string) (*Benchmark, error) {
	fields := strings.Fields(line)

	// Two required, positional fields: Name and iterations.
	if len(fields) < 2 {
		return nil, errors.Errorf("two fields required, have %d", len(fields))
	}
	if !strings.HasPrefix(fields[0], "Benchmark") {
		return nil, errors.Errorf(`first field does not start with "Benchmark"`)
	}
	n, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, err
	}
	b := &Benchmark{Name: fields[0], N: n}

	// Parse any remaining pairs of fields; we've parsed one pair already.
	for i := 1; i < len(fields)/2; i++ {
		b.parseMeasurement(fields[i*2], fields[i*2+1])
	}
	return b, nil
}

func (b *Benchmark) parseMeasurement(quant string, unit string) {
	switch unit {
	case "ns/op":
		if f, err := strconv.ParseFloat(quant, 64); err == nil {
			b.NsPerOp = f
			b.Measured |= NsPerOp
		}
	case "MB/s":
		if f, err := strconv.ParseFloat(quant, 64); err == nil {
			b.MBPerS = f
			b.Measured |= MBPerS
		}
	case "B/op":
		if i, err := strconv.ParseUint(quant, 10, 64); err == nil {
			b.AllocedBytesPerOp = i
			b.Measured |= AllocedBytesPerOp
		}
	case "allocs/op":
		if i, err := strconv.ParseUint(quant, 10, 64); err == nil {
			b.AllocsPerOp = i
			b.Measured |= AllocsPerOp
		}
	default:
		if b.Extra == nil {
			b.Extra = make(map[string]float64)
		}
		if f, err := strconv.ParseFloat(quant, 64); err == nil {
			b.Extra[unit] = f
		}
	}
}

// Set is a collection of benchmarks from one
// testing.B run, keyed by name to facilitate comparison.
type Set map[string][]*Benchmark

// ParseSet extracts a Set from testing.B output.
// ParseSet preserves the order of benchmarks that have identical
// names.
func ParseSet(r io.Reader) (Set, error) {
	bb := make(Set)
	scan := bufio.NewScanner(r)
	ord := 0
	for scan.Scan() {
		if b, err := ParseLine(scan.Text()); err == nil {
			b.Ord = ord
			ord++
			bb[b.Name] = append(bb[b.Name], b)
		}
	}

	if err := scan.Err(); err != nil {
		return nil, err
	}

	return bb, nil
}
