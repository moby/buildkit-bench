package testutil

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOfficialImages(t *testing.T) {
	namespace := runtime.GOARCH
	if namespace == "arm64" {
		namespace = "arm64v8"
	} else if namespace != "amd64" {
		namespace = "library"
	}
	for _, tc := range []struct {
		name     string
		images   []string
		expected map[string]string
	}{
		{
			name:   "host architecture",
			images: []string{"busybox:latest"},
			expected: map[string]string{
				"library/busybox:latest": "docker.io/" + namespace + "/busybox:latest@" + pins["busybox:latest"][namespace],
			},
		},
		{
			name:   "multi-platform override does not affect subsequent images",
			images: []string{"library/busybox:latest", "alpine:latest"},
			expected: map[string]string{
				"library/busybox:latest": "docker.io/library/busybox:latest@" + pins["busybox:latest"]["library"],
				"library/alpine:latest":  "docker.io/" + namespace + "/alpine:latest@" + pins["alpine:latest"][namespace],
			},
		},
		{
			name:   "unpinned explicit namespace",
			images: []string{"library/busybox:unlisted"},
			expected: map[string]string{
				"library/busybox:unlisted": "docker.io/library/busybox:unlisted",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, OfficialImages(tc.images...))
		})
	}
}
