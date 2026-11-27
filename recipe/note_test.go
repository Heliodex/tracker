package recipe

import (
	"math"
	"testing"
)

func TestNoteBadNamePanics(t *testing.T) {
	for _, name := range []string{"", "C4", "H-4", "C-", "Cb4", "C-x"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Note(256, %q): expected a panic", name)
				}
			}()

			Note(256, name)
		}()
	}
}

func TestNoteIsWholeCycles(t *testing.T) {
	for _, frames := range []int{64, 256, 400, 512, 1600, 2000} {
		for _, name := range []string{"C-2", "C#3", "A-4", "G-5", "B-6"} {
			f := Note(frames, name)

			if cycles := f * float64(frames) / Rate; math.Abs(cycles-math.Round(cycles)) > 1e-9 || cycles < 1 {
				t.Errorf("Note(%d, %q) = %v, which is %v cycles", frames, name, f, cycles)
			}
		}
	}
}

func TestNoteSnapsToLoop(t *testing.T) {
	tests := []struct {
		frames int
		name string
		want float64
	}{
		{256, "C-4", 261.34375},
		{256, "C-3", 130.671875},
		{256, "C-5", 522.6875},
		{256, "G-5", 784.03125},
		{256, "C-2", 65.3359375},
		{512, "C-4", 261.34375},
		{2048, "C-4", 261.34375},
	}

	for _, tt := range tests {
		if got := Note(tt.frames, tt.name); got != tt.want {
			t.Errorf("Note(%d, %q) = %v, want %v", tt.frames, tt.name, got, tt.want)
		}
	}
}

func TestNoteWithinHalfCycleOfPitch(t *testing.T) {
	// A-4 is 440 Hz, and the snapped value must be within half a cycle of it
	const frames = 256

	half := Rate / float64(frames) / 2
	if got := Note(frames, "A-4"); got < 440-half || got > 440+half {
		t.Errorf("Note(%d, A-4) = %v, want within %v of 440", frames, got, half)
	}
}
