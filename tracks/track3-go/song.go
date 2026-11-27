// Command track3-go writes the night drive track as an XM module, song.xm, in this folder.
//
// Run it from anywhere inside the repository with go run ./tracks/track3-go, or with tracker build.
package main

// night drive – DeepSeek V4.1 Flash

import (
	"log"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/song"
)

type (
	// chordBar is one bar of a chord progression
	chordBar struct {
		chord string
		padOct int
		// bass is the bass note for the whole bar, and bassTwo the two eighth notes when the bass plays eighths
		bass string
		bassTwo [2]string
		lead, arp, saw []string
	}

	// kit holds the instruments, numbered as they are added to the song
	kit struct {
		lead, bass, kick, snare, hat, pad, crash, bell, saw, warm, tom, openHat *song.Instrument
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
	chBass
	chDrum
	chHat
	chPadLo
	chPadMid
	chPadHi
	chArp
)

const rows = 64

var (
	// saw bass with a sine sub, tuned to C-4
	bassRecipe = recipe.New(256).Saw(recipe.Note(256, "C-4"), 0.45).Sine(recipe.Note(256, "C-3"), 0.35).String()

	// bell-like pluck, tuned to C-4
	bellRecipe = recipe.New(2000).Sine(recipe.Note(2000, "C-4"), 0.5).Sine(783, 0.18).Sine(1306.7, 0.08).String()

	// long bright noise crash
	crashRecipe = recipe.New(8000).Noise(1000, 1).Noise(3000, 0.4).String()

	// short closed hat
	hatRecipe = recipe.New(800).Noise(2000, 1).String()

	// low sine thump, played low
	kickRecipe = recipe.New(2000).Sine(55, 1).Sine(110, 0.25).String()

	// Recipes for the samples, as they were in tracks/track3/samples

	// thin square lead, tuned to C-4 (261.34 Hz)
	leadRecipe = recipe.New(256).Square(recipe.Note(256, "C-4"), 0.45).Duty(0.25).Sine(recipe.Note(256, "C-5"), 0.2).String()

	// open hat, longer noise tail
	openHatRecipe = recipe.New(3000).Noise(2500, 1).String()

	// soft pad, tuned to C-4
	padRecipe = recipe.New(512).Triangle(recipe.Note(512, "C-4"), 0.35).Sine(recipe.Note(512, "C-3"), 0.25).String()

	// saw lead for chorus harmony, tuned to C-4
	sawLeadRecipe = recipe.New(256).Saw(recipe.Note(256, "C-4"), 0.4).Square(recipe.Note(256, "C-3"), 0.25).String()

	// noise snare with a little body
	snareRecipe = recipe.New(3000).Noise(1000, 1).Triangle(185, 0.25).String()

	// tom, a low sine with a short decay
	tomRecipe = recipe.New(1200).Sine(110, 1).Sine(165, 0.3).String()

	// warm pad with a slow attack, tuned to C-4
	warmPadRecipe = recipe.New(512).Triangle(recipe.Note(512, "C-4"), 0.3).Saw(recipe.Note(512, "C-3"), 0.12).Sine(recipe.Note(512, "C-2"), 0.15).String()
	bridgeChords = []chordBar{
		{
			chord: "Dm",
			padOct: 3,
			bass: "D-2",
			bassTwo: [2]string{"D-2", "A-2"},
			lead: []string{"D-4", "F-4", "A-4", "F-4"},
			arp: []string{"D-4", "F-4", "A-4", "D-5", "A-4", "F-4", "D-4", "A-3"},
		},
		{
			chord: "Am",
			padOct: 3,
			bass: "A-2",
			bassTwo: [2]string{"A-2", "E-2"},
			lead: []string{"C-5", "E-5", "A-5", "E-5"},
			arp: []string{"C-5", "E-5", "A-5", "C-6", "A-5", "E-5", "C-5", "A-4"},
		},
		{
			chord: "F",
			padOct: 3,
			bass: "F-2",
			bassTwo: [2]string{"F-2", "C-3"},
			lead: []string{"F-4", "A-4", "C-5", "A-4"},
			arp: []string{"F-4", "A-4", "C-5", "F-5", "C-5", "A-4", "F-4", "C-4"},
		},
		{
			chord: "G",
			padOct: 3,
			bass: "G-2",
			bassTwo: [2]string{"G-2", "D-3"},
			lead: []string{"D-4", "G-4", "B-4", "D-5"},
			arp: []string{"D-4", "G-4", "B-4", "D-5", "B-4", "G-4", "D-4", "G-4"},
		},
	}
	chorusChords = []chordBar{
		{
			chord: "F",
			padOct: 3,
			bass: "F-2",
			bassTwo: [2]string{"F-2", "C-3"},
			lead: []string{"C-5", "A-4", "F-4", "A-4"},
			arp: []string{"F-4", "A-4", "C-5", "F-5", "C-5", "A-4", "F-4", "C-4"},
			saw: []string{"F-3", "A-3", "C-4", "A-3"},
		},
		{
			chord: "C",
			padOct: 4,
			bass: "C-3",
			bassTwo: [2]string{"C-3", "G-2"},
			lead: []string{"E-5", "C-5", "G-4", "C-5"},
			arp: []string{"C-4", "E-4", "G-4", "C-5", "G-4", "E-4", "C-4", "G-3"},
			saw: []string{"G-3", "C-4", "E-4", "C-4"},
		},
		{
			chord: "G",
			padOct: 3,
			bass: "G-2",
			bassTwo: [2]string{"G-2", "D-3"},
			lead: []string{"D-5", "B-4", "G-4", "B-4"},
			arp: []string{"G-3", "B-3", "D-4", "G-4", "D-4", "B-3", "G-3", "D-3"},
			saw: []string{"G-3", "B-3", "D-4", "B-3"},
		},
		{
			chord: "Am",
			padOct: 3,
			bass: "A-2",
			bassTwo: [2]string{"A-2", "E-2"},
			lead: []string{"C-5", "A-4", "E-4", "A-4"},
			arp: []string{"A-3", "C-4", "E-4", "A-4", "E-4", "C-4", "A-3", "E-3"},
			saw: []string{"A-3", "C-4", "E-4", "C-4"},
		},
	}

	// solo lines for the four bars of the verse progression
	soloLead = [][]string{
		{"E-5", "D-5", "C-5", "A-4"},
		{"C-5", "A-4", "F-4", "A-4"},
		{"G-4", "E-4", "G-4", "C-5"},
		{"D-5", "B-4", "G-4", "D-4"},
	}
	soloSaw = [][]string{
		{"E-4", "D-4", "C-4", "A-3"},
		{"C-4", "A-3", "F-3", "A-3"},
		{"G-3", "E-3", "G-3", "C-4"},
		{"D-4", "B-3", "G-3", "D-3"},
	}
	verseChords = []chordBar{
		{
			chord: "Am",
			padOct: 3,
			bass: "A-2",
			bassTwo: [2]string{"A-2", "E-2"},
			lead: []string{"A-4", "C-5", "E-5", "C-5"},
			arp: []string{"A-4", "C-5", "E-5", "A-5", "E-5", "C-5", "A-4", "E-4"},
		},
		{
			chord: "F",
			padOct: 3,
			bass: "F-2",
			bassTwo: [2]string{"F-2", "C-3"},
			lead: []string{"F-4", "A-4", "C-5", "A-4"},
			arp: []string{"F-4", "A-4", "C-5", "F-5", "C-5", "A-4", "F-4", "C-4"},
		},
		{
			chord: "C",
			padOct: 4,
			bass: "C-3",
			bassTwo: [2]string{"C-3", "G-2"},
			lead: []string{"E-4", "G-4", "C-5", "G-4"},
			arp: []string{"E-4", "G-4", "C-5", "E-5", "C-5", "G-4", "E-4", "C-4"},
		},
		{
			chord: "G",
			padOct: 3,
			bass: "G-2",
			bassTwo: [2]string{"G-2", "D-3"},
			lead: []string{"D-4", "G-4", "B-4", "G-4"},
			arp: []string{"D-4", "G-4", "B-4", "D-5", "B-4", "G-4", "D-4", "G-4"},
		},
	}
)

// at returns the row at offset r within bar b
func at(b, r int) int {
	return b*song.RowsPerBar + r
}

// pad returns the pad triad for the bar
func (cb chordBar) pad() []string {
	return song.Triad(cb.chord, cb.padOct)
}

func contains(list []int, v int) bool {
	return slices.Contains(list, v)
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

func (w pattern) bass(b int, cb chordBar, eighths bool, vol int) {
	if eighths {
		for i, note := range cb.bassTwo {
			w.p.Hit(at(b, i*8), chBass, note, w.k.bass, withVol(vol)...)
		}

		return
	}

	w.p.Hit(at(b, 0), chBass, cb.bass, w.k.bass, withVol(vol)...)
}

func (w pattern) bell(b int, notes []string, step, vol int) {
	w.p.Line(chArp, at(b, 0), step, notes, w.k.bell, withVol(vol)...)
}

func (w pattern) crash(b, vol int) {
	w.p.Hit(at(b, 0), chHat, "C-4", w.k.crash, withVol(vol)...)
}

// hats plays closed hats every step rows, skipping the rows in skip, and alternates panning
func (w pattern) hats(b, step int, skip []int, baseVol, accentVol int) {
	for i, r := range rangeStep(0, song.RowsPerBar, step) {
		if contains(skip, r) {
			continue
		}

		vol := baseVol

		if r%4 == 0 {
			vol = accentVol
		}

		pan := 0xD0

		if i%2 == 0 {
			pan = 0x30
		}

		w.p.Hit(at(b, r), chHat, "C-5", w.k.hat, withVol(vol, song.Pan(pan))...)
	}
}

// lead plays notes on the lead channel, one every step rows
func (w pattern) lead(b int, notes []string, step, vol int, extra ...song.Opt) {
	w.p.Line(chLead, at(b, 0), step, notes, w.k.lead, withVol(vol, extra...)...)
}

func (w pattern) openHats(b int) {
	for _, r := range []int{0, 4, 8, 12} {
		pan := 0xD0

		if r%8 == 0 {
			pan = 0x30
		}

		w.p.Hit(at(b, r), chHat, "C-5", w.k.hat, withVol(0x40, song.Pan(pan))...)
	}

	for _, r := range []int{2, 6, 10, 14} {
		pan := 0x30

		if r%8 == 2 {
			pan = 0xD0
		}

		w.p.Hit(at(b, r), chHat, "C-4", w.k.openHat, withVol(0x34, song.Pan(pan))...)
	}
}

func (w pattern) pads(b int, cb chordBar, ins *song.Instrument, vol int) {
	pans := [3]int{0x60, 0x80, 0xA0}

	for i, note := range cb.pad() {
		w.p.Hit(at(b, 0), chPadLo+i, note, ins, withVol(vol, song.Pan(pans[i]))...)
	}
}

func break2(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.pads(b, cb, k.warm, 0x24)
		w.bass(b, cb, false, 0x28)
		w.bell(b, cb.arp, 2, 0x20)
	}

	return w.p
}

func (w pattern) saw(b int, notes []string, vol int) {
	w.p.Line(chArp, at(b, 0), 4, notes, w.k.saw, withVol(vol)...)
}

func (w pattern) snare(b, r, vol int) {
	w.p.Hit(at(b, r), chDrum, "C-4", w.k.snare, withVol(vol)...)
}

func breakdown(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.pads(b, cb, k.warm, 0x24)
		w.bass(b, cb, false, 0)
		w.bell(b, cb.arp, 2, 0x28)
		w.hats(b, 4, nil, 0x24, 0x28)
	}

	w.crash(0, 0x28)

	for i, vol := range []int{0x20, 0x28, 0x30, 0x38, 0x40} {
		r := 15

		if i < 4 {
			r = 8 + i*2
		}

		w.snare(3, r, vol)
	}

	return w.p
}

