package project

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Heliodex/tracker/load"
	"github.com/Heliodex/tracker/loadtext"
	"github.com/Heliodex/tracker/save"
)

func TestBuildRejectsPathsOutsideFolder(t *testing.T) {
	text := `Name: x
NumChannels: 1
NumPatterns: 0
NumInstruments: 1
Speed: 6
Tempo: 125
LinearSlides: 1
Order:

Instrument 0
Name: x
Type: 0
Keymap: ` + strings.TrimSpace(strings.Repeat("0 ", 96)) + `
VolumeEnvelope:
VolumeSustain: 0
VolumeLoop: 0 0
VolumeFlags: 0
PanEnvelope:
PanSustain: 0
PanLoop: 0 0
PanFlags: 0
Vibrato: 0 0 0 0
VolumeFadeout: 0
Samples: 1

Sample 0
Name: x
File: ../secret.wav
Volume: 64
Finetune: 0
Panning: 128
RelativeNote: 0
LoopMode: 0
LoopStart: 0
LoopLength: 0
`

	if _, err := Build(strings.NewReader(text), t.TempDir()); err == nil {
		t.Error("expected an error for a path outside the folder")
	}
}

func TestSlug(t *testing.T) {
	tests := map[string]string{"": "sample", "  ": "sample", "Bass 2": "bass-2", "Kick Drum!": "kick-drum"}

	for in, want := range tests {
		if got := slug(in); got != want {
			t.Errorf("slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func loadModule(t *testing.T, path string) (data []byte) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func TestExportRefusesToOverwrite(t *testing.T) {
	orig, err := load.Read(bytes.NewBuffer(loadModule(t, "../track3.xm")))
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(t.TempDir(), "song")

	if err := Export(orig, dir, false); err != nil {
		t.Fatalf("first export: %v", err)
	}

	if err := Export(orig, dir, false); err == nil {
		t.Error("expected an error without force")
	}

	if err := Export(orig, dir, true); err != nil {
		t.Errorf("forced export: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func TestExportRefusesOverRecipe(t *testing.T) {
	orig, err := load.Read(bytes.NewBuffer(loadModule(t, "../track3.xm")))
	if err != nil {
		t.Fatal(err)
	}

	// export once to find the name the first sample is written to
	probe := filepath.Join(t.TempDir(), "probe")

	if err := Export(orig, probe, false); err != nil {
		t.Fatal(err)
	}

	tf, err := loadtext.ReadText(strings.NewReader(readFile(t, filepath.Join(probe, SongFile))))
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(t.TempDir(), "song")

	if err := os.MkdirAll(filepath.Join(dir, "samples"), 0o755); err != nil {
		t.Fatal(err)
	}

	recipePath := filepath.Join(dir, filepath.FromSlash(tf.Instruments[0].Samples[0].File)+".txt")

	if err := os.WriteFile(recipePath, []byte("Frames: 8\nLayer: sine 100 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Export(orig, dir, true); err == nil {
		t.Error("expected an error")
	}
}

// songWithSample returns a song.txt with one instrument whose sample refers to file
func songWithSample(file string) string {
	return `Name: x
NumChannels: 1
NumPatterns: 0
NumInstruments: 1
Speed: 6
Tempo: 125
LinearSlides: 1
Order:

Instrument 0
Name: x
Type: 0
Keymap: ` + strings.TrimSpace(strings.Repeat("0 ", 96)) + `
VolumeEnvelope:
VolumeSustain: 0
VolumeLoop: 0 0
VolumeFlags: 0
PanEnvelope:
PanSustain: 0
PanLoop: 0 0
PanFlags: 0
Vibrato: 0 0 0 0
VolumeFadeout: 0
Samples: 1

Sample 0
Name: x
File: ` + file + `
Volume: 64
Finetune: 0
Panning: 128
RelativeNote: 0
LoopMode: 0
LoopStart: 0
LoopLength: 0
`
}

func TestBuildAcceptsExtensionInReference(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "lead.txt"), []byte("Frames: 8\nLayer: sine 100 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(strings.NewReader(songWithSample("lead.txt")), dir); err != nil {
		t.Errorf("Build: %v", err)
	}
}

func TestBuildFailsWithNoSampleFile(t *testing.T) {
	if _, err := Build(strings.NewReader(songWithSample("samples/missing")), t.TempDir()); err == nil {
		t.Error("expected an error")
	}
}

func TestBuildRejectsWavAndRecipe(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, "samples"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "samples", "lead.txt"), []byte("Frames: 8\nLayer: sine 100 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "samples", "lead.wav"), []byte("not really a wav"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(strings.NewReader(songWithSample("samples/lead")), dir); err == nil {
		t.Error("expected an error when both a WAV and a recipe exist")
	}
}

func TestBuildRendersRecipe(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, "samples"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "samples", "lead.txt"), []byte("Frames: 64\nLayer: square 130 0.5\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	f, err := Build(strings.NewReader(songWithSample("samples/lead")), dir)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if got := len(f.Instruments[0].Samples[0].SampleData); got != 64 {
		t.Errorf("got %d bytes of sample data, want 64", got)
	}
}

// writeText returns the text form of tf, for comparing modules
func writeText(t *testing.T, tf *loadtext.TextFormat) string {
	t.Helper()

	var b bytes.Buffer

	if err := loadtext.WriteText(&b, tf); err != nil {
		t.Fatalf("WriteText: %v", err)
	}

	return b.String()
}

// TestExportBuildRoundTrip checks a module with instruments survives export to a folder and build back
func TestExportBuildRoundTrip(t *testing.T) {
	data := loadModule(t, "../track3.xm")

	orig, err := load.Read(bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("load.Read: %v", err)
	}

	dir := filepath.Join(t.TempDir(), "song")

	if err := Export(orig, dir, false); err != nil {
		t.Fatalf("Export: %v", err)
	}

	song, err := os.Open(filepath.Join(dir, SongFile))
	if err != nil {
		t.Fatal(err)
	}

	defer song.Close()

	exported, err := loadtext.ReadText(song)
	if err != nil {
		t.Fatalf("ReadText: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(exported.Instruments[0].Samples[0].File)+".wav")); err != nil {
		t.Errorf("sample file missing: %v", err)
	}

	f2, err := Build(strings.NewReader(readFile(t, filepath.Join(dir, SongFile))), dir)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if len(f2.Instruments) != len(orig.Instruments) {
		t.Fatalf("got %d instruments, want %d", len(f2.Instruments), len(orig.Instruments))
	}

	for i := range orig.Instruments {
		for s := range orig.Instruments[i].Samples {
			got := f2.Instruments[i].Samples[s].SampleData
			if want := orig.Instruments[i].Samples[s].SampleData; !bytes.Equal(got, want) {
				t.Errorf("instrument %d sample %d: sample data changed", i, s)
			}
		}
	}

	// the text must match what was exported, apart from the sample paths, which the XM file does not hold
	rebuilt, err := loadtext.FromFile(f2)
	if err != nil {
		t.Fatalf("FromFile: %v", err)
	}

	for i := range rebuilt.Instruments {
		for s := range rebuilt.Instruments[i].Samples {
			rebuilt.Instruments[i].Samples[s].File = exported.Instruments[i].Samples[s].File
		}
	}

	if got, want := writeText(t, rebuilt), readFile(t, filepath.Join(dir, SongFile)); got != want {
		t.Errorf("text changed in the round trip:\n%s\nwant:\n%s", got, want)
	}

	// and the rebuilt module is a valid XM file
	var xm bytes.Buffer

	if err := save.Write(&xm, f2); err != nil {
		t.Errorf("save.Write: %v", err)
	}
}
