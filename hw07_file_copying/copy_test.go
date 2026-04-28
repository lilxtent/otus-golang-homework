package main

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Run("Offset must be smaller then file size", func(t *testing.T) {
		maxInt64Value := math.MaxInt64
		err := Copy("testdata/input.txt", "out_test_tmp.txt", int64(maxInt64Value), 0)

		require.Error(t, ErrOffsetExceedsFileSize, err)
	})

	t.Run("Limit is greater than file size", func(t *testing.T) {
		tmpDir := t.TempDir()
		fromPath := filepath.Join(tmpDir, "input.txt")
		toPath := filepath.Join(tmpDir, "out.txt")

		data := []byte("hello world")
		require.NoError(t, os.WriteFile(fromPath, data, 0o644))

		limit := int64(len(data) + 10)
		err := Copy(fromPath, toPath, 0, limit)

		require.NoError(t, err)

		out, err := os.ReadFile(toPath)
		require.NoError(t, err)
		require.Equal(t, data, out)
	})
}

func TestCopyShouldBeSucces(t *testing.T) {
	testCases := []struct {
		name         string
		offset       int64
		limit        int64
		expectedPath string
	}{
		{
			name:         "offset0_limit0",
			offset:       0,
			limit:        0,
			expectedPath: "testdata/out_offset0_limit0.txt",
		},
		{
			name:         "offset0_limit10",
			offset:       0,
			limit:        10,
			expectedPath: "testdata/out_offset0_limit10.txt",
		},
		{
			name:         "offset0_limit1000",
			offset:       0,
			limit:        1000,
			expectedPath: "testdata/out_offset0_limit1000.txt",
		},
		{
			name:         "offset0_limit10000",
			offset:       0,
			limit:        10000,
			expectedPath: "testdata/out_offset0_limit10000.txt",
		},
		{
			name:         "offset100_limit1000",
			offset:       100,
			limit:        1000,
			expectedPath: "testdata/out_offset100_limit1000.txt",
		},
		{
			name:         "offset6000_limit1000",
			offset:       6000,
			limit:        1000,
			expectedPath: "testdata/out_offset6000_limit1000.txt",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outPath := filepath.Join(tmpDir, "out.txt")

			err := Copy("testdata/input.txt", outPath, testCase.offset, testCase.limit)
			require.NoError(t, err)

			got, err := os.ReadFile(outPath)
			require.NoError(t, err)

			resultCopyFileContent, err := os.ReadFile(testCase.expectedPath)
			require.NoError(t, err)

			require.Equal(t, resultCopyFileContent, got)
		})
	}
}