func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.pads(b, cb, k.warm, 0x24)
		w.bass(b, cb, false, 0x30)
		w.bell(b, cb.arp, 2, 0x20)

		if b >= 1 {
			w.hats(b, 4, nil, 0x28, 0x30)
		}

		if b >= 2 {
			w.lead(b, cb.lead, 4, 0x38)
		}
	}

	w.crash(0, 0x28)

	for i, vol := range []int{0x20, 0x28, 0x30, 0x38} {
		w.snare(3, 12+i, vol)
	}

	return w.p
}

// drums plays the kick and snare for a bar, with optional fills, ghost notes and extra kicks
func (w pattern) drums(b int, fill, ghosts, extraKicks bool) {
	w.p.Hit(at(b, 0), chDrum, "C-2", w.k.kick)
	w.p.Hit(at(b, 8), chDrum, "C-2", w.k.kick)

	if extraKicks {
		w.p.Hit(at(b, 6), chDrum, "C-2", w.k.kick, withVol(0x38)...)
		w.p.Hit(at(b, 14), chDrum, "C-2", w.k.kick, withVol(0x38)...)
	}

	w.snare(b, 4, 0)
	w.snare(b, 12, 0)

	if ghosts {
		w.snare(b, 15, 0x18)
	}

	if fill {
		w.snare(b, 13, 0x28)
		w.snare(b, 14, 0x38)
		w.snare(b, 15, 0x48)
	}
}

