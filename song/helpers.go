package song

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// At returns the row of row r in bar b, counting rows from the start of the pattern
func At(bar, row int) int {
	return bar*RowsPerBar + row
}

// VolOpts returns the volume option when vol is set, followed by any extra options. A zero volume sets none
func VolOpts(vol int, extra ...Opt) []Opt {
	var opts []Opt

	if vol != 0 {
		opts = append(opts, Vol(vol))
	}

	return append(opts, extra...)
}

// Steps returns the rows from up to, but not including, to, stepping by step
func Steps(from, to, step int) []int {
	var rows []int

	for r := from; r < to; r += step {
		rows = append(rows, r)
	}

	return rows
}

// Points returns envelope points from pairs of tick and volume, so Points(0, 0, 24, 64) is points at tick 0 with
// volume 0 and tick 24 with volume 64. It panics on an odd number of values, since that is a mistake in the code.
func Points(ticksAndVolumes ...int) []Point {
	if len(ticksAndVolumes)%2 != 0 {
		panic(fmt.Sprintf("song: Points needs tick and volume pairs, got %d values", len(ticksAndVolumes)))
	}

	points := make([]Point, 0, len(ticksAndVolumes)/2)

	for i := 0; i < len(ticksAndVolumes); i += 2 {
		points = append(points, Point{X: ticksAndVolumes[i], Y: ticksAndVolumes[i+1]})
	}

	return points
}

// WriteBeside writes the song as song.xm in the folder of the Go file that calls it, so a song.go writes its module
// next to itself, wherever it is run from
func (s *Song) WriteBeside() error {
	_, src, _, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("song: cannot find the calling file")
	}

	return s.Write(filepath.Join(filepath.Dir(src), "song.xm"))
}
