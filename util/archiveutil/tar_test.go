package archiveutil

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestManyFileManifest(t *testing.T) {
	sizes, err := manyFileSizes()
	require.NoError(t, err)
	require.Len(t, sizes, manyFileDirs*manyFilesPerDir)
	require.Equal(t, int64(536870912), sumInt64(sizes))
	require.Equal(t, "08c480e41c080b6181e35d1e2b869a1c69457a20a07fce07c39b900bcbef3742", manyFileManifestDigest(sizes))
}

func manyFileManifestDigest(sizes []int64) string {
	h := sha256.New()
	var buf [8]byte
	fileNum := 0
	for dirIndex := 0; dirIndex < manyFileDirs; dirIndex++ {
		for fileIndex := 0; fileIndex < manyFilesPerDir; fileIndex++ {
			h.Write([]byte(manyFileName(dirIndex, fileIndex)))
			h.Write([]byte{0})
			binary.LittleEndian.PutUint64(buf[:], uint64(sizes[fileNum]))
			h.Write(buf[:])
			fileNum++
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
