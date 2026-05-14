package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDir(t *testing.T) {
	env, err := ReadDir(filepath.Join("testdata", "env"))
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	expected := Environment{
		"HELLO": {Value: `"hello"`, NeedRemove: false},
		"BAR":   {Value: "bar", NeedRemove: false},
		"FOO":   {Value: "   foo\nwith new line", NeedRemove: false},
		"UNSET": {Value: "", NeedRemove: true},
		"EMPTY": {Value: "", NeedRemove: false},
	}

	if len(env) != len(expected) {
		t.Fatalf("unexpected env size: %d", len(env))
	}

	for key, expectedValue := range expected {
		value, ok := env[key]
		if !ok {
			t.Fatalf("missing env variable: %s", key)
		}

		if value != expectedValue {
			t.Fatalf("unexpected value for %s: %+v", key, value)
		}
	}
}

func TestReadDirInvalidName(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "INVALID=NAME"), []byte("value"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if _, err := ReadDir(dir); err == nil {
		t.Fatalf("expected ReadDir to return error for invalid filename")
	}
}
