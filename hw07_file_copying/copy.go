package main

import (
	"errors"
	"io"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

const (
	bytesPerCopyDefault int64 = 256
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	fileFrom, err := os.Open(fromPath)
	if err != nil {
		return err
	}

	defer fileFrom.Close()

	fileStats, err := fileFrom.Stat()
	if err != nil {
		return ErrUnsupportedFile
	}

	fileSizeBytes := fileStats.Size()

	if offset > fileSizeBytes {
		return ErrOffsetExceedsFileSize
	}

	if _, err = fileFrom.Seek(offset, 0); err != nil {
		return err
	}

	fileTo, err := os.Create(toPath)
	if err != nil {
		return err
	}

	defer fileTo.Close()

	reader := io.Reader(fileFrom)
	if limit > 0 {
		reader = io.LimitReader(fileFrom, limit)
	}

	buffer := make([]byte, int(bytesPerCopyDefault))
	if _, err := io.CopyBuffer(fileTo, reader, buffer); err != nil {
		return err
	}

	return nil
}
