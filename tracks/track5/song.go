// Command track5 writes "midnight coffee", a lo-fi hip-hop track, as an XM module, song.xm, in this folder.
//
// Build it from anywhere inside the repository with go run ./tracks/track5, or with tracker build.
package main

// midnight coffee – Muse Spark 1.3

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/song"
)

type (
	// chordBar is one bar of the loop: a jazzy voicing over a bass root
	chordBar struct {
		name string
		// lo and hi are the two low and two high notes of the Rhodes voicing
		lo, hi [2]string
		// bass is the root, plus the fifth and octave the bassline moves to
		bass, fifth, octave string
	}

	// hit is one note: a row within the bar, and the note name
	hit struct {
		row int
		note string
	}

	// kit holds the instruments, numbered as they are added to the song
	kit struct {
		lead, keys, bass, kick, snare, hat, openHat, vinyl *song.Instrument
	}

	// pattern wraps a song pattern with the instruments to play it with
	pattern struct {
		p *song.Pattern
		k *kit
	}
)

// channels, in the order they appear in the tracker
const (
	chLead = iota
	chKeysLo
	chKeysHi
	chBass
	chKick
	chSnare
	chHat
	chVinyl
)

const (
	rows = 64
	// swung 8ths behind a laid-back groove
	speed = 6
	tempo = 84
)

var (
	// soft sine lead for sleepy melodies, tuned to C-4
	leadRecipe = recipe.New(256).Sixteen().
		Triangle(261.34375, 0.155).
		Sine(522.6875, 0.05).String()

	// dusty rhodes-ish keys, tuned to C-4
	keysRecipe = recipe.New(256).Sixteen().
		Triangle(261.34375, 0.14).
		Sine(261.34375, 0.10).
		Sine(522.6875, 0.03).String()

	// round sub bass, tuned to C-4
	bassRecipe = recipe.New(256).Sixteen().
		Sine(261.34375, 0.22).
		Triangle(261.34375, 0.12).String()

	// soft lo-fi thump
	kickRecipe = recipe.New(2000).Sixteen().
		Sine(55, 0.32).
		Sine(110, 0.07).
		Noise(3000, 0.02).String()

	// dusty noise snare with a little body
	snareRecipe = recipe.New(3000).Sixteen().
		Noise(1500, 0.22).
		Triangle(185, 0.09).String()

	// short closed hat
	hatRecipe = recipe.New(800).Sixteen().
		Noise(1, 0.28).
		Square(4100, 0.03).String()

	// open hat, longer tail
	openHatRecipe = recipe.New(2000).Sixteen().
		Noise(1, 0.28).
		Square(4100, 0.03).String()

	// vinyl crackle bed, loops under the whole song
	vinylRecipe = recipe.New(8000).Sixteen().
		Noise(1, 0.5).
		Noise(3000, 0.12).String()
)

var (
	// the loop: Dm9 G13 Cmaj9 A7, a ii-V-I-VI in C with smooth voice leading
	loop = []chordBar{
		{
			name: "Dm9",
			lo: [2]string{"D-3", "F-3"},
			hi: [2]string{"C-4", "E-4"},
			bass: "D-2",
			fifth: "A-2",
			octave: "D-3",
		},
		{
			name: "G13",
			lo: [2]string{"D-3", "F-3"},
			hi: [2]string{"B-3", "E-4"},
			bass: "G-1",
			fifth: "D-2",
			octave: "G-2",
		},
		{
			name: "Cmaj9",
			lo: [2]string{"E-3", "G-3"},
			hi: [2]string{"B-3", "D-4"},
			bass: "C-2",
			fifth: "G-2",
			octave: "C-3",
		},
		{
			name: "A7",
			lo: [2]string{"C#3", "G-3"},
			hi: [2]string{"B-3", "E-4"},
			bass: "A-1",
			fifth: "E-2",
			octave: "A-2",
		},
	}
	verseLead = [][]hit{
		{{0, "A-4"}, {6, "C-5"}, {10, "D-5"}},
		{{0, "B-4"}, {8, "D-5"}},
		{{0, "E-5"}, {6, "D-5"}, {10, "C-5"}},
		{{0, "C#5"}, {8, "B-4"}, {12, "A-4"}},
	}
	hookLead = [][]hit{
		{{0, "E-5"}, {4, "D-5"}, {6, "C-5"}, {10, "A-4"}},
		{{0, "G-4"}, {4, "B-4"}, {8, "D-5"}, {12, "E-5"}},
		{{0, "E-5"}, {4, "G-5"}, {8, "E-5"}, {12, "D-5"}},
		{{0, "C#5"}, {4, "D-5"}, {8, "E-5"}, {12, "B-4"}},
	}
	breakLead = [][]hit{
		{{0, "D-5"}, {8, "C-5"}},
		{{0, "B-4"}, {8, "A-4"}},
		{{0, "G-4"}, {8, "B-4"}},
		{{0, "A-4"}, {4, "B-4"}, {8, "C-5"}, {12, "B-4"}},
	}
	outroLead = [][]hit{{{0, "A-4"}, {6, "C-5"}, {10, "D-5"}}, {{0, "E-5"}}, {}, {}}
	verseBassMoves = map[int]string{0: "root", 7: "root", 10: "fifth"}
	hookBassMoves = map[int]string{0: "root", 3: "octave", 7: "root", 10: "fifth", 12: "octave"}
	breakBassMoves = map[int]string{0: "root", 10: "fifth"}
	verseKicks = []int{0, 7, 10}
	hookKicks = []int{0, 7, 10, 14}
)