func bridgeBars(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range bridgeChords {
		w.pads(b, cb, k.warm, 0x24)
		w.bass(b, cb, false, 0)
		w.lead(b, cb.lead, 4, 0)
		w.bell(b, cb.arp, 2, 0x24)

		crashVol := 0x28

		if b == 0 {
			crashVol = 0x30
		}

		w.crash(b, crashVol)
		w.hats(b, 2, []int{0}, 0x28, 0x34)
		w.drums(b, false, false, false)
	}

	return w.p
}

func chorus(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range chorusChords {
		w.pads(b, cb, k.pad, 0x28)
		w.bass(b, cb, true, 0)
		w.lead(b, cb.lead, 4, 0)
		w.saw(b, cb.saw, 0x38)
		w.openHats(b)

		crashVol := 0x28

		if b == 0 || b == 2 {
			crashVol = 0x40
		}

		w.crash(b, crashVol)
		w.drums(b, false, false, true)
	}

	w.drums(3, true, false, true)

	return w.p
}

func chorusVariation(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range chorusChords {
		w.pads(b, cb, k.pad, 0x28)
		w.bass(b, cb, true, 0)
		w.lead(b, cb.arp, 2, 0)
		w.saw(b, cb.saw, 0x3C)
		w.openHats(b)

		crashVol := 0x28

		if b == 0 || b == 2 {
			crashVol = 0x38
		}

		w.crash(b, crashVol)
		w.drums(b, false, true, true)
	}

	w.drums(3, true, true, true)

	return w.p
}

