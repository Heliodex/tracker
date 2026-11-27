package recipe

import (
	"fmt"
	"math"
)

// semitones gives each note letter's offset from C
var semitones = map[byte]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}

// Note returns the frequency of a note in Hz, snapped to a whole number of cycles over a loop of frames.
// A layer at that frequency loops cleanly, since it holds the same number of cycles as the sample.
//
// The name is a letter, '-' for a natural or '#' for a sharp, and an octave digit, so "C-4" is middle C and "F#3" is F sharp below it:
//
//	recipe.New(256).Square(recipe.Note(256, "C-4"), 0.2)
//
// The result is the nearest multiple of Rate / frames to the equal-tempered pitch, so it can differ slightly from it.
// Note panics on a bad name, since that is a mistake in the code.
func Note(frames int, name string) float64 {
	if len(name) != 3 || name[2] < '0' || name[2] > '9' || (name[1] != '-' && name[1] != '#') {
		panic(fmt.Sprintf("recipe: bad note name %q, want a letter, '-' or '#', and an octave like C-4", name))
	}

	semi, ok := semitones[name[0]]
	if !ok {
		panic(fmt.Sprintf("recipe: bad note name %q, unknown letter", name))
	}

	if name[1] == '#' {
		semi++
	}

	octave := int(name[2] - '0')
	midi := (octave+1)*12 + semi

	pitch := 440 * math.Pow(2, float64(midi-69)/12)

	return snap(frames, pitch)
}

// snap returns the nearest whole-cycle frequency over a loop of frames, and never less than one cycle
func snap(frames int, freq float64) float64 {
	cycles := max(math.Round(freq*float64(frames)/Rate), 1)

	return cycles * Rate / float64(frames)
}
