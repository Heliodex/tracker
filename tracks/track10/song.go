// Command track10 writes "low tide", a deep house track in F minor, as an XM module, song.xm, in this folder.
//
// Build it from anywhere inside the repository with go run ./tracks/track10, or with tracker build.
package main

// low tide – DeepSeek V4.1 Flash

import (
	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/song"
	"log"
	"strconv"
)

type (
	// chordBar is one bar of the Fm9 - Db - Abmaj7 - Eb7 progression
	chordBar struct {
		name, root, octave, fifth string
		pad, stab [3]string
		arp [16]string
	}

	// hit is one melody note at a row within the bar
	hit struct {
		row int
		note string
	}

	// kit holds the instruments, numbered as they are added to the song
	kit struct {
		lead, arp, stab, bass, pad, kick, snare, hat, openHat, shaker, crash, riser *song.Instrument
	}

	// pattern wraps a song pattern with the instruments to play it with
	pattern struct {
		p *song.Pattern
		k *kit
	}

	// section describes the layers and intensity of a four-bar pattern
	section struct {
		padVol, stabVol, arpStep, arpVol, bassStyle, bassVol, drumStyle, leadVol int
		phrase [][]hit
		fill, riser bool
	}
)

// channels, in the order they appear in the tracker
const (
	chLead = iota
	chArp
	chStab1
	chStab2
	chStab3
	chBass
	chPad1
	chPad2
	chPad3
	chKick
	chSnare
	chHat
	chOpenHat
	chShaker
	chFx
	chRiser
	chSpark
)

const (
	rows = 64
	// Six ticks per row at 118 BPM gives a sixteenth-note row and a 2.03-second bar
	speed = 6
	tempo = 118
)

