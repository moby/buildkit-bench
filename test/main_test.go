package test

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/moby/buildkit-bench/util/testutil"
)

func TestMain(m *testing.M) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "1" {
		os.Exit(0)
	}
	flag.Parse()
	if filename := os.Getenv("TEST_EXECUTION_FILE"); filename != "" {
		if err := testutil.WriteExecutionMetadata(filename); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}
