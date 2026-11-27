// Command track9 writes "glass harbour", a slow track in D minor, as an XM module, song.xm, in this folder.
//
// Build it from anywhere inside the repository with go run ./tracks/track9, or with tracker build.
package main

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/song"
)

// channels, in the order they appear in the tracker
const (
	chPad = iota // three pad voices, chPad to chPad+2
	chBass = 3
	chKick = 4
	chSnare = 5
	chHat = 6
	numChannels = 7
)

const (
	rows = 64
	speed = 6
	tempo = 96
)

var (
	padRecipe = recipe.New(512).Sixteen().Triangle(261.34375, 0.16).Triangle(263.5, 0.12).Sine(130.671875, 0.08).String()
	bassRecipe = recipe.New(256).Sixteen().Square(130.671875, 0.12).Sine(130.671875, 0.22).Sine(65.3359375, 0.1).String()
	kickRecipe = recipe.New(2000).Sixteen().Sine(50, 0.9).Sine(100, 0.2).Noise(2000, 0.03).String()
	snareRecipe = recipe.New(2400).Sixteen().Noise(900, 0.5).Triangle(200, 0.2).String()
	hatRecipe = recipe.New(400).Sixteen().Noise(4000, 0.35).String()

	// progression is i - VI - III - VII in D minor
	progression = []string{"Dm", "A#", "F", "C"}
)

// kit holds the instruments, numbered as they are added to the song
type kit struct {
	pad, bass, kick, snare, hat *song.Instrument
}

// pattern wraps a song pattern with the instruments to play it with
type pattern struct {
	p *song.Pattern
	k *kit
}

// at returns the row of bar b, row r within the bar
func at(b, r int) int {
	return b*song.RowsPerBar + r
}

// withVol returns the options for a cell at volume vol, with any extra options
func withVol(vol int, extra ...song.Opt) []song.Opt {
	var opts []song.Opt

	if vol != 0 {
		opts = append(opts, song.Vol(vol))
	}

	return append(opts, extra...)
}

// pads holds the chord of bar b on the three pad voices, spread across the stereo field
func (w pattern) pads(b int, chord string, vol int) {
	if vol == 0 {
		return
	}

	pans := [3]int{0x50, 0x80, 0xB0}

	for i, note := range song.Triad(chord, 3) {
		w.p.Hit(at(b, 0), chPad+i, note, w.k.pad, withVol(vol, song.Pan(pans[i]))...)
	}
}

// bass plays the root of the chord on a syncopated figure, with the octave above on the off-beats
func (w pattern) bass(b int, chord string, vol int) {
	if vol == 0 {
		return
	}

	low := song.Triad(chord, 2)[0]
	high := song.Triad(chord, 3)[0]

	for _, r := range []int{0, 3, 6, 8, 11, 14} {
		note := low

		if r == 11 || r == 14 {
			note = high
		}

		w.p.Hit(at(b, r), chBass, note, w.k.bass, withVol(vol)...)
	}
}

// drums plays a kick, snare and closed hats. Style 0 is silent, 1 is a sparse groove and 2 is a full one
func (w pattern) drums(b, style int) {
	if style == 0 {
		return
	}

	kicks := []int{0, 10}

	if style == 2 {
		kicks = []int{0, 7, 10}
	}

	for _, r := range kicks {
		w.p.Hit(at(b, r), chKick, "C-2", w.k.kick, withVol(0x3C)...)
	}

	for _, r := range []int{4, 12} {
		w.p.Hit(at(b, r), chSnare, "C-4", w.k.snare, withVol(0x34)...)
	}

	for r := 0; r < song.RowsPerBar; r += 2 {
		vol := 0x1C

		if r%4 == 0 {
			vol = 0x28
		}

		w.p.Hit(at(b, r), chHat, "C-5", w.k.hat, withVol(vol, song.Pan(0x60+r*4))...)
	}
}

// section lays out one pattern of four bars, with a layer for each of the pads, bass and drums
func section(s *song.Song, k *kit, padVol, bassVol, drumStyle int) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, chord := range progression {
		w.pads(b, chord, padVol)
		w.bass(b, chord, bassVol)
		w.drums(b, drumStyle)
	}

	return w.p
}

func main() {
	s := song.New("glass harbour", numChannels)

	s.Timing(speed, tempo)

	k := &kit{
		pad: s.Instrument(song.Instrument{
			Name: "pad",
			Recipe: padRecipe,
			Loop: true,
			Volume: 28,
			Pan: 128,
			Envelope: []song.Point{
				{X: 0, Y: 0},
				{X: 24, Y: 64},
			},
			Fadeout: 1024,
		}),
		bass: s.Instrument(song.Instrument{
			Name: "bass",
			Recipe: bassRecipe,
			Loop: true,
			Volume: 48,
			Pan: 128,
			Envelope: []song.Point{
				{X: 0, Y: 0},
				{X: 2, Y: 64},
			},
		}),
		kick: s.Instrument(song.Instrument{
			Name: "kick",
			Recipe: kickRecipe,
			Volume: 58,
			Pan: 128,
			Envelope: []song.Point{
				{X: 0, Y: 64},
				{X: 14, Y: 0},
			},
		}),
		snare: s.Instrument(song.Instrument{
			Name: "snare",
			Recipe: snareRecipe,
			Volume: 44,
			Pan: 136,
			Envelope: []song.Point{
				{X: 0, Y: 64},
				{X: 16, Y: 0},
			},
		}),
		hat: s.Instrument(song.Instrument{
			Name: "hat",
			Recipe: hatRecipe,
			Volume: 30,
			Pan: 128,
			Envelope: []song.Point{
				{X: 0, Y: 64},
				{X: 3, Y: 0},
			},
		}),
	}

	intro := section(s, k, 0x24, 0, 0)
	groove := section(s, k, 0x2C, 0x30, 1)
	full := section(s, k, 0x30, 0x38, 2)
	outro := section(s, k, 0x20, 0x28, 0)

	s.Play(intro, groove, full, outro)

	_, src, _, _ := runtime.Caller(0)

	if err := s.Write(filepath.Join(filepath.Dir(src), "song.xm")); err != nil {
		log.Fatal(err)
	}
}
