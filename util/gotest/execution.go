package gotest

import "time"

// Execution describes the process that ran the benchmarks, not the exporter.
type Execution struct {
	ID          string            `json:"id"`
	CapturedAt  time.Time         `json:"captured_at,omitzero"`
	GoVersion   string            `json:"go_version"`
	NumCPU      int               `json:"num_cpu"`
	GOMAXPROCS  int               `json:"gomaxprocs"`
	Settings    map[string]string `json:"settings"`
	Environment map[string]string `json:"environment"`
	Binaries    []BinaryIdentity  `json:"binaries"`
}

// BinaryIdentity identifies a file available to the benchmark harness.
type BinaryIdentity struct {
	Project string `json:"project"`
	Ref     string `json:"ref"`
	Name    string `json:"name"`
	SHA256  string `json:"sha256"`
	Version string `json:"version,omitempty"`
}
