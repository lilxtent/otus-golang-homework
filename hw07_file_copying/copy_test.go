package main

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Run("Offset must be smaller then file size", func(t *testing.T) {
		maxInt64Value := math.MaxInt64
		err := Copy("testdata/input.txt", "out_test_tmp.txt", int64(maxInt64Value), 0)

		require.Error(t, ErrOffsetExceedsFileSize, err)
	})
}
