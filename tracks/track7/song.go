// Command track7 writes "afterglow", an uplifting trance track, as an XM module, song.xm, in this folder.
//
// Build it from anywhere inside the repository with go run ./tracks/track7, or with tracker build.
package main

// afterglow – HungerG

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/song"
)

type (
	// chordBar is one bar of F#m9 - Dmaj9 - Aadd9 - Eadd9
	chordBar struct {
		name, root, fifth, octave string
		pad [4]string
		arp [16]string
		pluck [8]string
	}

	// hit is one melody note at a row within the bar
	hit struct {
		row int
		note string
	}

	// kit holds the instruments, numbered as they are added to the song
	kit struct {
		lead, arp, pluck, bass, pad, kick, snare, hat, openHat, crash, riser *song.Instrument
	}

	// pattern wraps a song pattern with the instruments to play it with
	pattern struct {
		p *song.Pattern
		k *kit
	}

	// section describes the layers and intensity of a four-bar pattern
	section struct {
		padVol, arpStep, arpVol, pluckStep, pluckVol, bassStep, bassVol, drumStyle, leadVol int
		phrase [][]hit
		echo, fill bool
	}
)

// channels, in the order they appear in the tracker
const (
	chLead = iota
	chEcho
	chArp
	chPluck
	chBass
	chPadRoot
	chPadThird
	chPadFifth
	chPadNinth
	chKick
	chSnare
	chHat
	chOpenHat
	chFx
)

const (
	rows = 64
	// Six ticks per row at 138 BPM gives a sixteenth-note row and a 1.74-second bar
	speed = 6
	tempo = 138
)