var (
	padCh = [3]int{chPad1, chPad2, chPad3}

	// stabCh and padCh are the three voices of the chord layers, spread across the field
	stabCh = [3]int{chStab1, chStab2, chStab3}

	// A warm, slightly hollow lead that sits between a flute and a voice
	leadRecipe = recipe.New(256).Triangle(recipe.Note(256, "C-4"), 0.32).Sine(recipe.Note(256, "C-4"), 0.12).Sine(recipe.Note(256, "C-5"), 0.07).Square(recipe.Note(256, "C-4"), 0.04).Duty(0.3).String()

	// A round mallet for the arpeggio
	arpRecipe = recipe.New(1024).Triangle(recipe.Note(1024, "C-4"), 0.24).Sine(recipe.Note(1024, "C-5"), 0.1).Sine(1046.875, 0.03).String()

	// A Rhodes-like chord voice, mostly sine with a bell on top
	stabRecipe = recipe.New(2048).Sine(recipe.Note(2048, "C-4"), 0.34).Sine(recipe.Note(2048, "C-5"), 0.1).Sine(recipe.Note(2048, "G-5"), 0.045).Triangle(recipe.Note(2048, "C-4"), 0.08).String()

	// A deep sine bass with a little grit, tuned to C-4
	bassRecipe = recipe.New(256).Sine(recipe.Note(256, "C-4"), 0.28).Sine(recipe.Note(256, "C-3"), 0.12).Saw(recipe.Note(256, "C-4"), 0.1).String()

	// A wide, slow pad
	padRecipe = recipe.New(512).Saw(recipe.Note(512, "C-4"), 0.1).Triangle(recipe.Note(512, "C-4"), 0.14).Sine(recipe.Note(512, "C-4"), 0.1).Saw(recipe.Note(512, "C-5"), 0.03).String()

	// A kick with a click, played at C-2 so its 200 Hz sine lands at 50 Hz
	kickRecipe = recipe.New(2000).Sine(200, 0.85).Sine(400, 0.15).Noise(3000, 0.02).String()

	// A tight clap over a small tonal body
	snareRecipe = recipe.New(2400).Noise(1000, 0.45).Noise(3200, 0.18).Triangle(190, 0.1).String()

	// A short closed hat
	hatRecipe = recipe.New(400).Noise(4000, 0.36).Square(4100, 0.03).String()

	// A longer open hat for the offbeats
	openHatRecipe = recipe.New(1600).Noise(3500, 0.34).Square(3500, 0.03).String()

	// A small shaker for the sixteenths
	shakerRecipe = recipe.New(700).Noise(4000, 0.28).String()

	// A long bright crash
	crashRecipe = recipe.New(8000).Noise(1200, 0.4).Noise(4000, 0.2).String()

	// A noise swell that rises across a whole pattern
	riserRecipe = recipe.New(60000).Noise(1500, 0.32).Noise(3500, 0.1).String()
	progression = []chordBar{
		{
			name: "Fm9",
			root: "F-2",
			octave: "F-3",
			fifth: "C-3",
			pad: [3]string{"F-3", "G#3", "G-4"},
			stab: [3]string{"F-4", "G#4", "C-5"},
			arp: [16]string{
				"F-3",
				"C-4",
				"G#4",
				"C-5",
				"F-5",
				"C-5",
				"G#4",
				"C-4",
				"F-3",
				"G#4",
				"C-5",
				"F-5",
				"G-5",
				"F-5",
				"C-5",
				"G#4",
			},
		},
		{
			name: "Db",
			root: "C#2",
			octave: "C#3",
			fifth: "G#2",
			pad: [3]string{"C#3", "F-3", "G#3"},
			stab: [3]string{"C#4", "F-4", "G#4"},
			arp: [16]string{
				"C#3",
				"G#3",
				"F-4",
				"G#4",
				"C#5",
				"G#4",
				"F-4",
				"G#3",
				"C#3",
				"F-4",
				"G#4",
				"C#5",
				"D#5",
				"C#5",
				"G#4",
				"F-4",
			},
		},
		{
			name: "Abmaj7",
			root: "G#2",
			octave: "G#3",
			fifth: "D#3",
			pad: [3]string{"G#3", "C-4", "G-4"},
			stab: [3]string{"D#4", "G#4", "C-5"},
			arp: [16]string{
				"G#3",
				"D#4",
				"C-5",
				"D#5",
				"G#5",
				"D#5",
				"C-5",
				"D#4",
				"G#3",
				"D#4",
				"C-5",
				"D#5",
				"A#5",
				"G#5",
				"D#5",
				"C-5",
			},
		},
		{
			name: "Eb7",
			root: "D#2",
			octave: "D#3",
			fifth: "A#2",
			pad: [3]string{"D#3", "G-3", "C#4"},
			stab: [3]string{"C#4", "G-4", "A#4"},
			arp: [16]string{
				"D#3",
				"A#3",
				"G-4",
				"A#4",
				"D#5",
				"A#4",
				"G-4",
				"A#3",
				"D#3",
				"G-4",
				"A#4",
				"D#5",
				"F-5",
				"D#5",
				"A#4",
				"G-4",
			},
		},
	}

	// hookA is the main melody; hookB is its higher answer
	hookA = [][]hit{
		{{0, "C-5"}, {3, "G#4"}, {4, "C-5"}, {6, "C#5"}, {8, "C-5"}, {12, "G#4"}},
		{{0, "F-4"}, {4, "G#4"}, {6, "C-5"}, {8, "C#5"}, {10, "C-5"}, {12, "G#4"}},
		{{0, "D#5"}, {2, "C-5"}, {4, "G#4"}, {8, "C-5"}, {10, "D#5"}, {12, "F-5"}},
		{{0, "D#5"}, {4, "C#5"}, {6, "C-5"}, {8, "A#4"}, {12, "C-5"}, {14, "C#5"}},
	}
	hookB = [][]hit{
		{{0, "F-5"}, {2, "D#5"}, {4, "C-5"}, {8, "G#4"}, {12, "C-5"}},
		{{0, "G#5"}, {4, "F-5"}, {6, "C#5"}, {8, "C-5"}, {12, "G#4"}},
		{{0, "C-5"}, {2, "D#5"}, {4, "G#5"}, {8, "G-5"}, {10, "D#5"}, {12, "C-5"}},
		{{0, "A#4"}, {4, "C-5"}, {8, "C#5"}, {12, "D#5"}, {14, "F-5"}},
	}
	verseLead = [][]hit{
		{{0, "C-5"}, {6, "G#4"}, {8, "C-5"}, {14, "C#5"}},
		{{0, "C-5"}, {6, "G#4"}, {8, "F-4"}, {12, "G#4"}},
		{{0, "D#5"}, {6, "C-5"}, {8, "G#4"}, {12, "C-5"}},
		{{0, "C#5"}, {6, "A#4"}, {8, "C-5"}, {14, "F-4"}},
	}
	introLead = [][]hit{
		{},
		{{8, "C-5"}, {12, "C#5"}},
		{{0, "D#5"}, {8, "C-5"}},
		{{0, "G#4"}, {8, "A#4"}, {12, "C-5"}},
	}
)

