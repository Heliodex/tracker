// Command track8 writes "neon dusk", a dark synthwave track, as an XM module, song.xm, in this folder.
//
// Build it from anywhere inside the repository with go run ./tracks/track8, or with tracker build.
package main

// neon dusk – OpenCode

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/song"
)

type (
	// chordBar is one bar of the Am - F - C - G progression
	chordBar struct {
		root, fifth, octave string
		pad [3]string
		arp [8]string
		bass [4]string
	}

	// hit is one melody note at a row within the bar
	hit struct {
		row int
		note string
	}

	// kit holds the instruments, numbered as they are added to the song
	kit struct {
		lead, bass, pad, arp, kick, snare, hat, openHat, crash *song.Instrument
	}

	// pattern wraps a song pattern with the instruments to play it with
	pattern struct {
		p *song.Pattern
		k *kit
	}

	// section describes the layers and intensity of a four-bar pattern
	section struct {
		padVol, arpStep, arpVol, bassStep, bassVol int
		drumStyle, leadVol int
		phrase [][]hit
		fill bool
	}
)

// channels, in the order they appear in the tracker
const (
	chLead = iota
	chArp
	chBass
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
	speed = 6
	tempo = 120
)

var (
	leadRecipe = recipe.New(256).Sixteen().
		Saw(261.34375, 0.22).Saw(262.1, 0.22).Phase(0.5).
		Square(130.671875, 0.08).Duty(0.25).
		Sine(522.6875, 0.04).String()

	bassRecipe = recipe.New(256).Sixteen().
		Saw(261.34375, 0.24).Saw(262.0, 0.16).Phase(0.5).
		Sine(130.671875, 0.28).Sine(65.3359375, 0.12).String()

	padRecipe = recipe.New(512).Sixteen().
		Triangle(261.34375, 0.18).Saw(130.671875, 0.08).
		Sine(65.3359375, 0.10).String()

	arpRecipe = recipe.New(512).Sixteen().
		Square(261.34375, 0.18).Duty(0.125).
		Sine(522.6875, 0.06).String()

	kickRecipe = recipe.New(2000).Sixteen().
		Sine(55, 0.9).Sine(110, 0.25).Noise(2000, 0.03).String()

	snareRecipe = recipe.New(2400).Sixteen().
		Noise(1200, 0.55).Noise(3200, 0.2).
		Triangle(180, 0.18).String()

	hatRecipe = recipe.New(500).Sixteen().
		Noise(4000, 0.4).Square(4000, 0.03).String()

	openHatRecipe = recipe.New(1800).Sixteen().
		Noise(3500, 0.34).Square(3500, 0.03).String()

	crashRecipe = recipe.New(8000).Sixteen().
		Noise(1500, 0.45).Noise(4000, 0.2).String()

	progression = []chordBar{
		{
			root: "A-2", fifth: "E-3", octave: "A-3",
			pad: [3]string{"A-3", "C-4", "E-4"},
			arp: [8]string{"A-3", "C-4", "E-4", "A-4", "E-4", "C-4", "A-3", "E-3"},
			bass: [4]string{"A-2", "A-2", "E-2", "A-2"},
		},
		{
			root: "F-2", fifth: "C-3", octave: "F-3",
			pad: [3]string{"F-3", "A-3", "C-4"},
			arp: [8]string{"F-3", "A-3", "C-4", "F-4", "C-4", "A-3", "F-3", "C-3"},
			bass: [4]string{"F-2", "F-2", "C-3", "F-2"},
		},
		{
			root: "C-3", fifth: "G-3", octave: "C-4",
			pad: [3]string{"C-4", "E-4", "G-4"},
			arp: [8]string{"C-4", "E-4", "G-4", "C-5", "G-4", "E-4", "C-4", "G-3"},
			bass: [4]string{"C-3", "C-3", "G-2", "C-3"},
		},
		{
			root: "G-2", fifth: "D-3", octave: "G-3",
			pad: [3]string{"G-3", "B-3", "D-4"},
			arp: [8]string{"G-3", "B-3", "D-4", "G-4", "D-4", "B-3", "G-3", "D-3"},
			bass: [4]string{"G-2", "G-2", "D-3", "G-2"},
		},
	}

	hookA = [][]hit{
		{{0, "A-4"}, {3, "C-5"}, {4, "E-5"}, {6, "D-5"}, {8, "C-5"}, {12, "A-4"}},
		{{0, "F-4"}, {4, "A-4"}, {6, "C-5"}, {8, "A-4"}, {12, "F-4"}, {14, "G-4"}},
		{{0, "E-5"}, {2, "D-5"}, {4, "C-5"}, {8, "G-4"}, {10, "C-5"}, {12, "E-5"}},
		{{0, "D-5"}, {4, "B-4"}, {8, "G-4"}, {12, "B-4"}, {14, "D-5"}},
	}
	hookB = [][]hit{
		{{0, "E-5"}, {4, "C-5"}, {8, "A-4"}, {10, "C-5"}, {12, "E-5"}, {14, "D-5"}},
		{{0, "C-5"}, {4, "A-4"}, {8, "F-4"}, {12, "A-4"}},
		{{0, "G-4"}, {2, "C-5"}, {4, "E-5"}, {8, "G-5"}, {12, "E-5"}},
		{{0, "D-5"}, {4, "G-5"}, {8, "B-4"}, {12, "D-5"}, {14, "B-4"}},
	}
	introLead = [][]hit{
		{},
		{{8, "A-4"}, {12, "C-5"}},
		{{0, "E-5"}, {8, "D-5"}},
		{{0, "C-5"}, {8, "B-4"}, {12, "G-4"}},
	}
)

