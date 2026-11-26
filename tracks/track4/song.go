// Command track4 writes "pixel rush", a chiptune track, as an XM module, song.xm, in this folder.
//
// Build it from anywhere inside the repository with go run ./tracks/track4, or with tracker build.
package main

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/Heliodex/tracker/song"
)

type (
	// chordBar is one bar of a chord progression
	chordBar struct {
		chord string
		// bass is the root note of the bar, in the bass octave
		bass string
	}

	// hit is one note: a row within the bar, and the note name
	hit struct {
		row int
		note string
	}

	// kit holds the instruments, numbered as they are added to the song
	kit struct {
		lead, arp, bass, harm, mellow, kick, snare, hat, openHat, crash, tom *song.Instrument
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
	chArp
	chBass
	chHarm
	chKick
	chSnare
	chHat
	chOpen
	chFx
)

const (
	rows = 64
	bars = rows / song.RowsPerBar
	speed = 6
	// tempo gives a row of 0.1 seconds, a bar of 1.6 seconds
	tempo = 150

	// Sample recipes, all tuned to C-4 (261.34375 Hz), so 256 frames is a clean loop.
	// Amplitudes are kept low so nine channels sum without clipping.

	leadRecipe = `# thin pulse lead for the melody
Frames: 256
Bits: 16
Layer: square 261.34375 0.1725 duty=0.25
Layer: square 522.6875 0.05175 duty=0.25
Layer: sine 130.671875 0.05175
`
	arpRecipe = `# bright pulse arpeggio
Frames: 256
Bits: 16
Layer: square 261.34375 0.15525 duty=0.125
Layer: sine 522.6875 0.0345
`
	bassRecipe = `# triangle bass with a little body
Frames: 256
Bits: 16
Layer: triangle 261.34375 0.18975
Layer: sine 261.34375 0.1035
Layer: square 522.6875 0.02875
`
	harmRecipe = `# fat pulse for harmony lines and stabs
Frames: 256
Bits: 16
Layer: square 261.34375 0.1725 duty=0.5
Layer: square 522.6875 0.0414 duty=0.5
`
	mellowRecipe = `# soft triangle lead for the bridge
Frames: 256
Bits: 16
Layer: triangle 261.34375 0.18975
Layer: sine 522.6875 0.05175
`
	kickRecipe = `# low sine thump with a little click
Frames: 2000
Bits: 16
Layer: sine 55 0.345
Layer: sine 110 0.0805
Layer: noise 4000 0.023
`
	snareRecipe = `# noise snare with a little body
Frames: 3000
Bits: 16
Layer: noise 1500 0.2415
Layer: triangle 185 0.1035
`
	hatRecipe = `# short closed hat
Frames: 800
Bits: 16
Layer: noise 1 0.3105
Layer: square 4100 0.0345
`
	openHatRecipe = `# open hat, longer tail
Frames: 2000
Bits: 16
Layer: noise 1 0.3105
Layer: square 4100 0.0345
`
	crashRecipe = `# long bright noise crash
Frames: 8000
Bits: 16
Layer: noise 1 0.2415
Layer: noise 3000 0.09775
`
	tomRecipe = `# tom, a low sine with a short decay
Frames: 1200
Bits: 16
Layer: sine 110 0.2645
Layer: sine 165 0.0805
`
)

var (
	// Progressions: verse Em C G D, chorus G D Em C, bridge Am Em C D

	verseChords = []chordBar{
		{
			chord: "Em",
			bass: "E-2",
		},
		{
			chord: "C",
			bass: "C-2",
		},
		{
			chord: "G",
			bass: "G-2",
		},
		{
			chord: "D",
			bass: "D-2",
		},
	}
	chorusChords = []chordBar{
		{
			chord: "G",
			bass: "G-2",
		},
		{
			chord: "D",
			bass: "D-2",
		},
		{
			chord: "Em",
			bass: "E-2",
		},
		{
			chord: "C",
			bass: "C-2",
		},
	}
	outroChords = []chordBar{
		{
			chord: "Em",
			bass: "E-2",
		},
		{
			chord: "C",
			bass: "C-2",
		},
		{
			chord: "G",
			bass: "G-2",
		},
		{
			chord: "Em",
			bass: "E-2",
		},
	}
	bridgeChords = []chordBar{
		{
			chord: "Am",
			bass: "A-2",
		},
		{
			chord: "Em",
			bass: "E-2",
		},
		{
			chord: "C",
			bass: "C-2",
		},
		{
			chord: "D",
			bass: "D-2",
		},
	}

	// Melodies, one []hit per bar of the progression

	verse1Lead = [][]hit{
		{{0, "B-4"}, {6, "G-4"}, {8, "A-4"}, {12, "B-4"}},
		{{0, "C-5"}, {6, "B-4"}, {8, "A-4"}, {12, "G-4"}},
		{{0, "G-4"}, {4, "A-4"}, {8, "B-4"}, {12, "D-5"}},
		{{0, "A-4"}, {6, "F#4"}, {8, "A-4"}, {12, "D-5"}},
	}
	verse2Lead = [][]hit{
		{{0, "E-5"}, {2, "D-5"}, {4, "B-4"}, {8, "G-4"}, {10, "A-4"}, {12, "B-4"}},
		{{0, "C-5"}, {4, "E-5"}, {8, "D-5"}, {10, "C-5"}, {12, "B-4"}},
		{{0, "D-5"}, {4, "B-4"}, {6, "D-5"}, {8, "G-5"}, {12, "D-5"}},
		{{0, "F#5"}, {4, "E-5"}, {6, "D-5"}, {8, "A-4"}, {12, "D-5"}},
	}
	chorus1Lead = [][]hit{
		{{0, "B-4"}, {4, "D-5"}, {8, "G-5"}},
		{{0, "F#5"}, {6, "E-5"}, {8, "D-5"}},
		{{0, "E-5"}, {4, "D-5"}, {8, "B-4"}, {14, "D-5"}},
		{{0, "C-5"}, {6, "D-5"}, {8, "E-5"}},
	}
	chorus1Harm = [][]hit{
		{{0, "G-4"}, {4, "B-4"}, {8, "D-5"}},
		{{0, "D-5"}, {6, "A-4"}, {8, "A-4"}},
		{{0, "B-4"}, {4, "G-4"}, {8, "G-4"}, {14, "B-4"}},
		{{0, "E-4"}, {6, "G-4"}, {8, "C-5"}},
	}
	chorus2Lead = [][]hit{
		{{0, "B-4"}, {4, "D-5"}, {8, "G-5"}},
		{{0, "F#5"}, {6, "E-5"}, {8, "D-5"}},
		{{0, "G-5"}, {4, "F#5"}, {8, "E-5"}, {12, "D-5"}},
		{{0, "E-5"}, {4, "D-5"}, {8, "C-5"}, {14, "B-4"}},
	}
	chorus2Harm = [][]hit{
		{{0, "G-4"}, {4, "B-4"}, {8, "D-5"}},
		{{0, "D-5"}, {6, "A-4"}, {8, "A-4"}},
		{{0, "E-5"}, {4, "D-5"}, {8, "B-4"}, {12, "B-4"}},
		{{0, "C-5"}, {4, "G-4"}, {8, "E-4"}, {14, "G-4"}},
	}
	bridgeLead = [][]hit{
		{{0, "A-4"}, {8, "C-5"}},
		{{0, "B-4"}, {8, "G-4"}},
		{{0, "E-5"}, {8, "D-5"}},
		{{0, "A-4"}, {4, "B-4"}, {8, "C-5"}, {12, "D-5"}},
	}
	// outro lead, reusing the verse phrase then a final E
	outroLead = [][]hit{
		{{0, "B-4"}, {6, "G-4"}, {8, "A-4"}, {12, "B-4"}},
		{{0, "C-5"}, {6, "B-4"}, {8, "A-4"}, {12, "G-4"}},
		{{0, "G-4"}, {8, "B-4"}},
		{{0, "E-5"}},
	}
)

// at returns the row at offset r within bar b
func at(b, r int) int {
	return b*song.RowsPerBar + r
}

// fifthUp returns the fifth above the root note, such as E-2 to B-2
func fifthUp(root string) string {
	const tones string = "C-C#D-D#E-F-F#G-G#A-A#B-"

	for i := 0; i < len(tones); i += 2 {
		if tones[i:i+2] == root[:2] {
			if i+14 <= len(tones) {
				return tones[i+14:i+16] + root[2:]
			}

			// the fifth wraps into the next octave
			return tones[i+14-len(tones):i+16-len(tones)] + string(root[2]+1)
		}
	}

	return root
}

// octaveUp returns the note an octave higher, such as E-4 to E-5
func octaveUp(note string) string {
	if note[2] == '7' {
		return note
	}

	return note[:2] + string(note[2]+1)
}

// rangeStep returns from, from+step, ... below to
func rangeStep(from, to, step int) (out []int) {
	for r := from; r < to; r += step {
		out = append(out, r)
	}

	return
}

// triadOctave returns the triad for the chord, with the root in the given octave,
// plus the root an octave up, as the four notes of an arpeggio beat
func triadOctave(chord string, octave int) []string {
	t := song.Triad(chord, octave)

	return append(t, octaveUp(t[0]))
}

// withVol returns the options for a volume column value, which is left out when zero, followed by extra
func withVol(vol int, extra ...song.Opt) []song.Opt {
	var opts []song.Opt

	if vol != 0 {
		opts = append(opts, song.Vol(vol))
	}

	return append(opts, extra...)
}

// arp plays the running arpeggio for a bar: the triad up and down, one note per step rows
func (w pattern) arp(b int, cb chordBar, step, vol int) {
	// one beat of the cycle: root third fifth third, so the bar repeats four times
	notes := triadOctave(cb.chord, 4)
	cycle := []string{notes[0], notes[1], notes[2], notes[1]}

	for i, r := range rangeStep(0, song.RowsPerBar, step) {
		w.p.Hit(at(b, r), chArp, cycle[i%len(cycle)], w.k.arp, withVol(vol)...)
	}
}

// bass plays the bassline for a bar: the pattern of eighths named by the notes
func (w pattern) bass(b int, cb chordBar, notes []string, vol int) {
	for i, r := range rangeStep(0, song.RowsPerBar, 2) {
		w.p.Hit(at(b, r), chBass, notes[i], w.k.bass, withVol(vol)...)
	}
}

// chorusBass is the driving chorus pattern: alternating root and octave
func chorusBass(cb chordBar) []string {
	oct := octaveUp(cb.bass)

	return []string{cb.bass, oct, cb.bass, oct, cb.bass, fifthUp(cb.bass), oct, cb.bass}
}

// crash plays a crash on the fx channel
func (w pattern) crash(b, vol int) {
	w.p.Hit(at(b, 0), chFx, "C-4", w.k.crash, withVol(vol)...)
}

// drums plays the kick and snare for a bar: kick on the beats, snare on the backbeat
func (w pattern) drums(b int, fill bool) {
	w.p.Hit(at(b, 0), chKick, "C-2", w.k.kick)
	w.p.Hit(at(b, 8), chKick, "C-2", w.k.kick)
	w.p.Hit(at(b, 4), chSnare, "C-4", w.k.snare, withVol(0x40)...)
	w.p.Hit(at(b, 12), chSnare, "C-4", w.k.snare, withVol(0x40)...)

	if fill {
		// a 16th pickup into the next bar
		w.p.Hit(at(b, 14), chKick, "C-2", w.k.kick, withVol(0x38)...)
	}
}

// hats plays closed hats, one per step rows, accenting the beats and alternating panning
func (w pattern) hats(b, step, baseVol, accentVol int) {
	for i, r := range rangeStep(0, song.RowsPerBar, step) {
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

// intro opens with the arpeggio alone, hats joining halfway
func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.crash(0, 0x30)

	for b, cb := range verseChords {
		w.arp(b, cb, 1, 0x2C)

		if b >= 2 {
			w.hats(b, 2, 0x20, 0x28)
		}
	}

	// a snare pickup into the next pattern
	for i, vol := range []int{0x20, 0x28, 0x30, 0x38} {
		w.p.Hit(at(3, 12+i), chSnare, "C-4", k.snare, withVol(vol)...)
	}

	return w.p
}

// openHats plays open hats on the offbeat eighths, alternating panning
func (w pattern) openHats(b int) {
	for i, r := range []int{2, 6, 10, 14} {
		pan := 0xD0

		if i%2 == 0 {
			pan = 0x30
		}

		w.p.Hit(at(b, r), chOpen, "C-5", w.k.openHat, withVol(0x2C, song.Pan(pan))...)
	}
}

// play puts a list of hits on a channel, one per row within the bar they name
func (w pattern) play(ch, b int, hits []hit, ins *song.Instrument, vol int, extra ...song.Opt) {
	for _, h := range hits {
		w.p.Hit(at(b, h.row), ch, h.note, ins, withVol(vol, extra...)...)
	}
}

// bridge drops to a half-time feel with a soft lead line
func bridge(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.crash(0, 0x30)

	// the chorus harmony rings no further
	w.p.Hit(0, chHarm, "###", nil)

	for b, cb := range bridgeChords {
		w.arp(b, cb, 2, 0x2C)
		w.p.Hit(at(b, 0), chBass, cb.bass, k.bass, withVol(0x34)...)
		w.p.Hit(at(b, 8), chBass, fifthUp(cb.bass), k.bass, withVol(0x34)...)
		w.play(chLead, b, bridgeLead[b], k.mellow, 0x34)

		// half-time drums: kick on one, snare on three, hats joining halfway
		w.p.Hit(at(b, 0), chKick, "C-2", k.kick)

		if b >= 1 {
			w.p.Hit(at(b, 8), chSnare, "C-4", k.snare, withVol(0x38)...)
		}

		if b >= 2 {
			w.hats(b, 2, 0x20, 0x28)
		}
	}

	// a pickup into the build
	for i, vol := range []int{0x28, 0x30} {
		w.p.Hit(at(3, 14+i), chSnare, "C-4", k.snare, withVol(vol)...)
	}

	return w.p
}

// stabs plays offbeat chord-tone stabs on the harm channel
func (w pattern) stabs(b int, cb chordBar, vol int) {
	// the fifth of the chord in octave 4: E-2 gives B-4, C-2 gives G-4
	fifth := fifthUp(cb.bass[:1] + "-4")

	for _, r := range []int{2, 6, 10, 14} {
		w.p.Hit(at(b, r), chHarm, fifth, w.k.harm, withVol(vol)...)
	}
}

// tomFill plays a four-note tom run down, starting at row start of bar b
func (w pattern) tomFill(b, start int) {
	for i, note := range []string{"C-5", "A-4", "G-4", "E-4"} {
		w.p.Hit(at(b, start+i), chFx, note, w.k.tom, withVol(0x30+i*4)...)
	}
}

// chorus plays the anthemic melody with harmony beneath it
func chorus(s *song.Song, k *kit, lead, harm [][]hit, big bool) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	volLead, volHarm, volArp, bassVol := 0x44, 0x30, 0x30, 0x48

	if big {
		volLead, volHarm, volArp, bassVol = 0x4C, 0x38, 0x34, 0x44
	}

	w.crash(0, 0x40)
	w.crash(2, 0x1C)

	for b, cb := range chorusChords {
		w.arp(b, cb, 1, volArp)
		w.bass(b, cb, chorusBass(cb), bassVol)
		w.hats(b, 1, 0x20, 0x30)
		w.openHats(b)
		w.drums(b, b == 3)
		w.play(chLead, b, lead[b], k.lead, volLead)
		w.play(chHarm, b, harm[b], k.harm, volHarm)
	}

	if big {
		w.tomFill(3, 12)
	}

	return w.p
}

// verseBass is the bouncy verse pattern: root root octave root root fifth octave root
func verseBass(cb chordBar) []string {
	oct := octaveUp(cb.bass)
	fifth := fifthUp(cb.bass)

	return []string{cb.bass, cb.bass, oct, cb.bass, cb.bass, fifth, oct, cb.bass}
}

// build ramps up: full drums, then a rising scale over a snare roll with the bass dropped out
func build(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	// the bridge's lead rings no further
	w.p.Hit(0, chLead, "###", nil)

	scale := []string{"D-4", "E-4", "F#4", "G-4", "A-4", "B-4", "C-5", "D-5"}

	for b, cb := range verseChords {
		w.arp(b, cb, 1, 0x30)
		w.hats(b, 1, 0x24, 0x30)

		switch {
		case b < 2:
			w.bass(b, cb, verseBass(cb), 0x44)
			w.drums(b, false)
		case b == 2:
			// eighths of the snare build, over a 16th bass
			w.p.Hit(at(b, 0), chKick, "C-2", k.kick)
			w.p.Hit(at(b, 8), chKick, "C-2", k.kick)

			for i, r := range rangeStep(0, song.RowsPerBar, 2) {
				w.p.Hit(at(b, r), chSnare, "C-4", k.snare, withVol(0x28+i*2)...)
			}

			for _, r := range rangeStep(0, song.RowsPerBar, 1) {
				w.p.Hit(at(b, r), chBass, cb.bass, k.bass, withVol(0x40)...)
			}
		}
	}

	// the rising scale into the drop
	for i, note := range scale {
		w.p.Hit(at(2, i*2), chLead, note, k.lead, withVol(0x38+i*2)...)
	}

	for i, note := range []string{"E-5", "F#5", "G-5"} {
		w.p.Hit(at(3, i*4), chLead, note, k.lead, withVol(0x44+i*4)...)
	}

	// the last bar: a snare roll with no bass, then a tom fill
	w.p.Hit(at(3, 0), chBass, "###", nil)

	for i, r := range rangeStep(0, 12, 1) {
		w.p.Hit(at(3, r), chSnare, "C-4", k.snare, withVol(0x28+i*3)...)
	}

	w.p.Hit(at(3, 0), chKick, "C-2", k.kick)
	w.tomFill(3, 12)

	// the arpeggio swells through the roll
	for i, r := range rangeStep(0, song.RowsPerBar, 1) {
		w.p.Hit(at(3, r), chArp, triadOctave(verseChords[3].chord, 4)[i%4], k.arp, withVol(0x20+i*3)...)
	}

	return w.p
}

// intro2 brings in the bass and kick, then the full backbeat
func intro2(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.arp(b, cb, 1, 0x30)
		w.bass(b, cb, verseBass(cb), 0x38)

		if b >= 1 {
			w.hats(b, 2, 0x24, 0x2C)
		}

		if b >= 2 {
			w.drums(b, b == 3)
		}
	}

	return w.p
}

