package gotest

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type Result struct {
	Tests         map[string]Test
	Benchmarks    map[string]Benchmark
	BenchmarkInfo BenchmarkInfo
}

func MergeBenchmarks(dir string) (map[string]Benchmark, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", dir)
	}
	var benchmarks map[string]Benchmark
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fp := filepath.Join(dir, file.Name())
		result, err := loadResult(fp)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load result from file %s", fp)
		}
		if benchmarks == nil {
			benchmarks = result.Benchmarks
		} else {
			for k, v := range result.Benchmarks {
				benchmarks[k] = v
			}
		}
	}
	return benchmarks, nil
}

func loadResult(filename string) (*Result, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open result file %s", filename)
	}
	defer file.Close()
	var res Result
	if err := json.NewDecoder(file).Decode(&res); err != nil {
		return nil, errors.Wrapf(err, "failed to decode result file %s", filename)
	}
	return &res, nil
}