func at(b, r int) int {
	return b*song.RowsPerBar + r
}

func withVol(vol int, extra ...song.Opt) []song.Opt {
	var opts []song.Opt
	if vol != 0 {
		opts = append(opts, song.Vol(vol))
	}
	return append(opts, extra...)
}

func (w pattern) pads(b int, cb chordBar, vol int) {
	if vol == 0 {
		return
	}
	pans := [3]int{0x40, 0x80, 0xC0}
	for i, note := range cb.pad {
		w.p.Hit(at(b, 0), chPadLo+i, note, w.k.pad, withVol(vol, song.Pan(pans[i]))...)
	}
}

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

func (w pattern) bass(b int, cb chordBar, step, vol int) {
	if vol == 0 {
		return
	}
	for r := 0; r < song.RowsPerBar; r += step {
		note := cb.bass[r/4%len(cb.bass)]
		cellVol := vol
		if r%4 == 0 {
			cellVol -= 0x08
		}
		w.p.Hit(at(b, r), chBass, note, w.k.bass, withVol(cellVol)...)
	}
}

func (w pattern) crash(b, vol int) {
	if vol != 0 {
		w.p.Hit(at(b, 0), chFx, "C-4", w.k.crash, withVol(vol)...)
	}
}

func (w pattern) drums(b, style int) {
	if style == 0 {
		return
	}
	for _, r := range []int{0, 4, 8, 12} {
		vol := 0x3C
		if style >= 2 {
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
	if style == 2 {
		for _, r := range []int{2, 6, 10, 14} {
			w.p.Hit(at(b, r), chHat, "C-5", w.k.openHat, withVol(0x28, song.Pan(0xC0))...)
		}
	}
	step := 4
	if style >= 2 {
		step = 2
	}
	for i, r := 0, 0; r < song.RowsPerBar; i, r = i+1, r+step {
		if style == 2 && (r == 2 || r == 6 || r == 10 || r == 14) {
			continue
		}
		vol := 0x24
		if r%4 == 0 {
			vol = 0x30
		}
		pan := 0x58
		if i%2 == 1 {
			pan = 0xA8
		}
		w.p.Hit(at(b, r), chHat, "C-5", w.k.hat, withVol(vol, song.Pan(pan))...)
	}
	if style == 3 {
		for i, r := range []int{8, 10, 12, 13, 14, 15} {
			w.p.Hit(at(b, r), chSnare, "C-4", w.k.snare, withVol(0x24+i*5)...)
		}
	}
}

func (w pattern) fill(b int) {
	for i, r := range []int{12, 13, 14, 15} {
		vol := 0x48
		if i > 0 {
			vol = 0x30 + i*8
		}
		w.p.Hit(at(b, r), chSnare, "C-4", w.k.snare, withVol(vol)...)
	}
}

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

func (w pattern) arrange(sec section, chords []chordBar) {
	for b, cb := range chords {
		w.pads(b, cb, sec.padVol)
		w.arp(b, cb, sec.arpStep, sec.arpVol)
		w.bass(b, cb, sec.bassStep, sec.bassVol)
		w.drums(b, sec.drumStyle)
		if sec.phrase != nil {
			w.melody(b, sec.phrase[b], sec.leadVol)
		}
	}
	if sec.fill {
		w.fill(3)
	}
}

func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	for b, cb := range progression {
		w.pads(b, cb, 0x20)
		w.arp(b, cb, 4, 0x1C)
		w.melody(b, introLead[b], 0x34)
		if b >= 2 {
			w.bass(b, cb, 4, 0x24)
		}
	}
	w.crash(0, 0x20)
	return w.p
}

func verse(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	w.arrange(section{
		padVol: 0x24, arpStep: 2, arpVol: 0x24,
		bassStep: 4, bassVol: 0x30, drumStyle: 1, leadVol: 0x38, phrase: hookA,
	}, progression)
	w.crash(0, 0x28)
	return w.p
}

func build(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	for b, cb := range progression {
		w.pads(b, cb, 0x28)
		step := 4
		if b >= 2 {
			step = 2
		}
		w.bass(b, cb, step, 0x30+b*2)
		arpStep := 2
		if b >= 2 {
			arpStep = 1
		}
		w.arp(b, cb, arpStep, 0x24+b*2)
		if b == 3 {
			w.drums(b, 3)
		} else {
			w.drums(b, 2)
		}
	}
	w.crash(0, 0x28)
	w.melody(3, hookA[0][:4], 0x3C)
	return w.p
}

func chorus(s *song.Song, k *kit, answer, final bool) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	phrase := hookA
	if answer {
		phrase = hookB
	}
	sec := section{
		padVol: 0x2C, arpStep: 1, arpVol: 0x30,
		bassStep: 2, bassVol: 0x3C, drumStyle: 2, leadVol: 0x48, phrase: phrase, fill: true,
	}
	if final {
		sec.padVol = 0x30
		sec.arpVol = 0x34
		sec.bassVol = 0x40
		sec.leadVol = 0x4C
	}
	w.arrange(sec, progression)
	w.crash(0, 0x38)
	if !answer {
		w.crash(2, 0x28)
	}
	if final {
		w.crash(2, 0x30)
	}
	w.p.Hit(at(2, 14), chKick, "C-2", k.kick, withVol(0x30)...)
	return w.p
}

