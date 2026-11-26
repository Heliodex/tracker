package song

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Heliodex/tracker/types"
)

// parseChord splits a chord name such as "A", "Am", "F#" or "F#m" into its root tone and whether it is minor
func parseChord(name string) (tone string, minor bool, err error) {
	if name == "" || name[0] < 'A' || name[0] > 'G' {
		return "", false, fmt.Errorf("unknown chord %q", name)
	}

	letter := name[:1]
	rest := name[1:]

	if strings.HasPrefix(rest, "#") {
		tone = letter + "#"
		rest = rest[1:]
	} else {
		tone = letter + "-"
	}

	switch rest {
	case "":
		return tone, false, nil
	case "m":
		return tone, true, nil
	}

	return "", false, fmt.Errorf("unknown chord %q", name)
}

// shift returns the note semis half steps from n, panicking if that leaves the tracker's note range
func shift(n types.Note, semis int) string {
	v := int(n) + semis
	if v < 1 || v > 96 {
		panic(fmt.Sprintf("note %s shifted by %d is outside the tracker's range", n, semis))
	}

	return types.Note(v).String()
}

// Triad returns the notes of a major or minor triad, with its root in the given octave.
// "A" is A major and "Am" is A minor, and sharps are written "F#" and "F#m".
// Triad panics on an unknown chord name, because songs written in code should be fixed when they are written.
func Triad(chord string, octave int) []string {
	tone, minor, err := parseChord(chord)
	if err != nil {
		panic(err)
	}

	root, err := types.ParseNote(tone + strconv.Itoa(octave))
	if err != nil {
		panic(err)
	}

	third := 4

	if minor {
		third = 3
	}

	return []string{shift(root, 0), shift(root, third), shift(root, 7)}
}
