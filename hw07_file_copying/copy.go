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
	var bytesPerCopy int64

	if limit > 0 && limit < bytesPerCopyDefault {
		bytesPerCopy = limit
	} else {
		bytesPerCopy = bytesPerCopyDefault
	}

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

	var bytesCopiedCounter int64 = 0

	for {
		bytesCopied, err := io.CopyN(fileTo, fileFrom, bytesPerCopy)
		bytesCopiedCounter += bytesCopied

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if limit > 0 && bytesCopiedCounter == limit {
			break
		} else if limit != 0 && bytesCopiedCounter+bytesPerCopy > limit {
			bytesPerCopy = limit - bytesCopiedCounter
		}
	}

	return nil
}
