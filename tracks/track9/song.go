// Command track9 writes "glass harbour", a slow track in D minor, as an XM module, song.xm, in this folder.
//
// Build it from anywhere inside the repository with go run ./tracks/track9, or with tracker build.
package main

// glass harbour – Haiku 5.5

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/song"
)

type (
	// hit is one lead note at a row within the bar
	hit struct {
		row int
		note string
	}

	// kit holds the instruments, numbered as they are added to the song
	kit struct {
		pad, bass, lead, kick, snare, hat *song.Instrument
	}

	// pattern wraps a song pattern with the instruments to play it with
	pattern struct {
		p *song.Pattern
		k *kit
	}

	// section describes the layers of a four-bar pattern. A nil phrase plays no lead
	section struct {
		padVol, bassVol, drumStyle, leadVol int
		phrase [][]hit
		fill bool
	}
)

// channels, in the order they appear in the tracker
const (
	chPad = iota // three pad voices, chPad to chPad+2
	chBass = 3
	chKick = 4
	chSnare = 5
	chHat = 6
	chLead = 7
	numChannels = 8
)

const (
	rows = 64
	speed = 6
	tempo = 96
)

var (
	padRecipe = recipe.New(512).Triangle(261.34375, 0.16).Triangle(263.5, 0.12).Sine(130.671875, 0.08).String()
	bassRecipe = recipe.New(256).Square(130.671875, 0.12).Sine(130.671875, 0.22).Sine(65.3359375, 0.1).String()
	leadRecipe = recipe.New(256).Square(261.34375, 0.2).Duty(0.3).Sine(523.25, 0.06).String()
	kickRecipe = recipe.New(2000).Sine(50, 0.9).Sine(100, 0.2).Noise(2000, 0.03).String()
	snareRecipe = recipe.New(2400).Noise(900, 0.5).Triangle(200, 0.2).String()
	hatRecipe = recipe.New(400).Noise(4000, 0.35).String()

	// progression is i - VI - III - VII in D minor
	progression = []string{"Dm", "A#", "F", "C"}

	// verse is the question: it climbs to the top of the chord and falls back to the root
	verse = [][]hit{
		{{0, "A-4"}, {4, "D-5"}, {6, "F-5"}, {8, "E-5"}, {12, "D-5"}},
		{{0, "F-5"}, {4, "D-5"}, {8, "A#4"}, {10, "C-5"}, {12, "D-5"}},
		{{0, "A-4"}, {4, "C-5"}, {6, "F-5"}, {8, "A-5"}, {12, "G-5"}},
		{{0, "G-5"}, {4, "E-5"}, {6, "C-5"}, {8, "G-4"}, {12, "C-5"}, {14, "D-5"}},
	}

	// chorus is the answer: it starts high and settles on the chord tones
	chorus = [][]hit{
		{{0, "D-5"}, {2, "F-5"}, {4, "A-5"}, {8, "F-5"}, {12, "D-5"}},
		{{0, "F-5"}, {4, "A-5"}, {8, "D-5"}, {12, "C-5"}},
		{{0, "A-5"}, {4, "F-5"}, {8, "C-5"}, {10, "A-4"}, {12, "C-5"}},
		{{0, "G-5"}, {4, "C-5"}, {8, "E-5"}, {12, "G-5"}, {14, "A-5"}},
	}
)

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

// fill rolls the snare into the next pattern over the last beat of bar b
func (w pattern) fill(b int) {
	for i, r := range []int{12, 13, 14, 15} {
		w.p.Hit(at(b, r), chSnare, "C-4", w.k.snare, withVol(0x30+i*6)...)
	}
}

// melody plays a phrase of lead notes in bar b, each held until the next note
func (w pattern) melody(b int, phrase []hit, vol int) {
	if vol == 0 {
		return
	}

	for i, h := range phrase {
		w.p.Hit(at(b, h.row), chLead, h.note, w.k.lead, withVol(vol, song.Vibrato(4, 2))...)

		off := song.RowsPerBar - 1

		if i+1 < len(phrase) {
			off = phrase[i+1].row - 1
		}

		if off > h.row {
			w.p.Hit(at(b, off), chLead, "###", nil)
		}
	}
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

// arrange lays out a four-bar section, with the layers that section asks for
func (w pattern) arrange(sec section) {
	for b, chord := range progression {
		w.pads(b, chord, sec.padVol)
		w.bass(b, chord, sec.bassVol)
		w.drums(b, sec.drumStyle)

		if sec.phrase != nil {
			w.melody(b, sec.phrase[b], sec.leadVol)
		}
	}

	if sec.fill {
		w.fill(3)
	}
}

// newSection starts a pattern of four bars and lays out the section in it
func newSection(s *song.Song, k *kit, sec section) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(sec)

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
				{
					X: 0,
					Y: 0,
				},
				{
					X: 24,
					Y: 64,
				},
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
				{
					X: 0,
					Y: 0,
				},
				{
					X: 2,
					Y: 64,
				},
			},
		}),
		lead: s.Instrument(song.Instrument{
			Name: "lead",
			Recipe: leadRecipe,
			Loop: true,
			Volume: 34,
			Pan: 96,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 2,
					Y: 64,
				},
			},
			Fadeout: 512,
		}),
		kick: s.Instrument(song.Instrument{
			Name: "kick",
			Recipe: kickRecipe,
			Volume: 58,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 14,
					Y: 0,
				},
			},
		}),
		snare: s.Instrument(song.Instrument{
			Name: "snare",
			Recipe: snareRecipe,
			Volume: 44,
			Pan: 136,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 16,
					Y: 0,
				},
			},
		}),
		hat: s.Instrument(song.Instrument{
			Name: "hat",
			Recipe: hatRecipe,
			Volume: 30,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 3,
					Y: 0,
				},
			},
		}),
	}
	intro := newSection(s, k, section{
		padVol: 0x20,
	})
	verseP := newSection(s, k, section{
		padVol: 0x28,
		bassVol: 0x30,
		drumStyle: 1,
		leadVol: 0x34,
		phrase: verse,
	})
	chorusP := newSection(s, k, section{
		padVol: 0x30,
		bassVol: 0x38,
		drumStyle: 2,
		leadVol: 0x40,
		phrase: chorus,
		fill: true,
	})
	breakP := newSection(s, k, section{
		padVol: 0x24,
		leadVol: 0x30,
		phrase: verse,
	})
	outro := newSection(s, k, section{
		padVol: 0x20,
		bassVol: 0x28,
		drumStyle: 1,
		leadVol: 0x2C,
		phrase: chorus,
	})

	s.Play(intro, verseP, verseP, chorusP, breakP, chorusP, outro)

	_, src, _, _ := runtime.Caller(0)

	if err := s.Write(filepath.Join(filepath.Dir(src), "song.xm")); err != nil {
		log.Fatal(err)
	}
}