var (
	// A wide saw lead, tuned to C-4
	leadRecipe = recipe.New(256).Saw(recipe.Note(256, "C-4"), 0.23).Saw(262.4, 0.23).Phase(0.5).Square(recipe.Note(256, "C-3"), 0.07).Duty(0.25).Sine(recipe.Note(256, "C-5"), 0.04).String()

	// A bright, short pulse for the sixteenth-note arpeggio
	arpRecipe = recipe.New(512).Square(recipe.Note(512, "C-4"), 0.19).Duty(0.25).Sine(recipe.Note(512, "C-5"), 0.06).Sine(1046.875, 0.025).String()

	// A soft mallet for the offbeat chord answer
	pluckRecipe = recipe.New(2048).Triangle(recipe.Note(2048, "C-4"), 0.20).Saw(recipe.Note(2048, "C-4"), 0.09).Sine(recipe.Note(2048, "C-5"), 0.05).String()

	// A detuned saw bass with a sine fundamental
	bassRecipe = recipe.New(256).Saw(recipe.Note(256, "C-4"), 0.20).Saw(262.2, 0.13).Phase(0.5).Sine(recipe.Note(256, "C-3"), 0.20).String()

	// A wide, restrained pad; its four notes are spread across four channels
	padRecipe = recipe.New(512).Saw(recipe.Note(512, "C-4"), 0.11).Saw(262.4, 0.11).Phase(0.5).Triangle(recipe.Note(512, "C-3"), 0.09).Sine(recipe.Note(512, "C-2"), 0.05).String()

	// A deep kick with a short click, played at C-2
	kickRecipe = recipe.New(2000).Sine(220, 0.72).Sine(440, 0.18).Noise(4000, 0.035).String()

	// A broad noise clap with a little tonal body
	snareRecipe = recipe.New(2400).Noise(1200, 0.50).Noise(3600, 0.18).Triangle(185, 0.12).String()

	// A short closed hat
	hatRecipe = recipe.New(500).Noise(1, 0.36).Square(4100, 0.035).String()

	// A longer open hat for the offbeats
	openHatRecipe = recipe.New(1800).Noise(1, 0.34).Square(3800, 0.03).String()

	// A long bright crash
	crashRecipe = recipe.New(8000).Noise(1000, 0.40).Noise(3000, 0.22).String()

	// A filtered-noise swell, long enough to run across a whole build without looping
	riserRecipe = recipe.New(60000).Noise(1500, 0.32).Noise(3500, 0.10).String()
	progression = []chordBar{
		{
			name: "F#m9",
			root: "F#2",
			fifth: "C#3",
			octave: "F#3",
			pad: [4]string{"F#3", "A-3", "C#4", "G#4"},
			arp: [16]string{
				"F#3",
				"A-3",
				"C#4",
				"G#4",
				"C#4",
				"A-3",
				"F#3",
				"C#4",
				"F#3",
				"A-3",
				"C#4",
				"G#4",
				"A-4",
				"G#4",
				"C#4",
				"A-3",
			},
			pluck: [8]string{"F#4", "C#5", "A-4", "G#4", "F#4", "C#5", "A-4", "C#5"},
		},
		{
			name: "Dmaj9",
			root: "D-2",
			fifth: "A-2",
			octave: "D-3",
			pad: [4]string{"D-3", "F#3", "A-3", "E-4"},
			arp: [16]string{
				"D-3",
				"F#3",
				"A-3",
				"E-4",
				"A-3",
				"F#3",
				"D-3",
				"A-3",
				"D-3",
				"F#3",
				"A-3",
				"E-4",
				"F#4",
				"E-4",
				"A-3",
				"F#3",
			},
			pluck: [8]string{"D-4", "A-4", "F#4", "E-4", "D-4", "A-4", "F#4", "A-4"},
		},
		{
			name: "Aadd9",
			root: "A-2",
			fifth: "E-3",
			octave: "A-3",
			pad: [4]string{"A-3", "C#4", "E-4", "B-4"},
			arp: [16]string{
				"A-3",
				"C#4",
				"E-4",
				"B-4",
				"E-4",
				"C#4",
				"A-3",
				"E-4",
				"A-3",
				"C#4",
				"E-4",
				"B-4",
				"C#5",
				"B-4",
				"E-4",
				"C#4",
			},
			pluck: [8]string{"A-4", "E-5", "C#5", "B-4", "A-4", "E-5", "C#5", "E-5"},
		},
		{
			name: "Eadd9",
			root: "E-2",
			fifth: "B-2",
			octave: "E-3",
			pad: [4]string{"E-3", "G#3", "B-3", "F#4"},
			arp: [16]string{
				"E-3",
				"G#3",
				"B-3",
				"F#4",
				"B-3",
				"G#3",
				"E-3",
				"B-3",
				"E-3",
				"G#3",
				"B-3",
				"F#4",
				"G#4",
				"F#4",
				"B-3",
				"G#3",
			},
			pluck: [8]string{"E-4", "B-4", "G#4", "F#4", "E-4", "B-4", "G#4", "B-4"},
		},
	}

	// The bridge turns the loop around: D - A - E - F#m
	bridgeProgression = []chordBar{progression[1], progression[2], progression[3], progression[0]}

	// hookA is the main melody; hookB is its higher, more open answer
	hookA = [][]hit{
		{
			{0, "F#5"},
			{2, "A-5"},
			{3, "G#5"},
			{4, "F#5"},
			{6, "E-5"},
			{8, "C#5"},
			{10, "E-5"},
			{12, "F#5"},
			{14, "G#5"},
		},
		{{0, "A-5"}, {2, "F#5"}, {4, "E-5"}, {6, "F#5"}, {8, "A-5"}, {10, "B-5"}, {12, "A-5"}, {14, "F#5"}},
		{{0, "E-5"}, {2, "G#5"}, {4, "A-5"}, {6, "B-5"}, {8, "C#6"}, {10, "B-5"}, {12, "A-5"}, {14, "E-5"}},
		{{0, "G#5"}, {2, "B-5"}, {4, "A-5"}, {6, "G#5"}, {8, "F#5"}, {10, "E-5"}, {12, "B-4"}, {14, "G#4"}},
	}
	hookB = [][]hit{
		{{0, "C#6"}, {4, "A-5"}, {6, "G#5"}, {8, "F#5"}, {12, "A-5"}},
		{{0, "D-6"}, {4, "A-5"}, {8, "F#5"}, {10, "E-5"}, {12, "D-5"}},
		{{0, "E-5"}, {4, "A-5"}, {8, "C#6"}, {12, "B-5"}, {14, "A-5"}},
		{{0, "B-5"}, {4, "G#5"}, {8, "F#5"}, {12, "E-5"}, {14, "B-4"}},
	}
	verseLead = [][]hit{
		{{0, "C#5"}, {4, "A-4"}, {8, "C#5"}, {12, "E-5"}},
		{{0, "D-5"}, {6, "F#5"}, {10, "E-5"}, {12, "D-5"}},
		{{0, "E-5"}, {4, "C#5"}, {8, "B-4"}, {12, "E-5"}},
		{{0, "G#4"}, {4, "B-4"}, {8, "E-5"}, {12, "D-5"}},
	}
	introLead = [][]hit{
		{},
		{{8, "F#5"}, {12, "G#5"}},
		{{0, "A-5"}, {8, "E-5"}},
		{{0, "F#5"}, {8, "E-5"}, {12, "C#5"}},
	}
	bridgeLead = [][]hit{
		{{0, "F#5"}, {4, "A-5"}, {8, "B-5"}, {12, "A-5"}},
		{{0, "E-5"}, {4, "G#5"}, {8, "A-5"}, {12, "B-5"}},
		{{0, "B-5"}, {6, "G#5"}, {8, "F#5"}, {12, "E-5"}},
		{{0, "C#5"}, {4, "E-5"}, {8, "F#5"}, {14, "C#5"}},
	}
)