// outro thins out and ends on a held E
func outro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.crash(0, 0x38)

	// the chorus harmony rings no further
	w.p.Hit(0, chHarm, "###", nil)

	for b, cb := range outroChords {
		w.play(chLead, b, outroLead[b], k.lead, 0x3C)

		switch b {
		case 0, 1:
			w.arp(b, cb, 1, 0x2C)
			w.bass(b, cb, verseBass(cb), 0x40)
			w.hats(b, 2, 0x20, 0x28)
			w.drums(b, false)
		case 2:
			w.arp(b, cb, 1, 0x2C)
			w.p.Hit(at(b, 0), chBass, cb.bass, k.bass, withVol(0x38)...)
			w.hats(b, 2, 0x1C, 0x24)
			w.p.Hit(at(b, 0), chKick, "C-2", k.kick)
		case 3:
			// the final hit, ringing out
			w.arp(b, cb, 2, 0x28)
			w.p.Hit(at(b, 0), chHarm, "B-4", k.harm, withVol(0x2C)...)
			w.p.Hit(at(b, 0), chBass, cb.bass, k.bass, withVol(0x3C)...)
			w.p.Hit(at(b, 0), chKick, "C-2", k.kick, withVol(0x40)...)
			w.p.Hit(at(b, 0), chSnare, "C-4", k.snare, withVol(0x34)...)
			w.crash(b, 0x24)

			for _, ch := range []int{chLead, chArp, chBass, chHarm} {
				w.p.Hit(rows-2, ch, "###", nil)
			}
		}
	}

	return w.p
}

