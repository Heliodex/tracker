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
		NumChannels: 1,
		NumPatterns: 2,
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

// TestXMRoundTripInstruments checks instruments and samples survive text -> XM bytes -> text
func TestXMRoundTripInstruments(t *testing.T) {
	for name, sixteen := range map[string]bool{"16-bit": true, "8-bit": false} {
		t.Run(name, func(t *testing.T) {
			text := withInstrument()

			tf, err := ReadText(strings.NewReader(text))
			if err != nil {
				t.Fatalf("ReadText: %v", err)
			}

			s := &tf.Instruments[0].Samples[0]
			s.Sixteen = sixteen
			s.Data = make([]uint8, 200)

			for i := range s.Data {
				s.Data[i] = uint8(i * 7)
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

			if got := tf2.Instruments[0].Samples[0].Data; !bytes.Equal(got, s.Data) {
				t.Errorf("sample data changed in the round trip")
			}

			// the WAV path is not in the XM file, so the exporter sets it again
			tf2.Instruments[0].Samples[0].File = s.File

			var out bytes.Buffer

			if err := WriteText(&out, tf2); err != nil {
				t.Fatalf("WriteText: %v", err)
			}

			if got := out.String(); got != text {
				t.Errorf("output mismatch:\n%s\nwant:\n%s", got, text)
			}
		})
	}
}