// at returns the row at offset r within bar b
func at(b, r int) int {
	return b*song.RowsPerBar + r
}

// withVol returns a volume-column option, omitted when zero, followed by extra
func withVol(vol int, extra ...song.Opt) []song.Opt {
	var opts []song.Opt

	if vol != 0 {
		opts = append(opts, song.Vol(vol))
	}

	return append(opts, extra...)
}

// arp plays the bar's arpeggio at the requested step
func (w pattern) arp(b int, cb chordBar, step, vol int) {
	if vol == 0 {
		return
	}

	for i, r := 0, 0; i < len(cb.arp); i, r = i+1, r+step {
		if r >= song.RowsPerBar {
			break
		}

		pan := 0x50

		if i%2 == 1 {
			pan = 0xB0
		}

		w.p.Hit(at(b, r), chArp, cb.arp[i], w.k.arp, withVol(vol, song.Pan(pan))...)
	}
}

// bass plays rolling sixteenths, offbeat eighths or quarter-note roots
func (w pattern) bass(b int, cb chordBar, step, vol int) {
	if vol == 0 {
		return
	}

	for r := 0; r < song.RowsPerBar; r += step {
		note := cb.root

		switch r % 16 {
		case 3, 15:
			note = cb.octave
		case 6, 14:
			note = cb.fifth
		}

		cellVol := vol

		// Make room for each kick without needing a sidechain effect
		if r%4 == 0 {
			cellVol -= 0x0C
		}

		w.p.Hit(at(b, r), chBass, note, w.k.bass, withVol(cellVol)...)
	}
}

// crash puts a crash at the top of a bar
func (w pattern) crash(b, vol int) {
	if vol != 0 {
		w.p.Hit(at(b, 0), chFx, "C-4", w.k.crash, withVol(vol)...)
	}
}

