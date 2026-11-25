package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Heliodex/tracker/load"
)

const sampleText = `Name: cool guy
NumChannels: 2
NumPatterns: 1
NumInstruments: 0
Speed: 6
Tempo: 125
LinearSlides: 1
Order: 0

Pattern 0
Rows: 1
Row 000: C-4 01 .. .. .. | ... .. .. .. ..
`

// buildTextOnly writes sampleText to dir as an XM file, for tests that need a module without instruments
func buildTextOnly(t *testing.T, dir string) string {
	t.Helper()

	in := filepath.Join(dir, "plain.txt")
	out := filepath.Join(dir, "plain.xm")

	if err := os.WriteFile(in, []byte(sampleText), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", in}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("build: %v", err)
	}

	return out
}

func TestExportTextOnlyToStdout(t *testing.T) {
	xm := buildTextOnly(t, t.TempDir())

	var stdout, stderr bytes.Buffer

	if err := run([]string{"export", xm}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("export: %v", err)
	}

	if !strings.Contains(stdout.String(), "NumChannels: 2\n") {
		t.Errorf("unexpected output:\n%s", stdout.String())
	}
}

func TestExportFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "song")

	var stdout, stderr bytes.Buffer

	if err := run([]string{"export", "-d", dir, "../../track3.xm"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("export: %v", err)
	}

	if stdout.Len() != 0 {
		t.Errorf("expected no stdout, got %d bytes", stdout.Len())
	}

	if _, err := os.Stat(filepath.Join(dir, "song.txt")); err != nil {
		t.Errorf("song.txt missing: %v", err)
	}

	matches, _ := filepath.Glob(filepath.Join(dir, "samples", "*.wav"))
	if len(matches) == 0 {
		t.Errorf("no WAV files written")
	}
}

func TestExportFolderRejectsOutputFile(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := run([]string{"export", "-o", filepath.Join(t.TempDir(), "x.txt"), "../../track3.xm"}, nil, &stdout, &stderr)
	if err == nil {
		t.Error("expected an error")
	}
}

func TestBuildFromFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "song")

	var stdout, stderr bytes.Buffer

	if err := run([]string{"export", "-d", dir, "../../track3.xm"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("export: %v", err)
	}

	if err := run([]string{"build", dir}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("build: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "song.xm"))
	if err != nil {
		t.Fatalf("song.xm not written: %v", err)
	}

	f, err := load.Read(bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("load.Read: %v", err)
	}

	if len(f.Instruments) != 3 {
		t.Errorf("got %d instruments, want 3", len(f.Instruments))
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

	if err := os.WriteFile(in, []byte(sampleText), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", in}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("first build: %v", err)
	}

	if err := run([]string{"build", in}, nil, &stdout, &stderr); err == nil {
		t.Error("expected an error without -f")
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