// at returns the row at offset r within bar b
// upOctave returns note an octave higher, for the bell that doubles the last chorus
func upOctave(note string) string {
	octave, err := strconv.Atoi(note[2:])
	if err != nil || octave >= 7 {
		panic("note cannot be doubled an octave up: " + note)
	}

	return note[:2] + strconv.Itoa(octave+1)
}

// withVol returns a volume-column option, omitted when zero, followed by extra
// arp plays the bar's arpeggio in eighths or sixteenths, alternating across the field
func (w pattern) arp(b int, cb chordBar, step, vol int) {
	if vol == 0 {
		return
	}

	for i, r := 0, 0; i < len(cb.arp); i, r = i+1, r+step {
		if r >= song.RowsPerBar {
			break
		}

		pan := 0x58

		if i%2 == 1 {
			pan = 0xA8
		}

		w.p.Hit(song.At(b, r), chArp, cb.arp[i], w.k.arp, song.VolOpts(vol, song.Pan(pan))...)
	}
}

// bass plays a syncopated figure. Style 1 is sparse, style 2 adds the fifth and the octave
func (w pattern) bass(b int, cb chordBar, style, vol int) {
	if vol == 0 {
		return
	}

	figure := []struct {
		row int
		note string
	}{
		{0, cb.root},
		{6, cb.octave},
		{8, cb.root},
		{14, cb.octave},
	}

	if style >= 2 {
		figure = []struct {
			row int
			note string
		}{
			{0, cb.root},
			{3, cb.root},
			{6, cb.octave},
			{8, cb.root},
			{11, cb.fifth},
			{14, cb.octave},
		}
	}

	for _, f := range figure {
		cellVol := vol

		// duck under the kick without needing a sidechain effect
		if f.row%4 == 0 {
			cellVol -= 0x08
		}

		w.p.Hit(song.At(b, f.row), chBass, f.note, w.k.bass, song.VolOpts(cellVol)...)
	}
}

// crash puts a crash at the top of a bar
func (w pattern) crash(b, vol int) {
	if vol != 0 {
		w.p.Hit(song.At(b, 0), chFx, "C-4", w.k.crash, song.VolOpts(vol)...)
	}
}

// drums plays the house groove. Style 1 is a soft pulse, 2 is full, 3 adds sixteenth shaker, 4 builds
func (w pattern) drums(b, style int) {
	if style == 0 {
		return
	}

	for _, r := range []int{0, 4, 8, 12} {
		vol := 0x3C

		if style == 1 {
			vol = 0x30
		}

		if r == 0 {
			vol += 0x04
		}

		w.p.Hit(song.At(b, r), chKick, "C-2", w.k.kick, song.VolOpts(vol)...)
	}

	if style >= 2 {
		w.p.Hit(song.At(b, 4), chSnare, "C-4", w.k.snare, song.VolOpts(0x44)...)
		w.p.Hit(song.At(b, 12), chSnare, "C-4", w.k.snare, song.VolOpts(0x44)...)
	}

	if style >= 2 {
		for _, r := range []int{2, 6, 10, 14} {
			w.p.Hit(song.At(b, r), chOpenHat, "C-5", w.k.openHat, song.VolOpts(0x2C, song.Pan(0xD0))...)
		}
	}

	step := 4

	if style >= 2 {
		step = 2
	}

	if style == 4 {
		step = 1
	}

	for i, r := 0, 0; r < song.RowsPerBar; i, r = i+1, r+step {
		// leave the offbeat to the open hat, except in the build
		if style >= 2 && style < 4 && r%4 == 2 {
			continue
		}

		vol := 0x1C

		if r%4 == 0 {
			vol = 0x2C
		}

		pan := 0x58

		if i%2 == 1 {
			pan = 0xA8
		}

		w.p.Hit(song.At(b, r), chHat, "C-5", w.k.hat, song.VolOpts(vol, song.Pan(pan))...)
	}

	if style == 3 {
		for r := 1; r < song.RowsPerBar; r += 2 {
			w.p.Hit(song.At(b, r), chShaker, "C-5", w.k.shaker, song.VolOpts(0x24, song.Pan(0xC8))...)
		}
	}

	if style == 4 {
		for i, r := range []int{8, 10, 12, 13, 14, 15} {
			w.p.Hit(song.At(b, r), chSnare, "C-4", w.k.snare, song.VolOpts(0x24+i*5)...)
		}
	}
}

