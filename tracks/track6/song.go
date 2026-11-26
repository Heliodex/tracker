// Command track6 writes "blue hour", a drum-and-bass track, as an XM module, song.xm, in this folder.
//
// Build it from anywhere inside the repository with go run ./tracks/track6, or with tracker build.
package main

// blue hour – Hy4 Flash

import (
	"log"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/Heliodex/tracker/song"
	"github.com/Heliodex/tracker/types"
)

type (
	// section builds a four-bar pattern: the parts list says which arrangement to use
	section struct {
		padVol, arpStep, arpVol, reeseVol, subVol, drumStyle, leadVol int
		phrase [][]hit
		harmonise, roll, crashes, fill bool
	}

	// chordBar is one bar of Dm - A# - F - C: the arp runs in sixteenths, and bass gives the eighth-note reese pattern for the bar
	chordBar struct {
		chord, root string
		padOct int
		bass [8]string
		arp [16]string
	}

	// hit is a melody note at a row within the bar
	hit struct {
		row int
		note string
	}

	// kit holds the instruments, numbered as they are added to the song
	kit struct {
		lead, harm, arp, reese, sub, pad, kick, snare, hat, openHat, crash, tom, riser *song.Instrument
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
	chHarm
	chArp
	chReese
	chSub
	chPadLo
	chPadMid
	chPadHi
	chKick
	chSnare
	chHat
	chFx
)

const (
	rows = 64
	// speed 6 at 172 BPM gives a row of about 0.087 s and a bar of 1.4 s
	speed = 6
	tempo = 172
	arpRecipe = `# bright sixteenth-note pluck, tuned to C-4
Frames: 256
Bits: 16
Layer: square 261.34375 0.3 duty=0.25
Layer: sine 1046.875 0.06
`
	crashRecipe = `# long bright crash
Frames: 8000
Bits: 16
Layer: noise 1000 1
Layer: noise 3000 0.5
`
	harmRecipe = `# soft harmony pad, tuned to C-4
Frames: 256
Bits: 16
Layer: triangle 261.34375 0.22
Layer: sine 130.671875 0.1
`
	hatRecipe = `# short tight hat
Frames: 400
Bits: 16
Layer: noise 3500 1
`
	kickRecipe = `# punchy low kick, tuned to C-4 so C-2 lands around 55 Hz
Frames: 1600
Bits: 16
Layer: sine 220 1
Layer: sine 440 0.25
`
	leadRecipe = `# singing saw lead, tuned to C-4
Frames: 256
Bits: 16
Layer: saw 261.34375 0.35
Layer: square 261.34375 0.2 duty=0.25
Layer: sine 522.6875 0.08
`
	openHatRecipe = `# open hat for the offbeats
Frames: 1600
Bits: 16
Layer: noise 3000 1
`
	padRecipe = `# warm sustained pad, tuned to C-4
Frames: 512
Bits: 16
Layer: triangle 261.34375 0.3
Layer: saw 130.671875 0.1
`
	reeseRecipe = `# detuned reese bass with a sine sub, tuned to C-4
Frames: 512
Bits: 16
Layer: saw 261.34375 0.35
Layer: saw 262.9 0.35 phase=0.5
Layer: sine 130.671875 0.25
`
	riserRecipe = `# filtered-noise swell for the build
Frames: 6000
Bits: 16
Layer: noise 1500 1
Layer: sine 392 0.12
`
	snareRecipe = `# snappy snare with a little body
Frames: 2000
Bits: 16
Layer: noise 1000 1
Layer: triangle 185 0.35
`
	subRecipe = `# round sub bass, tuned to C-4
Frames: 256
Bits: 16
Layer: sine 261.34375 0.45
Layer: triangle 130.671875 0.1
`
	tomRecipe = `# low tom for fills
Frames: 1200
Bits: 16
Layer: sine 110 1
Layer: sine 165 0.3
`
)

var (
	progression = []chordBar{
		{
			chord: "Dm",
			root: "D-2",
			padOct: 3,
			bass: [8]string{"D-2", "D-2", "D-3", "D-2", "A-2", "D-2", "D-2", "A-2"},
			arp: [16]string{
				"D-4",
				"F-4",
				"A-4",
				"D-5",
				"A-4",
				"F-4",
				"D-4",
				"F-4",
				"A-4",
				"D-5",
				"F-5",
				"D-5",
				"A-4",
				"F-4",
				"D-4",
				"A-3",
			},
		},
		{
			chord: "A#",
			root: "A#1",
			padOct: 2,
			bass: [8]string{"A#1", "A#1", "A#2", "A#1", "F-2", "A#1", "A#1", "F-2"},
			arp: [16]string{
				"A#3",
				"D-4",
				"F-4",
				"A#4",
				"F-4",
				"D-4",
				"A#3",
				"D-4",
				"F-4",
				"A#4",
				"D-5",
				"A#4",
				"F-4",
				"D-4",
				"A#3",
				"F-3",
			},
		},
		{
			chord: "F",
			root: "F-2",
			padOct: 3,
			bass: [8]string{"F-2", "F-2", "F-3", "F-2", "C-3", "F-2", "F-2", "C-3"},
			arp: [16]string{
				"F-3",
				"A-3",
				"C-4",
				"F-4",
				"C-4",
				"A-3",
				"F-3",
				"A-3",
				"C-4",
				"F-4",
				"A-4",
				"F-4",
				"C-4",
				"A-3",
				"F-3",
				"C-3",
			},
		},
		{
			chord: "C",
			root: "C-2",
			padOct: 3,
			bass: [8]string{"C-2", "C-2", "C-3", "C-2", "G-2", "C-2", "C-2", "G-2"},
			arp: [16]string{
				"C-4",
				"E-4",
				"G-4",
				"C-5",
				"G-4",
				"E-4",
				"C-4",
				"E-4",
				"G-4",
				"C-5",
				"E-5",
				"C-5",
				"G-4",
				"E-4",
				"C-4",
				"G-3",
			},
		},
	}

	// hookA is the main hook, hookB its higher answer, verse and bridge are quieter
	hookA = [][]hit{
		{{0, "A-4"}, {2, "A-4"}, {4, "C-5"}, {6, "D-5"}, {8, "F-5"}, {10, "E-5"}, {12, "D-5"}, {14, "A-4"}},
		{{0, "D-5"}, {2, "D-5"}, {4, "F-5"}, {6, "D-5"}, {8, "A#4"}, {10, "C-5"}, {12, "D-5"}, {14, "F-5"}},
		{{0, "A-4"}, {2, "C-5"}, {4, "F-5"}, {6, "E-5"}, {8, "C-5"}, {10, "A-4"}, {12, "G-4"}, {14, "A-4"}},
		{{0, "G-4"}, {4, "C-5"}, {8, "E-5"}, {12, "D-5"}, {14, "###"}},
	}
	hookB = [][]hit{
		{{0, "D-5"}, {3, "F-5"}, {6, "A-5"}, {9, "F-5"}, {12, "D-5"}},
		{{0, "F-5"}, {3, "A#5"}, {6, "D-5"}, {9, "A#4"}, {12, "F-5"}},
		{{0, "C-5"}, {3, "F-5"}, {6, "A-5"}, {9, "G-5"}, {12, "F-5"}},
		{{0, "E-5"}, {3, "G-5"}, {6, "C-6"}, {9, "G-4"}, {12, "E-5"}},
	}
	verse = [][]hit{
		{{0, "D-4"}, {4, "F-4"}, {8, "A-4"}, {12, "C-5"}},
		{{0, "A#3"}, {4, "D-4"}, {8, "F-4"}, {12, "D-4"}},
		{{0, "C-4"}, {4, "F-4"}, {8, "A-4"}, {12, "C-5"}},
		{{0, "G-3"}, {4, "C-4"}, {8, "E-4"}, {12, "G-4"}},
	}
	bridge = [][]hit{
		{{0, "A-4"}, {8, "D-5"}},
		{{0, "F-5"}, {8, "D-5"}},
		{{0, "C-5"}, {8, "A-4"}},
		{{0, "G-4"}, {8, "E-4"}},
	}
)

func abs(v int) int {
	if v < 0 {
		return -v
	}

	return v
}

// at returns the row at offset r within bar b
func at(b, r int) int {
	return b*song.RowsPerBar + r
}

// harmony returns the chord tone a fourth below note, so the harmony always fits the chord
func harmony(note string, cb chordBar) string {
	n, err := types.ParseNote(note)
	if err != nil {
		panic(err)
	}

	best, dist := "", 99

	for _, tone := range song.Triad(cb.chord, cb.padOct+2) {
		t, err := types.ParseNote(tone)
		if err != nil {
			panic(err)
		}

		for _, shift := range []int{-12, 0, 12, 24} {
			v := int(t) + shift
			if v < 1 || v > 96 {
				continue
			}

			if d := abs(int(n) - v - 5); d < dist {
				best, dist = types.Note(v).String(), d
			}
		}
	}

	return best
}

// withVol returns the options for a volume column value, left out when zero, followed by extra
func withVol(vol int, extra ...song.Opt) []song.Opt {
	var opts []song.Opt

	if vol != 0 {
		opts = append(opts, song.Vol(vol))
	}

	return append(opts, extra...)
}

// arp plays the bar's arpeggio, in eighths when step is 2 and sixteenths when it is 1
func (w pattern) arp(b int, cb chordBar, step, vol int) {
	for i, r := 0, 0; i < len(cb.arp); i, r = i+1, r+step {
		if r >= song.RowsPerBar {
			break
		}

		w.p.Hit(at(b, r), chArp, cb.arp[i], w.k.arp, withVol(vol)...)
	}
}

// bass plays the reese and sub for a bar: short in the verse, rolling in the drops
func (w pattern) bass(b int, cb chordBar, roll bool, vol int) {
	if roll {
		for i, note := range cb.bass {
			w.p.Hit(at(b, i*2), chReese, note, w.k.reese, withVol(vol)...)
		}

		w.p.Hit(at(b, 14), chReese, cb.bass[2], w.k.reese, withVol(vol)...)

		return
	}

	w.p.Hit(at(b, 0), chReese, cb.root, w.k.reese, withVol(vol)...)
	w.p.Hit(at(b, 8), chReese, cb.root, w.k.reese, withVol(vol)...)
}

// crash puts a crash at the top of a bar
func (w pattern) crash(b, vol int) {
	w.p.Hit(at(b, 0), chFx, "C-4", w.k.crash, withVol(vol)...)
}

// drums plays the break for a bar: 0 is sparse, 1 a full break, 2 a drop and 3 half time
func (w pattern) drums(b, style int) {
	kicks := []int{0, 10}

	switch style {
	case 0:
		kicks = []int{0, 8}
	case 1:
		kicks = []int{0, 6, 10}
	case 2:
		kicks = []int{0, 6, 10, 14}
	}

	for _, r := range kicks {
		var vol int

		if r == 0 {
			vol = 0x50
		}

		w.p.Hit(at(b, r), chKick, "C-2", w.k.kick, withVol(vol)...)
	}

	// the backbeat sits on 4 and 12, with ghost snares in the drops
	w.p.Hit(at(b, 4), chSnare, "C-4", w.k.snare, withVol(0x50)...)
	w.p.Hit(at(b, 12), chSnare, "C-4", w.k.snare, withVol(0x50)...)

	if style == 2 || style == 3 {
		w.p.Hit(at(b, 7), chSnare, "C-4", w.k.snare, withVol(0x24)...)
		w.p.Hit(at(b, 15), chSnare, "C-4", w.k.snare, withVol(0x2C)...)
	}

	if style == 2 {
		w.p.Hit(at(b, 11), chSnare, "C-4", w.k.snare, withVol(0x28)...)
	}

	openRows := []int{}

	if style == 1 || style == 2 {
		openRows = []int{2, 6, 10, 14}

		for i, r := range openRows {
			pan := 0x30

			if i%2 == 1 {
				pan = 0xD0
			}

			w.p.Hit(at(b, r), chHat, "C-5", w.k.openHat, withVol(0x28, song.Pan(pan))...)
		}
	}

	step := 2

	if style == 2 {
		step = 1
	}

	for i, r := 0, 0; r < song.RowsPerBar; i, r = i+1, r+step {
		if slices.Contains(openRows, r) {
			continue
		}

		vol := 0x24

		if r%4 == 0 {
			vol = 0x34
		}

		pan := 0x40

		if i%2 == 1 {
			pan = 0xC0
		}

		w.p.Hit(at(b, r), chHat, "C-5", w.k.hat, withVol(vol, song.Pan(pan))...)
	}
}

// fill plays a tom run at the end of the last bar
func (w pattern) fill(b int) {
	for i, note := range []string{"C-3", "G-3", "C-4", "F-4"} {
		w.p.Hit(at(b, 12+i), chFx, note, w.k.tom, withVol(0x30+i*4)...)
	}
}

// melody plays a phrase with optional vibrato, doubling it a fourth below on the harmony channel
func (w pattern) melody(b int, phrase []hit, cb chordBar, ins *song.Instrument, vol int, harmonise bool, extra ...song.Opt) {
	for _, h := range phrase {
		if h.note == "###" {
			// note offs cut the phrase, and there is no harmony to double
			w.p.Hit(at(b, h.row), chLead, "###", nil)
			w.p.Hit(at(b, h.row), chHarm, "###", nil)

			continue
		}

		w.p.Hit(at(b, h.row), chLead, h.note, ins, withVol(vol, extra...)...)

		if harmonise {
			w.p.Hit(at(b, h.row), chHarm, harmony(h.note, cb), w.k.harm, withVol(vol-0x14)...)
		}
	}
}

// pads holds the triad for a bar across three channels
func (w pattern) pads(b int, cb chordBar, vol int) {
	for i, note := range song.Triad(cb.chord, cb.padOct) {
		w.p.Hit(at(b, 0), chPadLo+i, note, w.k.pad, withVol(vol, song.Pan(0x50+i*0x30))...)
	}
}

// sub plays the root underneath, held for most of the bar
func (w pattern) sub(b int, cb chordBar, vol int) {
	w.p.Hit(at(b, 0), chSub, cb.root, w.k.sub, withVol(vol)...)
	w.p.Hit(at(b, 8), chSub, cb.root, w.k.sub, withVol(vol)...)
}

func breakdown(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression {
		w.pads(b, cb, 0x28)
		w.sub(b, cb, 0x28)
		w.arp(b, cb, 2, 0x20)
		w.melody(b, bridge[b], cb, k.lead, 0x38, true, song.Vibrato(4, 4))

		for _, r := range []int{2, 6, 10, 14} {
			w.p.Hit(at(b, r), chHat, "C-5", k.hat, withVol(0x1C, song.Pan(0x40))...)
		}
	}

	w.crash(0, 0x28)
	w.fill(3)

	return w.p
}

func bridgePattern(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression {
		w.pads(b, cb, 0x2C)
		w.sub(b, cb, 0x30)
		w.arp(b, cb, 2, 0x24)
		w.melody(b, bridge[b], cb, k.lead, 0x40, true, song.Vibrato(5, 5))
		w.drums(b, 3)
	}

	w.crash(0, 0x34)
	w.crash(2, 0x28)

	return w.p
}

func build(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression {
		w.pads(b, cb, 0x24)
		w.sub(b, cb, 0x30)
		w.arp(b, cb, b/2+1, 0x24)

		for i, note := range cb.bass {
			vol := 0x24 + i*2

			w.p.Hit(at(b, i*2), chReese, note, k.reese, withVol(vol)...)
		}

		for i, r := range []int{4, 8, 10, 12, 13, 14, 15} {
			w.p.Hit(at(b, r), chSnare, "C-4", k.snare, withVol(0x20+i*4)...)
		}

		for i, r := 0, 0; r < song.RowsPerBar; i, r = i+1, r+1 {
			w.p.Hit(at(b, r), chHat, "C-5", k.hat, withVol(0x20+i, song.Pan(0x40+i*8))...)
		}
	}

	// a noise swell across all four bars, cut at the end so it does not bleed into the drop
	w.p.Hit(0, chFx, "C-4", k.riser, withVol(0x24)...)
	w.p.Hit(rows-1, chFx, "###", nil)

	for _, h := range hookA[0][:4] {
		w.p.Hit(h.row, chLead, h.note, k.lead, withVol(0x38, song.Vibrato(4, 4))...)
	}

	w.fill(3)

	return w.p
}

func outro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression[:2] {
		w.pads(b, cb, 0x24)
		w.sub(b, cb, 0x28)
		w.arp(b, cb, 2, 0x20)
		w.melody(b, verse[b], cb, k.lead, 0x30, false, song.Vibrato(4, 3))
		w.drums(b, 0)
	}

	// a quiet bar of F to let the last chord ring out
	cb := progression[2]

	w.pads(2, cb, 0x28)
	w.sub(2, cb, 0x30)

	for i, note := range song.Triad("Dm", 4) {
		w.p.Hit(at(3, 0), chPadLo+i, note, k.pad, withVol(0x30)...)
	}

	w.p.Hit(at(3, 0), chLead, "D-5", k.lead, withVol(0x44, song.Vibrato(4, 4))...)
	w.p.Hit(at(3, 0), chSub, "D-2", k.sub, withVol(0x30)...)
	w.crash(3, 0x40)
	w.fill(3)

	for _, ch := range []int{chLead, chHarm, chArp, chReese, chSub, chPadLo, chPadMid, chPadHi} {
		w.p.Hit(rows-1, ch, "###", nil)
	}

	return w.p
}

