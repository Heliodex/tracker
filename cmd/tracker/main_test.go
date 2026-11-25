package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleText = `Name: cool guy
NumChannels: 2
NumPatterns: 1
Order: 0

Pattern 0
Rows: 1
Row 000: C-4 01 .. .. .. | ... .. .. .. ..
`

func TestExportToStdout(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := run([]string{"export", "../../track3.xm"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("export: %v", err)
	}

	if !strings.Contains(stdout.String(), "NumChannels: 8\n") {
		t.Errorf("unexpected output:\n%s", stdout.String())
	}

	if !strings.Contains(stderr.String(), "3 instruments dropped") {
		t.Errorf("expected an instrument warning, got %q", stderr.String())
	}
}

func TestBuildFromStdinToStdout(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", "-o", "-", "-"}, strings.NewReader(sampleText), &stdout, &stderr); err != nil {
		t.Fatalf("build: %v", err)
	}

	if !bytes.HasPrefix(stdout.Bytes(), []byte("Extended Module: ")) {
		t.Errorf("output is not an XM header")
	}
}

func TestBuildNeedsOutputForStdin(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", "-"}, strings.NewReader(sampleText), &stdout, &stderr); err == nil {
		t.Error("expected an error")
	}
}

func TestBuildRefusesToOverwrite(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "song.txt")
	out := filepath.Join(dir, "song.xm")

	if err := os.WriteFile(in, []byte(sampleText), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", in}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("first build: %v", err)
	}

	if err := run([]string{"build", in}, nil, &stdout, &stderr); err == nil {
		t.Errorf("expected %s to refuse overwriting", out)
	}

	if err := run([]string{"build", "-f", in}, nil, &stdout, &stderr); err != nil {
		t.Errorf("forced build: %v", err)
	}
}

func TestUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := run([]string{"bogus"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected an error")
	}
}