// drums plays the four-on-the-floor groove; style 1 is soft, 2 is full and 3 builds
func (w pattern) drums(b, style int) {
	if style == 0 {
		return
	}

	for _, r := range []int{0, 4, 8, 12} {
		vol := 0x38

		if style == 2 {
			vol = 0x44
		}

		if r == 0 {
			vol += 0x08
		}

		w.p.Hit(at(b, r), chKick, "C-2", w.k.kick, withVol(vol)...)
	}

	if style >= 2 {
		w.p.Hit(at(b, 4), chSnare, "C-4", w.k.snare, withVol(0x48)...)
		w.p.Hit(at(b, 12), chSnare, "C-4", w.k.snare, withVol(0x48)...)
	}

	if openRows := []int{}; style == 2 {
		openRows = []int{2, 6, 10, 14}

		for _, r := range openRows {
			w.p.Hit(at(b, r), chOpenHat, "C-5", w.k.openHat, withVol(0x28, song.Pan(0xD0))...)
		}
	}

	step := 4

	if style >= 2 {
		step = 2
	}

	if style == 2 && b%2 == 1 {
		step = 1
	}

	for i, r := 0, 0; r < song.RowsPerBar; i, r = i+1, r+step {
		if r == 2 || r == 6 || r == 10 || r == 14 {
			if style == 2 {
				continue
			}
		}

		vol := 0x20

		if r%4 == 0 {
			vol = 0x2C
		}

		pan := 0x50

		if i%2 == 1 {
			pan = 0xB0
		}

		w.p.Hit(at(b, r), chHat, "C-5", w.k.hat, withVol(vol, song.Pan(pan))...)
	}

	if style == 3 {
		for i, r := range []int{8, 10, 12, 13, 14, 15} {
			w.p.Hit(at(b, r), chSnare, "C-4", w.k.snare, withVol(0x24+i*5)...)
		}
	}
}

// fill plays a snare run at the end of the last bar, keeping the backbeat on row 12
func (w pattern) fill(b int) {
	for i, r := range []int{12, 13, 14, 15} {
		vol := 0x48

		if i > 0 {
			vol = 0x30 + i*8
		}

		w.p.Hit(at(b, r), chSnare, "C-4", w.k.snare, withVol(vol)...)
	}
}

// melody plays a phrase with a quiet dotted-eighth echo on the opposite side
func (w pattern) melody(b int, phrase []hit, vol int, echo bool) {
	if vol == 0 {
		return
	}

	for i, h := range phrase {
		w.p.Hit(at(b, h.row), chLead, h.note, w.k.lead, withVol(vol, song.Vibrato(5, 2))...)

		off := song.RowsPerBar - 1

		if i+1 < len(phrase) {
			off = phrase[i+1].row - 1
		}

		if off > h.row {
			w.p.Hit(at(b, off), chLead, "###", nil)
		}

		if !echo {
			continue
		}

		if echoRow := h.row + 3; echoRow < song.RowsPerBar {
			w.p.Hit(at(b, echoRow), chEcho, h.note, w.k.lead, withVol(vol-0x14, song.Pan(0xD0))...)
		}

		if off+3 > h.row+3 && off+3 < song.RowsPerBar {
			w.p.Hit(at(b, off+3), chEcho, "###", nil)
		}
	}

	if echo {
		w.p.Hit(at(b, song.RowsPerBar-1), chEcho, "###", nil)
	}
}

// pads holds the four-note voicing across the pad channels
func (w pattern) pads(b int, cb chordBar, vol int) {
	if vol == 0 {
		return
	}

	pans := [4]int{0x30, 0x58, 0xA8, 0xD8}

	for i, note := range cb.pad {
		w.p.Hit(at(b, 0), chPadRoot+i, note, w.k.pad, withVol(vol, song.Pan(pans[i]))...)
	}
}

// pluck plays the chord answer in eighths or quarters
func (w pattern) pluck(b int, cb chordBar, step, vol int) {
	if vol == 0 {
		return
	}

	for i, r := 0, 0; i < len(cb.pluck); i, r = i+1, r+step {
		if r >= song.RowsPerBar {
			break
		}

		pan := 0x70

		if i%2 == 1 {
			pan = 0xC0
		}

		w.p.Hit(at(b, r), chPluck, cb.pluck[i], w.k.pluck, withVol(vol, song.Pan(pan))...)
	}
}

func breakdown(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression {
		w.pads(b, cb, 0x2C)
		w.arp(b, cb, 2, 0x24)
		w.pluck(b, cb, 4, 0x20)
		w.bass(b, cb, 4, 0x2C)
		w.melody(b, verseLead[b], 0x3C, b >= 2)

		if b >= 2 {
			w.drums(b, 1)
		}
	}

	w.crash(0, 0x28)

	for i, vol := range []int{0x20, 0x28, 0x30, 0x38} {
		w.p.Hit(at(3, 12+i), chSnare, "C-4", k.snare, withVol(vol)...)
	}

	return w.p
}