// at returns the row at offset r within bar b
func at(b, r int) int {
	return b*song.RowsPerBar + r
}

// rangeStep returns from, from+step, ... below to
func rangeStep(from, to, step int) (out []int) {
	for r := from; r < to; r += step {
		out = append(out, r)
	}

	return
}

// withVol returns the options for a volume column value, which is left out when zero, followed by extra
func withVol(vol int, extra ...song.Opt) []song.Opt {
	var opts []song.Opt

	if vol != 0 {
		opts = append(opts, song.Vol(vol))
	}

	return append(opts, extra...)
}

// bassline plays the bass movement for a bar: rows mapped to root, fifth or octave
func (w pattern) bassline(b int, cb chordBar, moves map[int]string, vol int) {
	for r, which := range moves {
		var note string

		switch which {
		case "root":
			note = cb.bass
		case "fifth":
			note = cb.fifth
		case "octave":
			note = cb.octave
		}

		w.p.Hit(at(b, r), chBass, note, w.k.bass, withVol(vol)...)
	}
}

// drums plays the boom-bap for a bar: laid-back kick, backbeat snare, swung hats
func (w pattern) drums(b int, kicks []int, ghost, hats bool) {
	for _, r := range kicks {
		w.p.Hit(at(b, r), chKick, "C-2", w.k.kick)
	}

	w.p.Hit(at(b, 4), chSnare, "C-4", w.k.snare, withVol(0x40)...)
	w.p.Hit(at(b, 12), chSnare, "C-4", w.k.snare, withVol(0x40)...)

	if ghost {
		w.p.Hit(at(b, 15), chSnare, "C-4", w.k.snare, withVol(0x18)...)
	}

	if !hats {
		return
	}

	// swung 8ths: offbeats pushed a row late, accents on the beats
	for i, r := range []int{0, 3, 4, 7, 8, 11, 12, 15} {
		vol := 0x20

		if r%4 == 0 {
			vol = 0x2C
		}

		pan := 0xD0

		if i%2 == 0 {
			pan = 0x30
		}

		w.p.Hit(at(b, r), chHat, "C-5", w.k.hat, withVol(vol, song.Pan(pan))...)
	}
}

// keys plays the Rhodes voicing for a bar, with a soft anticipation stab on the swung "and" of four
func (w pattern) keys(b int, cb chordBar, vol, stabVol int) {
	for _, note := range cb.lo {
		w.p.Hit(at(b, 0), chKeysLo, note, w.k.keys, withVol(vol, song.Pan(0x60))...)
	}

	for _, note := range cb.hi {
		w.p.Hit(at(b, 0), chKeysHi, note, w.k.keys, withVol(vol, song.Pan(0xA0))...)
	}

	if stabVol == 0 {
		return
	}

	for _, note := range cb.lo {
		w.p.Hit(at(b, 14), chKeysLo, note, w.k.keys, withVol(stabVol, song.Pan(0x60))...)
	}

	for _, note := range cb.hi {
		w.p.Hit(at(b, 14), chKeysHi, note, w.k.keys, withVol(stabVol, song.Pan(0xA0))...)
	}
}

// openHats breathes on the swung offbeats of the hook
func (w pattern) openHats(b int) {
	for i, r := range []int{7, 15} {
		pan := 0xD0

		if i%2 == 0 {
			pan = 0x30
		}

		w.p.Hit(at(b, r), chHat, "C-5", w.k.openHat, withVol(0x28, song.Pan(pan))...)
	}
}

// play puts a list of hits on a channel, one per row within the bar they name
func (w pattern) play(ch, b int, hits []hit, ins *song.Instrument, vol int, extra ...song.Opt) {
	for _, h := range hits {
		w.p.Hit(at(b, h.row), ch, h.note, ins, withVol(vol, extra...)...)
	}
}

// breakdown strips back to keys and a soft lead, half-time drums joining halfway
func breakdown(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	// the hook's lead rings no further
	w.p.Hit(0, chLead, "###", nil)

	for b, cb := range loop {
		w.keys(b, cb, 0x34, 0)
		w.bassline(b, cb, breakBassMoves, 0x3C)
		w.play(chLead, b, breakLead[b], k.lead, 0x40)

		if b >= 2 {
			w.drums(b, verseKicks, false, b == 3)
		}
	}

	// a snare pickup back into the groove
	for i, vol := range []int{0x28, 0x34} {
		w.p.Hit(at(3, 14+i), chSnare, "C-4", k.snare, withVol(vol)...)
	}

	return w.p
}

