package song

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Heliodex/tracker/load"
)

const squareRecipe = "Frames: 256\nLayer: square 261.34375 0.5\n"

func TestTriad(t *testing.T) {
	tests := map[string][]string{
		"A": {"A-3", "C#4", "E-4"},
		"Am": {"A-3", "C-4", "E-4"},
		"C": {"C-3", "E-3", "G-3"},
		"F#m": {"F#3", "A-3", "C#4"},
	}

	for chord, want := range tests {
		if got := Triad(chord, 3); !reflect.DeepEqual(got, want) {
			t.Errorf("Triad(%q, 3) = %v, want %v", chord, got, want)
		}
	}
}

func TestTriadPanicsOnBadChord(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a panic")
		}
	}()

	Triad("H", 3)
}

func TestBadNoteIsReported(t *testing.T) {
	s := New("x", 1)
	ins := s.Instrument(Instrument{Name: "a", Recipe: squareRecipe})
	p := s.Pattern(4)
	p.Hit(0, 0, "H-4", ins)
	s.Play(p)

	if _, err := s.Build(); err == nil {
		t.Error("expected an error for a bad note")
	}
}

func TestBadRecipeIsReported(t *testing.T) {
	s := New("x", 1)
	s.Instrument(Instrument{Name: "a", Recipe: "Frames: nope\n"})

	if _, err := s.Build(); err == nil {
		t.Error("expected an error for a bad recipe")
	}
}

func TestOutOfRangeHitIsReported(t *testing.T) {
	s := New("x", 2)
	ins := s.Instrument(Instrument{Name: "a", Recipe: squareRecipe})
	p := s.Pattern(4)
	p.Hit(0, 2, "C-4", ins)

	if _, err := s.Build(); err == nil {
		t.Error("expected an error for a channel out of range")
	}
}

func TestWriteLoadsBack(t *testing.T) {
	s := New("demo", 2)
	s.Timing(4, 150)

	lead := s.Instrument(Instrument{Name: "lead", Recipe: squareRecipe, Loop: true, Envelope: []Point{{X: 0, Y: 64}, {X: 20, Y: 0}}, Pan: 64})
	kick := s.Instrument(Instrument{Name: "kick", Recipe: "Frames: 100\nLayer: sine 55 1\n"})

	p := s.Pattern(16)
	p.Line(0, 0, 4, Triad("Am", 4), lead, Vol(0x40))
	p.Hit(0, 1, "C-2", kick, Pan(0x80))
	p.Hit(8, 0, "###", nil)

	s.Play(p, p)

	path := filepath.Join(t.TempDir(), "demo.xm")

	if err := s.Write(path); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	f, err := load.Read(bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("load.Read: %v", err)
	}

	if f.Head.DefaultSpeed != 4 || f.Head.DefaultTempo != 150 {
		t.Errorf("got speed %d, tempo %d", f.Head.DefaultSpeed, f.Head.DefaultTempo)
	}

	if f.Head.SongLength != 2 || f.Head.OrderTable[0] != 0 || f.Head.OrderTable[1] != 0 {
		t.Errorf("got order length %d", f.Head.SongLength)
	}

	if len(f.Instruments) != 2 || len(f.Instruments[0].Samples) != 1 {
		t.Fatalf("got %d instruments", len(f.Instruments))
	}

	// the recipe is 8-bit, so loop lengths are in bytes, one per frame
	if s := f.Instruments[0].Samples[0]; s.LoopLength != 256 || s.Flags&0x03 != 1 {
		t.Errorf("got flags %#x, loop length %d, want forward loop of length 256", s.Flags, s.LoopLength)
	}
}

func TestHitReplacesTheCell(t *testing.T) {
	s := New("x", 1)
	ins := s.Instrument(Instrument{Name: "a", Recipe: squareRecipe})
	p := s.Pattern(2)
	p.Hit(0, 0, "C-4", ins, Pan(0x30))
	p.Hit(0, 0, "E-4", ins)

	if c := p.rows[0][0]; c.Effect != 0 || c.EffectParameter != 0 {
		t.Errorf("second hit kept the earlier effect: %+v", c)
	}
}
