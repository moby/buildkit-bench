package testutil

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
	Description string   `yaml:"description"`
	Count       int      `yaml:"count,omitempty" json:",omitempty"`
	Benchtime   string   `yaml:"benchtime,omitempty" json:",omitempty"`
	Metrics     []string `yaml:"metrics"`
}