func breakdown(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	for b, cb := range progression {
		w.pads(b, cb, 0x28)
		w.arp(b, cb, 2, 0x20)
		w.bass(b, cb, 4, 0x28)
		w.melody(b, hookA[b][:min(3, len(hookA[b]))], 0x38)
		if b >= 2 {
			w.drums(b, 1)
		}
	}
	w.crash(0, 0x24)
	for i, vol := range []int{0x20, 0x28, 0x30, 0x38} {
		w.p.Hit(at(3, 12+i), chSnare, "C-4", k.snare, withVol(vol)...)
	}
	return w.p
}

func outro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	for b, cb := range progression {
		w.pads(b, cb, 0x24)
		w.arp(b, cb, 4, 0x1C)
		w.melody(b, introLead[b], 0x34)
		if b < 2 {
			w.bass(b, cb, 4, 0x28)
			w.drums(b, 1)
		}
	}
	cb := progression[0]
	w.pads(3, cb, 0x28)
	w.bass(3, cb, 4, 0x2C)
	w.melody(3, []hit{{0, "A-4"}, {8, "E-5"}, {12, "A-4"}}, 0x38)
	w.crash(3, 0x30)
	for _, ch := range []int{chLead, chArp, chBass, chPadLo, chPadMid, chPadHi} {
		w.p.Hit(rows-2, ch, "###", nil)
	}
	return w.p
}

