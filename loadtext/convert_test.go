package loadtext

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Heliodex/tracker/load"
	"github.com/Heliodex/tracker/save"
)

func TestToFileRejectsBadPatternCount(t *testing.T) {
	tf := &TextFormat{
		NumPatterns: 2,
		NumChannels: 1,
	}

	if _, err := tf.ToFile(); err == nil {
		t.Error("expected an error")
	}
}

// TestXMPatternsSurviveLoad checks the cells read back from XM match the cells that were written
func TestXMPatternsSurviveLoad(t *testing.T) {
	tf, err := ReadText(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("ReadText: %v", err)
	}

	f, err := tf.ToFile()
	if err != nil {
		t.Fatalf("ToFile: %v", err)
	}

	var xm bytes.Buffer

	if err := save.Write(&xm, f); err != nil {
		t.Fatalf("save.Write: %v", err)
	}

	f2, err := load.Read(&xm)
	if err != nil {
		t.Fatalf("load.Read: %v", err)
	}

	if len(f2.Patterns) != len(f.Patterns) {
		t.Fatalf("got %d patterns, want %d", len(f2.Patterns), len(f.Patterns))
	}

	for i := range f.Patterns {
		for r := range f.Patterns[i].Data {
			for c := range f.Patterns[i].Data[r] {
				if got, want := f2.Patterns[i].Data[r][c], f.Patterns[i].Data[r][c]; got != want {
					t.Errorf("pattern %d row %d channel %d: got %+v, want %+v", i, r, c, got, want)
				}
			}
		}
	}
}

// TestXMRoundTrip checks text -> XM bytes -> text gives back the same text
func TestXMRoundTrip(t *testing.T) {
	tf, err := ReadText(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("ReadText: %v", err)
	}

	f, err := tf.ToFile()
	if err != nil {
		t.Fatalf("ToFile: %v", err)
	}

	var xm bytes.Buffer

	if err := save.Write(&xm, f); err != nil {
		t.Fatalf("save.Write: %v", err)
	}

	f2, err := load.Read(&xm)
	if err != nil {
		t.Fatalf("load.Read: %v", err)
	}

	tf2, err := FromFile(f2)
	if err != nil {
		t.Fatalf("FromFile: %v", err)
	}

	var text bytes.Buffer

	if err := WriteText(&text, tf2); err != nil {
		t.Fatalf("WriteText: %v", err)
	}

	if got := text.String(); got != sample {
		t.Errorf("output mismatch:\n%s\nwant:\n%s", got, sample)
	}
}
