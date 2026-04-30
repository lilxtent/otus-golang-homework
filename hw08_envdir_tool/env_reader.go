package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	environment := Environment{}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}

		dirEntryName := dirEntry.Name()
		envFilePath := filepath.Join(dir, dirEntryName)

		envFile, err := os.Open(filepath.Clean(envFilePath))
		if err != nil {
			return nil, err
		}

		envFileReader := bufio.NewReader(envFile)

		firstLineBytes, isPrefix, err := envFileReader.ReadLine()
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}

		if isPrefix {
			return nil, errors.New("env line too long")
		}

		firstLineBytes = bytes.ReplaceAll(firstLineBytes, []byte{0x00}, []byte("\n"))
		firstLineString := string(firstLineBytes)
		needRemove := len(firstLineString) == 0
		firstLineString = strings.TrimRight(firstLineString, " \t")

		environment[dirEntryName] = EnvValue{
			Value:      firstLineString,
			NeedRemove: needRemove,
		}
	}

	return environment, nil
}
