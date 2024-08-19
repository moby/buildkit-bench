package test

import (
	"testing"

	"github.com/moby/buildkit-bench/util/testutil"
)

func runTest(t *testing.T, testFunc func(t *testing.T, sb testutil.Sandbox)) {
	testutil.Run(t, testutil.TestFuncs(testFunc))
}
