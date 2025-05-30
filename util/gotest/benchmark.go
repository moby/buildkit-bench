package gotest

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/moby/buildkit-bench/util/gotest/benchmark"
	"github.com/pkg/errors"
	"gotest.tools/gotestsum/testjson"
)

type Benchmark struct {
	Name string
	Ref  string
	Runs []BenchmarkRun
}

type BenchmarkRun struct {
	benchmark.Benchmark
	BenchmarkRunInfo
}

type BenchmarkRunInfo struct {
	Test    string
	Run     int
	Ref     string
	Buildx  string
	Threads int
}

func newBenchmark(event testjson.TestEvent) (ParseEntry, error) {
	be, ok, err := parseBenchmark(event.Output)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse benchmark: %s", event.Output)
	} else if !ok {
		return nil, nil
	}
	return be, nil
}

func (b *Benchmark) ID() string {
	return fmt.Sprintf("%s/ref=%s", b.Name, b.Ref)
}

func (b *Benchmark) Update(event testjson.TestEvent) error {
	entry, err := newBenchmark(event)
	if err != nil {
		return err
	}
	if be, ok := entry.(*Benchmark); ok && be != nil {
		b.Runs = append(b.Runs, be.Runs...)
	}
	return nil
}

func parseBenchmark(b string) (*Benchmark, bool, error) {
	bb := &Benchmark{}

	binfo, err := benchmark.ParseLine(b)
	if err != nil {
		// not a benchmark line, return without error
		return nil, false, nil
	}
	brun := BenchmarkRun{
		Benchmark: *binfo,
	}

	bri, err := parseBenchmarkRunInfo(binfo.Name)
	if err != nil {
		return nil, false, err
	}
	if bri.Ref == "" || bri.Run == 0 {
		return nil, false, nil
	}

	brun.BenchmarkRunInfo = *bri
	bb.Name = brun.Test
	bb.Ref = brun.Ref
	bb.Runs = []BenchmarkRun{brun}
	return bb, true, nil
}

func parseBenchmarkRunInfo(line string) (*BenchmarkRunInfo, error) {
	bri := &BenchmarkRunInfo{}
	var attrs []string
	for _, part := range strings.Split(line, "/") {
		if !strings.Contains(part, "=") {
			if len(bri.Test) > 0 {
				bri.Test += "/"
			}
			bri.Test += part
		} else {
			attrs = append(attrs, part)
		}
	}
	if len(attrs) == 0 {
		return bri, nil
	}

	csvAttrs := strings.Join(attrs, ",")
	csvReader := csv.NewReader(strings.NewReader(csvAttrs))
	fields, err := csvReader.Read()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read benchmark attributes: %s", csvAttrs)
	}

	for _, field := range fields {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			return nil, errors.Errorf("invalid value %s", field)
		}
		switch key {
		case "ref":
			bri.Ref = value
		case "buildx":
			bri.Buildx = value
		case "run":
			rkey, rvalue, ok := strings.Cut(value, "-")
			if !ok {
				return nil, errors.Errorf("invalid benchmark run value %s", value)
			}
			rr, err := strconv.Atoi(rkey)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse benchmark run count value: %s", rkey)
			}
			bri.Run = rr
			rt, err := strconv.Atoi(rvalue)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse benchmark threads value: %s", rkey)
			}
			bri.Threads = rt
		}
	}

	return bri, nil
}

type BenchmarkInfo struct {
	OS           string
	Architecture string
	Package      string
	CPU          string
}

func (b *BenchmarkInfo) update(output string) bool {
	output = strings.TrimSpace(output)
	if output == "" {
		return false
	}
	// https://github.com/golang/go/blob/f38d42f2c4c6ad0d7cbdad5e1417cac3be2a5dcb/src/testing/benchmark.go#L246-L255
	if strings.HasPrefix(output, "goos: ") {
		b.OS = strings.TrimPrefix(output, "goos: ")
	} else if strings.HasPrefix(output, "goarch: ") {
		b.Architecture = strings.TrimPrefix(output, "goarch: ")
	} else if strings.HasPrefix(output, "pkg: ") {
		b.Package = strings.TrimPrefix(output, "pkg: ")
	} else if strings.HasPrefix(output, "cpu: ") {
		b.CPU = strings.TrimPrefix(output, "cpu: ")
	} else {
		return false
	}
	return true
}