func main() {
	s := song.New("neon dusk", 10)
	s.Timing(speed, tempo)

	k := &kit{
		lead: s.Instrument(song.Instrument{
			Name: "saw lead", Recipe: leadRecipe, Loop: true,
			Volume: 34, Pan: 96,
			Envelope: []song.Point{{X: 0, Y: 0}, {X: 2, Y: 64}},
			Fadeout: 512,
		}),
		bass: s.Instrument(song.Instrument{
			Name: "bass", Recipe: bassRecipe, Loop: true,
			Volume: 48, Pan: 128,
			Envelope: []song.Point{{X: 0, Y: 0}, {X: 2, Y: 64}},
		}),
		pad: s.Instrument(song.Instrument{
			Name: "pad", Recipe: padRecipe, Loop: true,
			Volume: 28, Pan: 128,
			Envelope: []song.Point{{X: 0, Y: 0}, {X: 16, Y: 64}},
			Fadeout: 1024,
		}),
		arp: s.Instrument(song.Instrument{
			Name: "arp", Recipe: arpRecipe,
			Volume: 30, Pan: 128,
			Envelope: []song.Point{{X: 0, Y: 64}, {X: 6, Y: 0}},
		}),
		kick: s.Instrument(song.Instrument{
			Name: "kick", Recipe: kickRecipe,
			Volume: 58, Pan: 128,
			Envelope: []song.Point{{X: 0, Y: 64}, {X: 14, Y: 0}},
		}),
		snare: s.Instrument(song.Instrument{
			Name: "snare", Recipe: snareRecipe,
			Volume: 44, Pan: 136,
			Envelope: []song.Point{{X: 0, Y: 64}, {X: 16, Y: 0}},
		}),
		hat: s.Instrument(song.Instrument{
			Name: "hat", Recipe: hatRecipe,
			Volume: 30, Pan: 128,
			Envelope: []song.Point{{X: 0, Y: 64}, {X: 3, Y: 0}},
		}),
		openHat: s.Instrument(song.Instrument{
			Name: "open hat", Recipe: openHatRecipe,
			Volume: 26, Pan: 192,
			Envelope: []song.Point{{X: 0, Y: 64}, {X: 14, Y: 0}},
		}),
		crash: s.Instrument(song.Instrument{
			Name: "crash", Recipe: crashRecipe,
			Volume: 28, Pan: 128,
			Envelope: []song.Point{{X: 0, Y: 64}, {X: 64, Y: 0}},
		}),
	}

	introP := intro(s, k)
	verseP := verse(s, k)
	buildP := build(s, k)
	chorusP := chorus(s, k, false, false)
	answerP := chorus(s, k, true, false)
	finalP := chorus(s, k, false, true)
	breakP := breakdown(s, k)
	outroP := outro(s, k)

	s.Play(introP, verseP, buildP, chorusP, answerP, breakP, buildP, chorusP, answerP, finalP, outroP)

	_, src, _, _ := runtime.Caller(0)
	if err := s.Write(filepath.Join(filepath.Dir(src), "song.xm")); err != nil {
		log.Fatal(err)
	}
}