// fill rolls the clap into the next bar
func (w pattern) fill(b int) {
	for i, r := range []int{12, 13, 14, 15} {
		vol := 0x44

		if i > 0 {
			vol = 0x2C + i*8
		}

		w.p.Hit(song.At(b, r), chSnare, "C-4", w.k.snare, song.VolOpts(vol)...)
	}
}

// melody plays a phrase of lead notes, each held until just before the next
func (w pattern) melody(b int, phrase []hit, vol int) {
	if vol == 0 {
		return
	}

	for i, h := range phrase {
		w.p.Hit(song.At(b, h.row), chLead, h.note, w.k.lead, song.VolOpts(vol, song.Vibrato(4, 2))...)

		off := song.RowsPerBar - 1

		if i+1 < len(phrase) {
			off = phrase[i+1].row - 1
		}

		if off > h.row {
			w.p.Hit(song.At(b, off), chLead, "###", nil)
		}
	}
}

// pads holds the bar's triad across the pad channels, spread across the stereo field
func (w pattern) pads(b int, cb chordBar, vol int) {
	if vol == 0 {
		return
	}

	pans := [3]int{0x40, 0x80, 0xC0}

	for i, note := range cb.pad {
		w.p.Hit(song.At(b, 0), padCh[i], note, w.k.pad, song.VolOpts(vol, song.Pan(pans[i]))...)
	}
}

// riser swells across the pattern, stopping before its last row
func (w pattern) riser(vol int) {
	w.p.Hit(0, chRiser, "C-4", w.k.riser, song.VolOpts(vol)...)
	w.p.Hit(rows-1, chRiser, "###", nil)
}

// stabs plays the bar's chord as short Rhodes hits at the given rows
func (w pattern) stabs(b int, cb chordBar, rows []int, vol int) {
	if vol == 0 {
		return
	}

	pans := [3]int{0x60, 0x80, 0xA0}

	for _, r := range rows {
		for i, note := range cb.stab {
			w.p.Hit(song.At(b, r), stabCh[i], note, w.k.stab, song.VolOpts(vol, song.Pan(pans[i]))...)
		}
	}
}

// build lifts the track into the last drops, with sixteenths, a snare roll and a riser
func build(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression {
		w.pads(b, cb, 0x2E+b*2)
		w.stabs(b, cb, []int{3, 6, 11, 14}, 0x24+b*2)

		step := 2

		if b >= 2 {
			step = 1
		}

		w.arp(b, cb, step, 0x24+b*2)
		w.bass(b, cb, 2, 0x2C+b*2)

		style := 2

		if b >= 2 {
			style = 4
		}

		w.drums(b, style)
	}

	w.crash(0, 0x2C)
	w.riser(0x2C)
	w.melody(3, hookA[2][:3], 0x38)

	return w.p
}

func interlude(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression {
		w.pads(b, cb, 0x2E)
		w.stabs(b, cb, []int{6, 14}, 0x24)
		w.arp(b, cb, 2, 0x24)
		w.bass(b, cb, 1, 0x28)
		w.drums(b, 2)

		if b >= 2 {
			w.melody(b, hookB[b], 0x38)
		}
	}

	w.crash(0, 0x28)

	return w.p
}

func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression {
		w.pads(b, cb, 0x22+b*2)
		w.arp(b, cb, 4, 0x20)

		if b >= 1 {
			w.stabs(b, cb, []int{6}, 0x18)
		}

		if b >= 2 {
			w.bass(b, cb, 1, 0x24)
		}
	}

	w.crash(0, 0x1C)
	w.drums(3, 1)

	return w.p
}