func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression {
		w.pads(b, cb, 0x20)
		w.pluck(b, cb, 4, 0x1C)
		w.melody(b, introLead[b], 0x34, false)

		// Leave the first bar to the pad and mallet, then add the arpeggio
		if b > 0 {
			w.arp(b, cb, 2, 0x20)
		}
	}

	w.crash(0, 0x20)

	// The bass and beat arrive one layer at a time
	w.bass(2, progression[2], 4, 0x24)
	w.bass(3, progression[3], 4, 0x28)
	w.drums(2, 1)
	w.drums(3, 1)

	return w.p
}

func outro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression {
		w.pads(b, cb, 0x28)
		w.arp(b, cb, 2, 0x24)
		w.pluck(b, cb, 4, 0x24)
		w.melody(b, verseLead[b], 0x3C, b == 0)

		if b < 2 {
			w.bass(b, cb, 4, 0x30)
			w.drums(b, 1)
		}
	}

	// End on the home chord and let the pads ring out
	cb := progression[0]

	w.pads(3, cb, 0x30)
	w.bass(3, cb, 4, 0x34)
	w.arp(3, cb, 2, 0x20)
	w.pluck(3, cb, 4, 0x24)
	w.melody(3, []hit{{0, "F#5"}, {8, "C#6"}, {12, "F#5"}}, 0x40, false)
	w.crash(3, 0x38)

	for _, ch := range []int{chLead, chEcho, chArp, chPluck, chBass, chPadRoot, chPadThird, chPadFifth, chPadNinth} {
		w.p.Hit(rows-2, ch, "###", nil)
	}

	return w.p
}

// arrange lays down the shared layers of a section over its progression
func (w pattern) arrange(sec section, chords []chordBar) {
	for b, cb := range chords {
		w.pads(b, cb, sec.padVol)
		w.arp(b, cb, sec.arpStep, sec.arpVol)
		w.pluck(b, cb, sec.pluckStep, sec.pluckVol)
		w.bass(b, cb, sec.bassStep, sec.bassVol)
		w.drums(b, sec.drumStyle)

		if sec.phrase != nil {
			w.melody(b, sec.phrase[b], sec.leadVol, sec.echo)
		}
	}

	if sec.fill {
		w.fill(3)
	}
}

func bridge(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x2C,
		arpStep: 1,
		arpVol: 0x30,
		pluckStep: 4,
		pluckVol: 0x2C,
		bassStep: 2,
		bassVol: 0x34,
		drumStyle: 2,
		leadVol: 0x44,
		phrase: bridgeLead,
		echo: true,
		fill: true,
	}, bridgeProgression)
	w.crash(0, 0x38)
	w.crash(2, 0x2C)

	return w.p
}

func drop(s *song.Song, k *kit, answer, final bool) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	phrase := hookA

	if answer {
		phrase = hookB
	}

	sec := section{
		padVol: 0x2C,
		arpStep: 1,
		arpVol: 0x30,
		pluckStep: 2,
		pluckVol: 0x30,
		bassStep: 1,
		bassVol: 0x3C,
		drumStyle: 2,
		leadVol: 0x48,
		phrase: phrase,
		echo: true,
		fill: true,
	}

	if answer {
		sec.pluckStep = 4
		sec.pluckVol = 0x2C
	}

	if final {
		sec.padVol = 0x30
		sec.arpVol = 0x34
		sec.bassVol = 0x40
		sec.leadVol = 0x4C
	}

	w.arrange(sec, progression)
	w.crash(0, 0x38)

	// Syncopated kicks give each drop a different answer before the fill
	for b, r := range []int{14, 10, 14} {
		w.p.Hit(at(b+1, r), chKick, "C-2", k.kick, withVol(0x30)...)
	}

	if !answer {
		w.crash(2, 0x28)
	}

	return w.p
}

func groove(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x24,
		arpStep: 2,
		arpVol: 0x28,
		pluckStep: 2,
		pluckVol: 0x28,
		bassStep: 2,
		bassVol: 0x34,
		drumStyle: 2,
		leadVol: 0x3C,
		phrase: verseLead,
	}, progression)
	w.crash(0, 0x2C)

	return w.p
}

