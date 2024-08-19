package test

import (
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestDaemonVersion(t *testing.T) {
	runTest(t, func(t *testing.T, sb testutil.Sandbox) {
		buildkitdPath := path.Join(sb.BinsDir(), sb.Name(), "buildkitd")

		output, err := exec.Command(buildkitdPath, "--version").Output()
		require.NoError(t, err)

		versionParts := strings.Fields(string(output))
		require.Len(t, versionParts, 4)
		require.Equal(t, "buildkitd", versionParts[0])
		t.Log("repo:", versionParts[1])
		t.Log("version:", versionParts[2])
		t.Log("commit:", versionParts[3])
	})
}