func outro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range progression[:2] {
		w.pads(b, cb, 0x2E)
		w.stabs(b, cb, []int{6}, 0x20)
		w.arp(b, cb, 4, 0x20)
		w.bass(b, cb, 1, 0x28)
		w.drums(b, 1)
	}

	// the last two bars hold a soft Fm9, then let it ring
	cb := progression[0]

	w.pads(2, cb, 0x32)
	w.stabs(2, cb, []int{0}, 0x28)
	w.arp(2, cb, 4, 0x20)
	w.bass(2, cb, 1, 0x30)
	w.crash(2, 0x2C)
	w.melody(2, []hit{{0, "C-5"}, {8, "G#4"}}, 0x38)
	w.pads(3, cb, 0x36)
	w.p.Hit(song.At(3, 0), chBass, "F-2", k.bass, song.Vol(0x34))
	w.stabs(3, cb, []int{0}, 0x2C)
	w.crash(3, 0x30)

	for _, ch := range []int{chLead, chArp, chStab1, chStab2, chStab3, chBass, chPad1, chPad2, chPad3} {
		w.p.Hit(rows-2, ch, "###", nil)
	}

	return w.p
}

// arrange lays down the shared layers of a section over the progression
func (w pattern) arrange(sec section, chords []chordBar) {
	for b, cb := range chords {
		w.pads(b, cb, sec.padVol)
		w.stabs(b, cb, []int{3, 6, 11, 14}, sec.stabVol)
		w.arp(b, cb, sec.arpStep, sec.arpVol)
		w.bass(b, cb, sec.bassStyle, sec.bassVol)
		w.drums(b, sec.drumStyle)

		if sec.phrase != nil {
			w.melody(b, sec.phrase[b], sec.leadVol)
		}
	}

	if sec.fill {
		w.fill(3)
	}

	if sec.riser {
		w.riser(0x2C)
	}
}

func breakdown(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x32,
		stabVol: 0x20,
		arpStep: 2,
		arpVol: 0x28,
		drumStyle: 0,
		leadVol: 0x3C,
		phrase: hookA,
		riser: true,
	}, progression)
	w.crash(0, 0x2C)

	// a soft pulse and a clap roll bring the drop in
	for _, r := range []int{0, 8} {
		w.p.Hit(song.At(3, r), chKick, "C-2", k.kick, song.VolOpts(0x30)...)
	}

	for i, r := range []int{8, 10, 12, 13, 14, 15} {
		w.p.Hit(song.At(3, r), chSnare, "C-4", k.snare, song.VolOpts(0x24+i*5)...)
	}

	return w.p
}

func drop(s *song.Song, k *kit, answer, final bool) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	phrase := hookA

	if answer {
		phrase = hookB
	}

	sec := section{
		padVol: 0x32,
		stabVol: 0x2C,
		arpStep: 1,
		arpVol: 0x2C,
		bassStyle: 2,
		bassVol: 0x3C,
		drumStyle: 3,
		leadVol: 0x44,
		phrase: phrase,
		fill: true,
	}

	if final {
		sec.padVol = 0x36
		sec.stabVol = 0x30
		sec.bassVol = 0x40
		sec.leadVol = 0x48
		sec.arpVol = 0x30
	}

	w.arrange(sec, progression)
	w.crash(0, 0x34)
	w.crash(2, 0x24)

	if final {
		for b, hits := range phrase {
			for _, h := range hits {
				w.p.Hit(song.At(b, h.row), chSpark, upOctave(h.note), k.arp, song.VolOpts(0x44, song.Pan(0xB0))...)
			}
		}
	}

	// an extra kick pushes the last beat of the pattern
	w.p.Hit(song.At(3, 14), chKick, "C-2", k.kick, song.VolOpts(0x30)...)

	return w.p
}

func groove(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x2A,
		stabVol: 0x24,
		arpStep: 4,
		arpVol: 0x24,
		bassStyle: 2,
		bassVol: 0x2C,
		drumStyle: 2,
		leadVol: 0x30,
		phrase: verseLead,
	}, progression)
	w.crash(0, 0x24)

	return w.p
}

func grooveFull(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x2E,
		stabVol: 0x28,
		arpStep: 2,
		arpVol: 0x28,
		bassStyle: 2,
		bassVol: 0x30,
		drumStyle: 3,
		leadVol: 0x34,
		phrase: verseLead,
		fill: true,
	}, progression)
	w.crash(0, 0x28)
	w.crash(2, 0x1C)

	return w.p
}