func introTwo(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x24,
		arpStep: 1,
		arpVol: 0x24,
		pluckStep: 2,
		pluckVol: 0x24,
		bassStep: 2,
		bassVol: 0x30,
		drumStyle: 1,
		leadVol: 0x38,
		phrase: verseLead,
	}, progression)
	w.crash(0, 0x2C)

	return w.p
}

// riser swells across the pattern, then stops before the drop
func (w pattern) riser() {
	w.p.Hit(0, chFx, "C-4", w.k.riser, withVol(0x18)...)
	w.p.Hit(rows-1, chFx, "###", nil)
}

func build(s *song.Song, k *kit, final bool) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression {
		w.pads(b, cb, 0x28)
		w.bass(b, cb, []int{4, 4, 2, 1}[b], 0x30+b*2)

		arpStep := 2

		if b >= 2 {
			arpStep = 1
		}

		w.arp(b, cb, arpStep, 0x24+b*2)
		w.pluck(b, cb, 2, 0x20+b*2)

		if b == 3 {
			w.drums(b, 3)
		} else {
			w.drums(b, 2)
		}
	}

	w.riser()
	w.crash(0, 0x28)

	if final {
		w.melody(2, hookA[2][:5], 0x3C, false)
		w.melody(3, hookA[3][:4], 0x44, false)
	} else {
		w.melody(3, hookA[0][:4], 0x38, false)
	}

	return w.p
}

func main() {
	s := song.New("afterglow", 14)

	s.Timing(speed, tempo)

	k := &kit{
		lead: s.Instrument(song.Instrument{
			Name: "saw lead",
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
					X: 1,
					Y: 64,
				},
			},
			Fadeout: 512,
		}),
		arp: s.Instrument(song.Instrument{
			Name: "arp",
			Recipe: arpRecipe,
			Volume: 34,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 5,
					Y: 0,
				},
			},
		}),
		pluck: s.Instrument(song.Instrument{
			Name: "pluck",
			Recipe: pluckRecipe,
			Volume: 32,
			Pan: 176,
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
			Fadeout: 512,
		}),
		bass: s.Instrument(song.Instrument{
			Name: "rolling bass",
			Recipe: bassRecipe,
			Loop: true,
			Volume: 44,
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
		pad: s.Instrument(song.Instrument{
			Name: "wide pad",
			Recipe: padRecipe,
			Loop: true,
			Volume: 26,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 20,
					Y: 64,
				},
			},
			Fadeout: 1024,
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
			Name: "clap",
			Recipe: snareRecipe,
			Volume: 42,
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
			Name: "closed hat",
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
		openHat: s.Instrument(song.Instrument{
			Name: "open hat",
			Recipe: openHatRecipe,
			Volume: 26,
			Pan: 208,
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
		crash: s.Instrument(song.Instrument{
			Name: "crash",
			Recipe: crashRecipe,
			Volume: 26,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 64,
					Y: 0,
				},
			},
		}),
		riser: s.Instrument(song.Instrument{
			Name: "riser",
			Recipe: riserRecipe,
			Volume: 22,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 384,
					Y: 64,
				},
			},
			Fadeout: 256,
		}),
	}
	introP := intro(s, k)
	introTwoP := introTwo(s, k)
	grooveP := groove(s, k)
	buildP := build(s, k, false)
	dropP := drop(s, k, false, false)
	answerP := drop(s, k, true, false)
	breakP := breakdown(s, k)
	bridgeP := bridge(s, k)
	finalBuildP := build(s, k, true)
	finalDropP := drop(s, k, false, true)
	outroP := outro(s, k)

	s.Play(introP, introTwoP, grooveP, buildP, dropP, answerP, breakP, buildP, dropP, answerP, bridgeP, finalBuildP, finalDropP, finalDropP, answerP, finalDropP, outroP)

	// write next to this file, wherever the program is run from
	_, src, _, _ := runtime.Caller(0)

	if err := s.Write(filepath.Join(filepath.Dir(src), "song.xm")); err != nil {
		log.Fatal(err)
	}
}
