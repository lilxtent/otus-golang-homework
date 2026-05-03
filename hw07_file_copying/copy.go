package main

import (
	"errors"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
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
		return err
	}

	fileSizeBytes := fileStats.Size()

	if fileSizeBytes == 0 {
		return ErrUnsupportedFile
	}

	if offset > fileSizeBytes {
		return ErrOffsetExceedsFileSize
	}

	byteToCopy := fileSizeBytes - offset
	if limit > 0 && limit < byteToCopy {
		byteToCopy = limit
	}

	if byteToCopy <= 0 {
		return nil
	}

	if _, err = fileFrom.Seek(offset, 0); err != nil {
		return err
	}

	fileTo, err := os.Create(toPath)
	if err != nil {
		return err
	}

	defer fileTo.Close()

	bar := pb.New64(byteToCopy)
	reader := bar.NewProxyReader(resolveReader(fileFrom, limit))

	bar.Start()
	defer bar.Finish()

	buffer := make([]byte, int(bytesPerCopyDefault))
	if _, err := io.CopyBuffer(fileTo, reader, buffer); err != nil {
		return err
	}

	return nil
}

func resolveReader(file *os.File, limit int64) io.Reader {
	if limit > 0 {
		return io.LimitReader(file, limit)
	}

	return io.Reader(file)
}