func introTwo(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(section{
		padVol: 0x26,
		stabVol: 0x1C,
		arpStep: 2,
		arpVol: 0x24,
		bassStyle: 1,
		bassVol: 0x28,
		drumStyle: 1,
		leadVol: 0x2C,
		phrase: introLead,
	}, progression)
	w.crash(0, 0x20)

	return w.p
}

func main() {
	s := song.New("low tide", 17)

	s.Timing(speed, tempo)

	k := &kit{
		lead: s.Instrument(song.Instrument{
			Name: "soft lead",
			Recipe: leadRecipe,
			Loop: true,
			Volume: 34,
			Pan: 96,
			Envelope: song.Points(0, 0, 2, 64),
			Fadeout: 512,
		}),
		arp: s.Instrument(song.Instrument{
			Name: "mallet",
			Recipe: arpRecipe,
			Volume: 28,
			Pan: 128,
			Envelope: song.Points(0, 64, 8, 0),
		}),
		stab: s.Instrument(song.Instrument{
			Name: "rhodes",
			Recipe: stabRecipe,
			Volume: 30,
			Pan: 128,
			Envelope: song.Points(0, 64, 8, 34, 18, 12, 28, 0),
			Fadeout: 512,
		}),
		bass: s.Instrument(song.Instrument{
			Name: "deep bass",
			Recipe: bassRecipe,
			Loop: true,
			Volume: 50,
			Pan: 128,
			Envelope: song.Points(0, 0, 2, 64),
		}),
		pad: s.Instrument(song.Instrument{
			Name: "wide pad",
			Recipe: padRecipe,
			Loop: true,
			Volume: 26,
			Pan: 128,
			Envelope: song.Points(0, 0, 8, 64),
			Fadeout: 1024,
		}),
		kick: s.Instrument(song.Instrument{
			Name: "kick",
			Recipe: kickRecipe,
			Volume: 58,
			Pan: 128,
			Envelope: song.Points(0, 64, 16, 0),
		}),
		snare: s.Instrument(song.Instrument{
			Name: "clap",
			Recipe: snareRecipe,
			Volume: 42,
			Pan: 136,
			Envelope: song.Points(0, 64, 14, 0),
		}),
		hat: s.Instrument(song.Instrument{
			Name: "closed hat",
			Recipe: hatRecipe,
			Volume: 28,
			Pan: 128,
			Envelope: song.Points(0, 64, 3, 0),
		}),
		openHat: s.Instrument(song.Instrument{
			Name: "open hat",
			Recipe: openHatRecipe,
			Volume: 24,
			Pan: 208,
			Envelope: song.Points(0, 64, 12, 0),
		}),
		shaker: s.Instrument(song.Instrument{
			Name: "shaker",
			Recipe: shakerRecipe,
			Volume: 22,
			Pan: 200,
			Envelope: song.Points(0, 64, 5, 0),
		}),
		crash: s.Instrument(song.Instrument{
			Name: "crash",
			Recipe: crashRecipe,
			Volume: 26,
			Pan: 128,
			Envelope: song.Points(0, 64, 64, 0),
		}),
		riser: s.Instrument(song.Instrument{
			Name: "riser",
			Recipe: riserRecipe,
			Volume: 20,
			Pan: 128,
			Envelope: song.Points(0, 0, 384, 64),
			Fadeout: 256,
		}),
	}
	introP := intro(s, k)
	introTwoP := introTwo(s, k)
	grooveP := groove(s, k)
	grooveFullP := grooveFull(s, k)
	breakP := breakdown(s, k)
	dropP := drop(s, k, false, false)
	answerP := drop(s, k, true, false)
	interludeP := interlude(s, k)
	buildP := build(s, k)
	finalP := drop(s, k, false, true)
	finalAnswerP := drop(s, k, true, true)
	outroP := outro(s, k)

	s.Play(introP, introTwoP, grooveP, grooveFullP, breakP, dropP, answerP, grooveFullP, interludeP, buildP, finalP, finalAnswerP, outroP)

	// write next to this file, wherever the program is run from
	if err := s.WriteBeside(); err != nil {
		log.Fatal(err)
	}
}
