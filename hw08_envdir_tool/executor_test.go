package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCmd(t *testing.T) {
	t.Setenv("HELLO", "SHOULD_REPLACE")
	t.Setenv("FOO", "SHOULD_REPLACE")
	t.Setenv("UNSET", "SHOULD_REMOVE")
	t.Setenv("ADDED", "from original env")
	t.Setenv("EMPTY", "SHOULD_BE_EMPTY")

	env, err := ReadDir(filepath.Join("testdata", "env"))
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	cmd := []string{
		"/bin/bash",
		filepath.Join("testdata", "echo.sh"),
		"arg1=1",
		"arg2=2",
	}

	origStdout := os.Stdout
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	os.Stderr = w
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	returnCode := RunCmd(cmd, env)

	_ = w.Close()
	outputBytes, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	_ = r.Close()

	if returnCode != 0 {
		t.Fatalf("unexpected return code: %d", returnCode)
	}

	result := strings.TrimRight(string(outputBytes), "\n")
	expected := `HELLO is ("hello")
BAR is (bar)
FOO is (   foo
with new line)
UNSET is ()
ADDED is (from original env)
EMPTY is ()
arguments are arg1=1 arg2=2`

	if result != expected {
		t.Fatalf("invalid output:\n%s", result)
	}
}