// verse is the first verse: melody over the full groove
func verse1(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.crash(0, 0x30)

	for b, cb := range verseChords {
		w.arp(b, cb, 1, 0x28)
		w.bass(b, cb, verseBass(cb), 0x40)
		w.hats(b, 2, 0x24, 0x30)
		w.stabs(b, cb, 0x24)
		w.drums(b, b == 3)
		w.play(chLead, b, verse1Lead[b], k.lead, 0x3C)
	}

	return w.p
}

// verse2 continues with a busier melody and 16th hats
func verse2(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.arp(b, cb, 1, 0x2C)
		w.bass(b, cb, verseBass(cb), 0x44)
		w.hats(b, 1, 0x1C, 0x2C)
		w.stabs(b, cb, 0x28)
		w.drums(b, b == 3)
		w.play(chLead, b, verse2Lead[b], k.lead, 0x40)
	}

	return w.p
}

func main() {
	s := song.New("pixel rush", 9)

	s.Timing(speed, tempo)

	k := &kit{
		lead: s.Instrument(song.Instrument{
			Name: "pulse lead",
			Recipe: leadRecipe,
			Loop: true,
			Volume: 48,
			Pan: 148,
		}),
		arp: s.Instrument(song.Instrument{
			Name: "arp",
			Recipe: arpRecipe,
			Loop: true,
			Volume: 40,
			Pan: 96,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 5,
					Y: 18,
				},
			},
		}),
		bass: s.Instrument(song.Instrument{
			Name: "triangle bass",
			Recipe: bassRecipe,
			Loop: true,
			Volume: 60,
			Pan: 128,
		}),
		harm: s.Instrument(song.Instrument{
			Name: "pulse harm",
			Recipe: harmRecipe,
			Loop: true,
			Volume: 44,
			Pan: 110,
		}),
		mellow: s.Instrument(song.Instrument{
			Name: "mellow lead",
			Recipe: mellowRecipe,
			Loop: true,
			Volume: 44,
			Pan: 120,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 4,
					Y: 56,
				},
			},
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
					X: 10,
					Y: 0,
				},
			},
		}),
		snare: s.Instrument(song.Instrument{
			Name: "snare",
			Recipe: snareRecipe,
			Volume: 52,
			Pan: 132,
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
			Volume: 36,
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
			Volume: 32,
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
		crash: s.Instrument(song.Instrument{
			Name: "crash",
			Recipe: crashRecipe,
			Volume: 44,
			Pan: 128,
			Envelope: []song.Point{
				{
					X: 0,
					Y: 64,
				},
				{
					X: 56,
					Y: 0,
				},
			},
		}),
		tom: s.Instrument(song.Instrument{
			Name: "tom",
			Recipe: tomRecipe,
			Volume: 48,
			Pan: 140,
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
	}
	introP := intro(s, k)
	intro2P := intro2(s, k)
	verse1P := verse1(s, k)
	verse2P := verse2(s, k)
	chorus1P := chorus(s, k, chorus1Lead, chorus1Harm, false)
	chorus2P := chorus(s, k, chorus2Lead, chorus2Harm, false)
	bridgeP := bridge(s, k)
	buildP := build(s, k)
	finalP := chorus(s, k, chorus1Lead, chorus1Harm, true)
	outroP := outro(s, k)

	s.Play(introP, intro2P, verse1P, verse2P, chorus1P, chorus2P, bridgeP, buildP, finalP, chorus2P, outroP)

	// write next to this file, wherever the program is run from
	_, src, _, _ := runtime.Caller(0)

	if err := s.Write(filepath.Join(filepath.Dir(src), "song.xm")); err != nil {
		log.Fatal(err)
	}
}
