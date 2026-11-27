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

func TestBuildFromFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "song")

	var stdout, stderr bytes.Buffer

	if err := run([]string{"export", "-d", dir, "../../track3.xm"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("export: %v", err)
	}

	if err := run([]string{"build", "-no-opus", dir}, nil, &stdout, &stderr); err != nil {
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

	if err := run([]string{"build", "-no-opus", "-o", "-", "-"}, strings.NewReader(sampleText), &stdout, &stderr); err != nil {
		t.Fatalf("build: %v", err)
	}

	if !bytes.HasPrefix(stdout.Bytes(), []byte("Extended Module: ")) {
		t.Errorf("output is not an XM header")
	}
}

func TestBuildNeedsOutputForStdin(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", "-no-opus", "-"}, strings.NewReader(sampleText), &stdout, &stderr); err == nil {
		t.Error("expected an error")
	}
}

func TestBuildNoOpusSkipsRender(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "plain.txt")

	if err := os.WriteFile(in, []byte(sampleText), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", "-no-opus", in}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("build: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "plain.opus")); err == nil {
		t.Errorf("plain.opus written despite -no-opus")
	}
}

func TestBuildRefusesToOverwrite(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "song.txt")

	if err := os.WriteFile(in, []byte(sampleText), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", "-no-opus", in}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("first build: %v", err)
	}

	if err := run([]string{"build", "-no-opus", in}, nil, &stdout, &stderr); err == nil {
		t.Error("expected an error without -f")
	}

	if err := run([]string{"build", "-no-opus", "-f", in}, nil, &stdout, &stderr); err != nil {
		t.Errorf("forced build: %v", err)
	}
}

func TestBuildSongGo(t *testing.T) {
	// track3-go is a song.go folder, which the program builds into song.xm
	const dir string = "../../tracks/track3-go"

	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", "-no-opus", dir}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("build: %v\n%s", err, stderr.String())
	}

	data, err := os.ReadFile(filepath.Join(dir, "song.xm"))
	if err != nil {
		t.Fatalf("song.xm not written: %v", err)
	}

	if _, err := load.Read(bytes.NewBuffer(data)); err != nil {
		t.Errorf("load.Read: %v", err)
	}
}

func TestBuildSongGoRejectsOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", "-no-opus", "-o", "x.xm", "../../tracks/track3-go"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected an error")
	}
}

func TestBuildWritesOpusAndNoWav(t *testing.T) {
	const want string = "plain.opus plain.txt plain.xm"

	if err := checkOpusTools(); err != nil {
		t.Skip(err)
	}

	dir := t.TempDir()
	in := filepath.Join(dir, "plain.txt")

	if err := os.WriteFile(in, []byte(sampleText), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", in}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("build: %v\n%s", err, stderr.String())
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	var names []string

	for _, e := range entries {
		names = append(names, e.Name())
	}

	if strings.Join(names, " ") != want {
		t.Errorf("got files %v, want %s", names, want)
	}

	data, err := os.ReadFile(filepath.Join(dir, "plain.opus"))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.HasPrefix(data, []byte("OggS")) {
		t.Errorf("plain.opus is not an Ogg stream")
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

	if matches, _ := filepath.Glob(filepath.Join(dir, "samples", "*.wav")); len(matches) == 0 {
		t.Errorf("no WAV files written")
	}
}

func TestExportFolderRejectsOutputFile(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := run([]string{"export", "-o", filepath.Join(t.TempDir(), "x.txt"), "../../track3.xm"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected an error")
	}
}

func TestNewSongCommand(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "mysong")

	var stdout, stderr bytes.Buffer

	if err := run([]string{"new", "-name", "demo", "-channels", "2", dir}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("new: %v", err)
	}

	if err := run([]string{"build", "-no-opus", dir}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("build: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "song.xm"))
	if err != nil {
		t.Fatalf("expected song.xm: %v", err)
	}

	f, err := load.Read(bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("load.Read: %v", err)
	}

	if f.Head.NumChannels != 2 {
		t.Errorf("got %d channels, want 2", f.Head.NumChannels)
	}
}

func TestRenderNeedsWavOutput(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "lead.txt")

	if err := os.WriteFile(in, []byte("Frames: 64\nLayer: sine 130 0.5\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	if err := run([]string{"render", "-o", filepath.Join(dir, "other.txt"), in}, nil, &stdout, &stderr); err == nil {
		t.Error("expected an error for a non-WAV output")
	}
}

func TestRenderRecipe(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "lead.txt")
	out := filepath.Join(dir, "preview.wav")

	if err := os.WriteFile(in, []byte("Frames: 64\nLayer: square 130 0.5\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	if err := run([]string{"render", "-o", out, in}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("render: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.HasPrefix(data, []byte("RIFF")) {
		t.Errorf("output is not a WAV file")
	}
}

func TestRenderRefusesBesideRecipe(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "lead.txt")

	if err := os.WriteFile(in, []byte("Frames: 64\nLayer: sine 130 0.5\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	if err := run([]string{"render", "-o", filepath.Join(dir, "lead.wav"), in}, nil, &stdout, &stderr); err == nil {
		t.Error("expected an error")
	}
}

func TestUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := run([]string{"bogus"}, nil, &stdout, &stderr); err == nil {
		t.Error("expected an error")
	}
}

// buildTextOnly writes sampleText to dir as an XM file, for tests that need a module without instruments
func buildTextOnly(t *testing.T, dir string) string {
	t.Helper()

	in := filepath.Join(dir, "plain.txt")
	out := filepath.Join(dir, "plain.xm")

	if err := os.WriteFile(in, []byte(sampleText), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	if err := run([]string{"build", "-no-opus", in}, nil, &stdout, &stderr); err != nil {
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