func outro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords[:2] {
		w.pads(b, cb, k.pad, 0x24)
		w.bass(b, cb, false, 0)
		w.lead(b, cb.lead, 4, 0)
		w.bell(b, cb.arp, 2, 0x28)

		crashVol := 0x28

		if b == 0 {
			crashVol = 0x40
		}

		w.crash(b, crashVol)
		w.hats(b, 2, []int{0}, 0x30, 0x40)
		w.drums(b, false, false, false)
	}

	// a quiet bar of C
	cb := verseChords[2]

	w.pads(2, cb, k.warm, 0x20)
	w.bass(2, cb, false, 0x30)
	w.bell(2, cb.arp, 2, 0x20)
	w.hats(2, 4, nil, 0x24, 0x28)

	for i, vol := range []int{0x20, 0x28, 0x30, 0x38} {
		w.snare(2, 12+i, vol)
	}

	// the final Am hit, with a crash and note offs to let it ring out
	am := verseChords[0]

	w.pads(3, am, k.pad, 0x28)
	w.p.Hit(at(3, 0), chLead, "A-4", k.lead, song.Vol(0x50), song.Vibrato(4, 4))
	w.p.Hit(at(3, 0), chBass, "A-2", k.bass)
	w.p.Hit(at(3, 0), chDrum, "C-2", k.kick)
	w.crash(3, 0x48)
	w.snare(3, 4, 0)

	for _, sn := range []struct {
		r, vol int
	}{
		{8, 0x28},
		{10, 0x30},
		{12, 0x38},
		{14, 0x48},
	} {
		w.snare(3, sn.r, sn.vol)
	}

	w.p.Hit(rows-1, chLead, "###", nil)
	w.p.Hit(rows-2, chBass, "###", nil)

	for _, ch := range []int{chPadLo, chPadMid, chPadHi} {
		w.p.Hit(rows-2, ch, "###", nil)
	}

	return w.p
}