func (w pattern) arrange(sec section) {
	for b, cb := range progression {
		w.pads(b, cb, sec.padVol)
		w.sub(b, cb, sec.subVol)
		w.bass(b, cb, sec.roll, sec.reeseVol)
		w.arp(b, cb, sec.arpStep, sec.arpVol)
		w.drums(b, sec.drumStyle)

		if sec.phrase != nil {
			w.melody(b, sec.phrase[b], cb, w.k.lead, sec.leadVol, sec.harmonise)
		}

		if sec.crashes && (b == 0 || b == 2) {
			w.crash(b, 0x30)
		}
	}

	if sec.fill {
		w.fill(3)
	}
}

func drop(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x24,
		arpStep: 1,
		arpVol: 0x28,
		reeseVol: 0x38,
		subVol: 0x30,
		drumStyle: 2,
		leadVol: 0x44,
		phrase: hookA,
		harmonise: true,
		roll: true,
		crashes: true,
		fill: true,
	})

	return w.p
}

func dropAnswer(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x24,
		arpStep: 1,
		arpVol: 0x2C,
		reeseVol: 0x38,
		subVol: 0x30,
		drumStyle: 2,
		leadVol: 0x40,
		phrase: hookB,
		harmonise: true,
		roll: true,
		crashes: true,
		fill: true,
	})

	return w.p
}

