package testutil

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type TestConfig struct {
	Defaults TestConfigDefaults       `yaml:"defaults"`
	Runs     map[string]TestConfigRun `yaml:"runs"`
}

type TestConfigDefaults struct {
	Count     int    `yaml:"count"`
	Benchtime string `yaml:"benchtime"`
}

type TestConfigRun map[string]TestConfigBenchmark

type TestConfigBenchmark struct {
	Description string                      `yaml:"description"`
	Count       int                         `yaml:"count,omitempty" json:",omitempty"`
	Benchtime   string                      `yaml:"benchtime,omitempty" json:",omitempty"`
	Metrics     map[string]TestConfigMetric `yaml:"metrics"`
}

type TestConfigMetric struct {
	Description string `yaml:"description"`
}

func LoadTestConfig(f string) (*TestConfig, error) {
	dt, err := os.ReadFile(f)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read test config")
	}
	var testconfig TestConfig
	if err := yaml.Unmarshal(dt, &testconfig); err != nil {
		return nil, errors.Wrap(err, "failed to decode test config")
	}
	return &testconfig, nil
}

func (t *TestConfig) BenchmarkConfig(name string) (*TestConfigBenchmark, error) {
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return nil, errors.Errorf("invalid benchmark name: %s", name)
	}
	run, ok := t.Runs[parts[0]]
	if !ok {
		return nil, errors.Errorf("benchmark run not found: %s", parts[0])
	}
	benchmark, ok := run[parts[1]]
	if !ok {
		return nil, errors.Errorf("benchmark not found: %s", parts[1])
	}
	return &benchmark, nil
}
