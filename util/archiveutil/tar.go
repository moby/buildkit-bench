package archiveutil

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
)

const (
	manyFileDirs          = 100
	manyFilesPerDir       = 100
	manyFileMinSizeBytes  = 10 * 1024
	manyFileMaxSizeBytes  = 2 * 1024 * 1024
	manyFileMaxTotalBytes = 512 * 1024 * 1024
	LargeFileSizeBytes    = 512 * 1024 * 1024
)

func WriteManyFilesTar(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}

	tw := tar.NewWriter(f)
	if err := writeManyFilesTar(tw); err != nil {
		_ = tw.Close()
		_ = f.Close()
		return err
	}
	if err := tw.Close(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func WriteLargeFileTar(name string, size int64) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}

	tw := tar.NewWriter(f)
	if err := writeTarDir(tw, "payload"); err != nil {
		_ = tw.Close()
		_ = f.Close()
		return err
	}
	if err := writeTarFile(tw, "payload/blob.bin", size, zeroReader{}); err != nil {
		_ = tw.Close()
		_ = f.Close()
		return err
	}
	if err := tw.Close(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func ManyFilesLastFileName() string {
	return manyFileName(manyFileDirs-1, manyFilesPerDir-1)
}

func writeManyFilesTar(tw *tar.Writer) error {
	if err := writeTarDir(tw, "payload"); err != nil {
		return err
	}
	sizes, err := manyFileSizes()
	if err != nil {
		return err
	}
	fileNum := 0
	for dirIndex := 0; dirIndex < manyFileDirs; dirIndex++ {
		dirName := fmt.Sprintf("payload/dir-%03d", dirIndex)
		if err := writeTarDir(tw, dirName); err != nil {
			return err
		}
		for fileIndex := 0; fileIndex < manyFilesPerDir; fileIndex++ {
			fileName := manyFileName(dirIndex, fileIndex)
			if err := writeTarFile(tw, fileName, sizes[fileNum], zeroReader{}); err != nil {
				return err
			}
			fileNum++
		}
	}
	return nil
}

func manyFileSizes() ([]int64, error) {
	totalFiles := manyFileDirs * manyFilesPerDir
	if int64(totalFiles*manyFileMinSizeBytes) > int64(manyFileMaxTotalBytes) {
		return nil, fmt.Errorf("minimum file sizes exceed total budget")
	}

	sizes := make([]int64, 0, totalFiles)
	remainingBudget := int64(manyFileMaxTotalBytes)
	for fileNum := 0; fileNum < totalFiles; fileNum++ {
		remainingFiles := totalFiles - fileNum
		maxSize := remainingBudget - int64(remainingFiles-1)*manyFileMinSizeBytes
		if maxSize < int64(manyFileMinSizeBytes) {
			return nil, fmt.Errorf("remaining budget below minimum file size")
		}
		maxSize = min(maxSize, int64(manyFileMaxSizeBytes))

		size := manyFileSize(fileNum)
		size = min(size, maxSize)
		sizes = append(sizes, size)
		remainingBudget -= size
	}
	if sumInt64(sizes) > int64(manyFileMaxTotalBytes) {
		return nil, fmt.Errorf("file sizes exceed total budget")
	}
	return sizes, nil
}

func manyFileName(dirIndex, fileIndex int) string {
	return fmt.Sprintf("payload/dir-%03d/file-%03d.txt", dirIndex, fileIndex)
}

func manyFileSize(fileNum int) int64 {
	hash := manyFileHash(uint64(fileNum))
	switch bucket := hash % 100; {
	case bucket == 0:
		return int64Range(hash>>8, 512*1024, manyFileMaxSizeBytes)
	case bucket < 9:
		return int64Range(hash>>8, 64*1024, 512*1024)
	default:
		return int64Range(hash>>8, manyFileMinSizeBytes, 32*1024)
	}
}

// Keep the many-file ADD payload independent from math/rand implementation changes.
func manyFileHash(v uint64) uint64 {
	v += 0x9e3779b97f4a7c15
	v = (v ^ (v >> 30)) * 0xbf58476d1ce4e5b9
	v = (v ^ (v >> 27)) * 0x94d049bb133111eb
	return v ^ (v >> 31)
}

func int64Range(v uint64, minValue, maxValue int64) int64 {
	if maxValue <= minValue {
		return minValue
	}
	return minValue + int64(v%uint64(maxValue-minValue+1))
}

func sumInt64(values []int64) int64 {
	var total int64
	for _, value := range values {
		total += value
	}
	return total
}

func writeTarDir(tw *tar.Writer, name string) error {
	return tw.WriteHeader(&tar.Header{
		Name:     name,
		Typeflag: tar.TypeDir,
		Mode:     0755,
	})
}

func writeTarFile(tw *tar.Writer, name string, size int64, r io.Reader) error {
	if err := tw.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0644,
		Size: size,
	}); err != nil {
		return err
	}

	_, err := io.CopyN(tw, r, size)
	return err
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}
