package test

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/moby/buildkit-bench/util/testutil"
	"github.com/stretchr/testify/require"
)

func BenchmarkPackage(b *testing.B) {
	testutil.Run(b, testutil.BenchFuncs(
		benchmarkPackageSize,
	))
}

func benchmarkPackageSize(b *testing.B, sb testutil.Sandbox) {
	for i := 0; i < b.N; i++ {
		var packageSize int64
		err := filepath.Walk(path.Join(sb.BinsDir(), sb.Name()), func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				packageSize += info.Size()
			}
			return nil
		})
		require.NoError(b, err)
		testutil.ReportMetric(b, float64(packageSize), testutil.MetricBytes)
	}
}