// hook is the fullest section: busier bass and kick, open hats, singing lead
func hook(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range loop {
		w.keys(b, cb, 0x3C, 0x28)
		w.bassline(b, cb, hookBassMoves, 0x48)
		w.drums(b, hookKicks, true, true)
		w.openHats(b)
		w.play(chLead, b, hookLead[b], k.lead, 0x4C)
	}

	return w.p
}

// outro thins out to keys and a final held E, then lets the vinyl ring out alone
func outro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range loop[:2] {
		w.keys(b, cb, 0x34, 0)
		w.bassline(b, cb, breakBassMoves, 0x3C)
		w.play(chLead, b, outroLead[b], k.lead, 0x40)
		w.drums(b, verseKicks, false, b == 1)
	}

	// the last keys hit, held while everything else stops
	cb := loop[2]

	w.keys(2, cb, 0x30, 0)

	for _, ch := range []int{chLead, chBass, chKick, chSnare, chHat} {
		w.p.Hit(at(2, 0), ch, "###", nil)
	}

	for _, ch := range []int{chKeysLo, chKeysHi} {
		w.p.Hit(rows-2, ch, "###", nil)
	}

	return w.p
}

// vinyl starts the crackle bed, which loops until the outro lets it ring out
func (w pattern) vinyl() {
	w.p.Hit(0, chVinyl, "C-4", w.k.vinyl, withVol(0x1A)...)
}

// intro opens with keys and vinyl alone, hats joining in the last bar
func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.vinyl()

	for b, cb := range loop {
		w.keys(b, cb, 0x34, 0)

		if b == 3 {
			w.drums(b, verseKicks, false, true)
		}
	}

	return w.p
}

// verseA is the groove settling in: keys, bass and drums, no lead yet
func verseA(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range loop {
		w.keys(b, cb, 0x38, 0x24)
		w.bassline(b, cb, verseBassMoves, 0x44)
		w.drums(b, verseKicks, false, true)
	}

	return w.p
}

// verseB adds the sleepy lead line over the groove
func verseB(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range loop {
		w.keys(b, cb, 0x38, 0x24)
		w.bassline(b, cb, verseBassMoves, 0x44)
		w.drums(b, verseKicks, true, true)
		w.play(chLead, b, verseLead[b], k.lead, 0x48)
	}

	return w.p
}

// verseC keeps the hook's melody over the sparser verse groove
func verseC(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range loop {
		w.keys(b, cb, 0x38, 0x24)
		w.bassline(b, cb, verseBassMoves, 0x44)
		w.drums(b, verseKicks, true, true)
		w.play(chLead, b, hookLead[b], k.lead, 0x40)
	}

	return w.p
}

func main() {
	s := song.New("midnight coffee", 8)

	s.Timing(speed, tempo)

	k := &kit{
		lead: s.Instrument(song.Instrument{
			Name: "soft lead",
			Recipe: leadRecipe,
			Loop: true,
			Volume: 48,
			Pan: 120,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 6,
					Y: 64,
				},
			},
		}),
		keys: s.Instrument(song.Instrument{
			Name: "rhodes",
			Recipe: keysRecipe,
			Loop: true,
			Volume: 56,
			Pan: 128,
			Fadeout: 1024,
		}),
		bass: s.Instrument(song.Instrument{
			Name: "sub bass",
			Recipe: bassRecipe,
			Loop: true,
			Volume: 60,
			Pan: 128,
		}),
		kick: s.Instrument(song.Instrument{
			Name: "kick",
			Recipe: kickRecipe,
			Volume: 60,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 10,
					Y: 0,
				},
			},
		}),
		snare: s.Instrument(song.Instrument{
			Name: "snare",
			Recipe: snareRecipe,
			Volume: 56,
			Pan: 136,
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
		hat: s.Instrument(song.Instrument{
			Name: "hat",
			Recipe: hatRecipe,
			Volume: 32,
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
		openHat: s.Instrument(song.Instrument{
			Name: "open hat",
			Recipe: openHatRecipe,
			Volume: 30,
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
		vinyl: s.Instrument(song.Instrument{
			Name: "vinyl",
			Recipe: vinylRecipe,
			Loop: true,
			Volume: 3,
			Pan: 128,
		}),
	}
	introP := intro(s, k)
	verseAP := verseA(s, k)
	verseBP := verseB(s, k)
	hookP := hook(s, k)
	breakP := breakdown(s, k)
	verseCP := verseC(s, k)
	outroP := outro(s, k)

	s.Play(introP, verseAP, verseBP, hookP, breakP, verseCP, hookP, outroP)

	// write next to this file, wherever the program is run from
	_, src, _, _ := runtime.Caller(0)

	if err := s.Write(filepath.Join(filepath.Dir(src), "song.xm")); err != nil {
		log.Fatal(err)
	}
}