func finalDrop(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x28,
		arpStep: 1,
		arpVol: 0x30,
		reeseVol: 0x40,
		subVol: 0x30,
		drumStyle: 2,
		leadVol: 0x4C,
		phrase: hookA,
		harmonise: true,
		roll: true,
		crashes: true,
		fill: true,
	})

	return w.p
}

func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x24,
		arpStep: 2,
		arpVol: 0x20,
		reeseVol: 0x20,
		subVol: 0x28,
		drumStyle: 0,
		leadVol: 0x30,
	})
	w.crash(0, 0x28)

	for _, h := range verse[2] {
		w.p.Hit(at(2, h.row), chLead, h.note, k.lead, withVol(0x34, song.Vibrato(4, 3))...)
	}

	for _, h := range verse[3] {
		w.p.Hit(at(3, h.row), chLead, h.note, k.lead, withVol(0x38, song.Vibrato(4, 3))...)
	}

	return w.p
}

func versePattern(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x20,
		arpStep: 2,
		arpVol: 0x24,
		reeseVol: 0x30,
		subVol: 0x2C,
		drumStyle: 1,
		leadVol: 0x34,
		phrase: verse,
		crashes: true,
	})

	return w.p
}

func verseTwo(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x20,
		arpStep: 1,
		arpVol: 0x20,
		reeseVol: 0x34,
		subVol: 0x2C,
		drumStyle: 2,
		leadVol: 0x30,
		phrase: verse,
		harmonise: true,
		crashes: true,
	})

	return w.p
}