// tomFill plays a four-note tom run starting at row start
func (w pattern) tomFill(b, start int) {
	for i, note := range []string{"C-3", "G-3", "C-4", "G-3"} {
		w.p.Hit(at(b, start+i), chDrum, note, w.k.tom, withVol(0x30+i*4)...)
	}
}

func build(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.pads(b, cb, k.warm, 0x24)
		w.bass(b, cb, b >= 2, 0)

		step := 2

		if b == 3 {
			step = 1
		}

		w.bell(b, cb.arp, step, 0x20+b*2)

		hatSteps := []int{4, 2, 2, 1}

		w.hats(b, hatSteps[b], nil, 0x20+b*4, 0x28+b*4)

		if b >= 1 {
			w.p.Hit(at(b, 0), chDrum, "C-2", k.kick)
			w.p.Hit(at(b, 8), chDrum, "C-2", k.kick)
		}
	}

	w.crash(0, 0x28)
	w.snare(0, 8, 0x20)

	for _, r := range []int{4, 12} {
		w.snare(1, r, 0x28)
	}

	for i, vol := range []int{0x28, 0x2C, 0x30, 0x38} {
		w.snare(2, i*4, vol)
	}

	for i, r := range rangeStep(0, 12, 2) {
		w.snare(3, r, 0x28+i*2)
	}

	w.tomFill(3, 12)

	return w.p
}

func build2(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.pads(b, cb, k.warm, 0x28)
		w.bass(b, cb, true, 0)
		w.bell(b, cb.arp, 1, 0x20+b*2)

		hatSteps := []int{2, 1, 1, 1}

		w.hats(b, hatSteps[b], nil, 0x24+b*4, 0x2C+b*4)

		if b >= 1 {
			w.p.Hit(at(b, 0), chDrum, "C-2", k.kick)
			w.p.Hit(at(b, 8), chDrum, "C-2", k.kick)
		}
	}

	w.crash(0, 0x30)

	for _, r := range []int{4, 12} {
		w.snare(0, r, 0x20)
	}

	for _, r := range []int{4, 12} {
		w.snare(1, r, 0x28)
	}

	for i, r := range rangeStep(0, 16, 2) {
		w.snare(2, r, 0x28+i*2)
	}

	for i, r := range rangeStep(0, 12, 1) {
		w.snare(3, r, 0x30+i*2)
	}

	w.tomFill(3, 12)

	return w.p
}

func finalChorus(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range chorusChords {
		w.pads(b, cb, k.pad, 0x2C)
		w.bass(b, cb, true, 0)
		w.lead(b, cb.lead, 4, 0x50)
		w.saw(b, cb.saw, 0x40)
		w.openHats(b)
		w.crash(b, 0x40)
		w.drums(b, false, true, true)
	}

	w.hats(3, 1, nil, 0x30, 0x40)
	w.crash(3, 0x40)
	w.tomFill(3, 12)

	return w.p
}

func solo(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.pads(b, cb, k.pad, 0x28)
		w.bass(b, cb, true, 0)
		w.openHats(b)

		crashVol := 0x28

		if b == 0 || b == 2 {
			crashVol = 0x40
		}

		w.crash(b, crashVol)
		w.drums(b, false, true, true)

		for i, note := range soloLead[b] {
			w.p.Hit(at(b, i*4), chLead, note, k.lead, song.Vol(0x50), song.Vibrato(4, 4))
		}

		w.saw(b, soloSaw[b], 0x34)
	}

	return w.p
}

