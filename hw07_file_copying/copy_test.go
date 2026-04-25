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