func main() {
	s := song.New("blue hour", 12)

	s.Timing(speed, tempo)

	k := &kit{
		lead: s.Instrument(song.Instrument{
			Name: "lead",
			Recipe: leadRecipe,
			Loop: true,
			Volume: 40,
			Pan: 150,
		}),
		harm: s.Instrument(song.Instrument{
			Name: "harmony",
			Recipe: harmRecipe,
			Loop: true,
			Volume: 32,
			Pan: 106,
			Fadeout: 512,
		}),
		arp: s.Instrument(song.Instrument{
			Name: "arp",
			Recipe: arpRecipe,
			Volume: 36,
			Pan: 170,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 7,
					Y: 0,
				},
			},
		}),
		reese: s.Instrument(song.Instrument{
			Name: "reese",
			Recipe: reeseRecipe,
			Loop: true,
			Volume: 44,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 3,
					Y: 64,
				},
			},
		}),
		sub: s.Instrument(song.Instrument{
			Name: "sub",
			Recipe: subRecipe,
			Loop: true,
			Volume: 40,
			Pan: 128,
		}),
		pad: s.Instrument(song.Instrument{
			Name: "pad",
			Recipe: padRecipe,
			Loop: true,
			Volume: 32,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 14,
					Y: 64,
				},
			},
			Fadeout: 1024,
		}),
		kick: s.Instrument(song.Instrument{
			Name: "kick",
			Recipe: kickRecipe,
			Volume: 56,
			Pan: 128,
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
					X: 12,
					Y: 0,
				},
			},
		}),
		hat: s.Instrument(song.Instrument{
			Name: "hat",
			Recipe: hatRecipe,
			Volume: 34,
			Pan: 120,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 2,
					Y: 0,
				},
			},
		}),
		openHat: s.Instrument(song.Instrument{
			Name: "open hat",
			Recipe: openHatRecipe,
			Volume: 30,
			Pan: 100,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 11,
					Y: 0,
				},
			},
		}),
		crash: s.Instrument(song.Instrument{
			Name: "crash",
			Recipe: crashRecipe,
			Volume: 34,
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
		tom: s.Instrument(song.Instrument{
			Name: "tom",
			Recipe: tomRecipe,
			Volume: 44,
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
		riser: s.Instrument(song.Instrument{
			Name: "riser",
			Recipe: riserRecipe,
			Loop: true,
			Volume: 28,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 380,
					Y: 64,
				},
			},
			Fadeout: 256,
		}),
	}
	patterns := []*song.Pattern{
		intro(s, k),
		versePattern(s, k),
		build(s, k),
		drop(s, k),
		dropAnswer(s, k),
		breakdown(s, k),
		verseTwo(s, k),
		bridgePattern(s, k),
		finalDrop(s, k),
		outro(s, k),
	}

	for _, i := range []int{0, 0, 1, 1, 2, 3, 4, 5, 6, 1, 2, 3, 4, 7, 8, 3, 4, 6, 9} {
		s.Play(patterns[i])
	}

	// write next to this file, wherever the program is run from
	_, src, _, _ := runtime.Caller(0)

	if err := s.Write(filepath.Join(filepath.Dir(src), "song.xm")); err != nil {
		log.Fatal(err)
	}
}
