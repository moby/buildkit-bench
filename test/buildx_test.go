package test

import (
	"os/exec"
	"testing"

	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestBuildx(t *testing.T) {
	testutil.Run(t, testutil.TestFuncs(
		testBuildxVersion,
	))
}

func testBuildxVersion(t *testing.T, sb testutil.Sandbox) {
	output, err := exec.Command(sb.BuildxBin(), "version").Output()
	require.NoError(t, err)
	t.Log(string(output))
}