func verseEighths(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.pads(b, cb, k.pad, 0x24)
		w.bass(b, cb, true, 0)
		w.lead(b, cb.arp, 2, 0)
		w.bell(b, cb.arp, 2, 0x24)

		crashVol := 0x28

		if b == 0 {
			crashVol = 0x38
		}

		w.crash(b, crashVol)
		w.hats(b, 2, []int{0}, 0x30, 0x40)
		w.drums(b, b == 3, false, true)
	}

	return w.p
}

func verseFull(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.pads(b, cb, k.pad, 0x24)
		w.bass(b, cb, false, 0)
		w.lead(b, cb.lead, 4, 0)
		w.bell(b, cb.arp, 2, 0x28)

		crashVol := 0x28

		if b == 0 {
			crashVol = 0x40
		}

		w.crash(b, crashVol)
		w.hats(b, 2, []int{0}, 0x30, 0x40)
		w.drums(b, false, false, false)
	}

	return w.p
}

func main() {
	s := song.New("night drive", 8)
	k := &kit{
		lead: s.Instrument(song.Instrument{
			Name: "lead",
			Recipe: leadRecipe,
			Loop: true,
			Volume: 40,
			Pan: 150,
		}),
		bass: s.Instrument(song.Instrument{
			Name: "bass",
			Recipe: bassRecipe,
			Loop: true,
			Volume: 56,
			Pan: 110,
		}),
		kick: s.Instrument(song.Instrument{
			Name: "kick",
			Recipe: kickRecipe,
			Volume: 64,
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
			Volume: 48,
			Pan: 136,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 18,
					Y: 0,
				},
			},
		}),
		hat: s.Instrument(song.Instrument{
			Name: "hat",
			Recipe: hatRecipe,
			Volume: 40,
			Pan: 120,
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
		pad: s.Instrument(song.Instrument{
			Name: "pad",
			Recipe: padRecipe,
			Loop: true,
			Volume: 36,
			Pan: 128,
			Fadeout: 512,
		}),
		crash: s.Instrument(song.Instrument{
			Name: "crash",
			Recipe: crashRecipe,
			Volume: 40,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 40,
					Y: 0,
				},
			},
		}),
		bell: s.Instrument(song.Instrument{
			Name: "bell",
			Recipe: bellRecipe,
			Volume: 40,
			Pan: 160,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 12,
					Y: 0,
				},
			},
		}),
		saw: s.Instrument(song.Instrument{
			Name: "saw lead",
			Recipe: sawLeadRecipe,
			Loop: true,
			Volume: 36,
			Pan: 100,
		}),
		warm: s.Instrument(song.Instrument{
			Name: "warm pad",
			Recipe: warmPadRecipe,
			Loop: true,
			Volume: 36,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 8,
					Y: 64,
				},
			},
			Fadeout: 512,
		}),
		tom: s.Instrument(song.Instrument{
			Name: "tom",
			Recipe: tomRecipe,
			Volume: 48,
			Pan: 150,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 8,
					Y: 0,
				},
			},
		}),
		openHat: s.Instrument(song.Instrument{
			Name: "open hat",
			Recipe: openHatRecipe,
			Volume: 40,
			Pan: 100,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 12,
					Y: 0,
				},
			},
		}),
	}
	patterns := []*song.Pattern{
		intro(s, k),
		verseFull(s, k),
		verseEighths(s, k),
		chorus(s, k),
		breakdown(s, k),
		build(s, k),
		chorusVariation(s, k),
		bridgeBars(s, k),
		solo(s, k),
		break2(s, k),
		build2(s, k),
		finalChorus(s, k),
		outro(s, k),
	}

	for _, i := range []int{0, 1, 2, 3, 3, 1, 2, 4, 5, 6, 7, 8, 9, 10, 11, 11, 3, 6, 11, 12} {
		s.Play(patterns[i])
	}

	// write next to this file, wherever the program is run from
	_, src, _, _ := runtime.Caller(0)

	if err := s.Write(filepath.Join(filepath.Dir(src), "song.xm")); err != nil {
		log.Fatal(err)
	}
}
