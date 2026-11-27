// Command track13 writes "voyage to fjords", an XM module, song.xm, in this folder.
//
// Run it from anywhere inside the repository with go run ./tracks/track13, or with tracker build.
// The samples come from Voyage to Eternity by Big Jim and Winds of Fjords by minomus.
package main

import (
	"embed"
	"log"

	"github.com/Heliodex/tracker/song"
)

// rows is the length of every pattern: four bars of sixteen rows
const rows = 64

// channels, in the order they appear in the tracker
const (
	chArp = iota
	chShimmer
	chLead
	chHarmony
	chPadLo
	chPadMid
	chPadHi
	chBass
	chKick
	chSnare
	chHat
	chFX
)

//go:embed samples
var samples embed.FS

// sample returns a WAV file from the samples folder
func sample(name string) []byte {
	b, err := samples.ReadFile("samples/" + name)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

type (
	// chord is one bar of harmony: a pad triad, an eight-note arpeggio and a bass note with its fifth
	chord struct {
		pad [3]string
		arp [8]string
		bass, fifth string
		minor bool
	}

	// kit holds the instruments, in the order they are added to the song
	kit struct {
		arp, shimmer, lead, harmony, pad, bass *song.Instrument
		kick, snare, hat *song.Instrument
		crash, riser, bell, brass, chordMin, chordMaj *song.Instrument
		bigjim, noise *song.Instrument
	}

	// pattern wraps a song pattern with the instruments to play it with
	pattern struct {
		p *song.Pattern
		k *kit
	}
)

var (
	dm = chord{
		pad: [3]string{"D-4", "F-4", "A-4"},
		arp: [8]string{"D-5", "F-5", "A-5", "D-6", "A-5", "F-5", "D-5", "A-4"},
		bass: "D-4", fifth: "A-3", minor: true,
	}
	bb = chord{
		pad: [3]string{"A#3", "D-4", "F-4"},
		arp: [8]string{"A#4", "D-5", "F-5", "A#5", "F-5", "D-5", "A#4", "F-4"},
		bass: "A#3", fifth: "F-3",
	}
	fMaj = chord{
		pad: [3]string{"F-3", "A-3", "C-4"},
		arp: [8]string{"F-4", "A-4", "C-5", "F-5", "C-5", "A-4", "F-4", "C-4"},
		bass: "F-3", fifth: "C-4",
	}
	cMaj = chord{
		pad: [3]string{"C-4", "E-4", "G-4"},
		arp: [8]string{"C-5", "E-5", "G-5", "C-6", "G-5", "E-5", "C-5", "G-4"},
		bass: "C-4", fifth: "G-3",
	}
	gm = chord{
		pad: [3]string{"G-3", "A#3", "D-4"},
		arp: [8]string{"G-4", "A#4", "D-5", "G-5", "D-5", "A#4", "G-4", "D-4"},
		bass: "G-3", fifth: "D-4", minor: true,
	}
	aMaj = chord{
		pad: [3]string{"A-3", "C#4", "E-4"},
		arp: [8]string{"A-4", "C#5", "E-5", "A-5", "E-5", "C#5", "A-4", "E-4"},
		bass: "A-3", fifth: "E-4",
	}

	// the four-bar progressions, one chord per bar
	verseChords = []chord{dm, bb, fMaj, cMaj}
	turnChords = []chord{dm, aMaj, bb, fMaj}
	breakChords = []chord{gm, bb, dm, aMaj}

	// sparse verse melody, four notes per bar
	verseLead = [][]string{
		{"A-5", "F-5", "D-5", "F-5"},
		{"A#5", "D-5", "F-5", "D-5"},
		{"C-6", "A-5", "F-5", "A-5"},
		{"E-5", "C-5", "G-4", "C-5"},
	}

	// the chorus melody, eight notes per bar
	chorusLead = [][]string{
		{"D-6", "C-6", "A-5", "F-5", "A-5", "C-6", "D-6", "F-6"},
		{"D-6", "A#5", "F-5", "D-5", "F-5", "A#5", "D-6", "F-6"},
		{"C-6", "A-5", "G-5", "F-5", "A-5", "C-6", "F-6", "C-6"},
		{"G-5", "E-5", "C-5", "E-5", "G-5", "C-6", "G-5", "E-5"},
	}

	// a higher harmony that answers the lead in the second chorus
	harmonyLead = [][]string{
		{"A-6", "F-6", "D-6", "F-6", "A-6", "D-7", "C-7", "A-6"},
		{"F-6", "D-6", "A#5", "D-6", "F-6", "A#6", "D-7", "F-6"},
		{"F-6", "C-6", "A-5", "C-6", "F-6", "A-6", "C-7", "F-6"},
		{"E-6", "C-6", "G-5", "C-6", "E-6", "G-6", "C-7", "G-6"},
	}

	// the break's bell line
	bellLine = [][]string{
		{"G-5", "A#5", "D-6", "A#5"},
		{"F-5", "A#5", "D-6", "F-6"},
		{"F-5", "A-5", "D-6", "F-6"},
		{"E-5", "A-5", "C#6", "E-6"},
	}
)

// pad plays the triad of a chord across the three pad channels
func (w pattern) pad(b int, cb chord, vol int) {
	for i, note := range cb.pad {
		w.p.Hit(song.At(b, 0), chPadLo+i, note, w.k.pad,
			song.VolOpts(vol, song.Pan(0x40+i*0x40))...)
	}
}

// arp plays notes every step rows from the start of bar b
func (w pattern) arp(b int, notes []string, step, vol int, ins *song.Instrument, ch int) {
	w.p.Line(ch, song.At(b, 0), step, notes, ins, song.VolOpts(vol)...)
}

// bass plays the root on the beats, or the root and fifth on every other eighth
func (w pattern) bass(b int, cb chord, eighths bool, vol int) {
	w.p.Hit(song.At(b, 0), chBass, cb.bass, w.k.bass, song.VolOpts(vol)...)
	w.p.Hit(song.At(b, 8), chBass, cb.bass, w.k.bass, song.VolOpts(vol)...)

	if eighths {
		for i, note := range []string{cb.fifth, cb.fifth} {
			r := 4 + i*8
			w.p.Hit(song.At(b, r), chBass, note, w.k.bass, song.VolOpts(vol-0x08)...)
		}
	}
}

// melody plays a line of notes on one channel, one every step rows
func (w pattern) melody(b int, notes []string, step, vol int, ins *song.Instrument, ch int) {
	w.arp(b, notes, step, vol, ins, ch)
}

// drums lays down the kick, snare and hats for one bar; fill adds a snare roll
func (w pattern) drums(b, fill int) {
	if fill == 0 {
		return
	}

	for _, r := range []int{0, 8} {
		w.p.Hit(song.At(b, r), chKick, "C-5", w.k.kick)
	}

	if fill >= 2 {
		w.p.Hit(song.At(b, 6), chKick, "C#5", w.k.kick, song.Vol(0x38))
		w.p.Hit(song.At(b, 14), chKick, "A-4", w.k.kick, song.Vol(0x38))
	}

	for _, r := range []int{4, 12} {
		w.p.Hit(song.At(b, r), chSnare, "C-5", w.k.snare, song.Vol(0x40))
	}

	w.p.Hit(song.At(b, 15), chSnare, "C-5", w.k.snare, song.Vol(0x20))

	step := 4

	if fill >= 2 {
		step = 2
	}

	for i, r := range song.Steps(0, song.RowsPerBar, step) {
		pan := 0x90
		if i%2 == 0 {
			pan = 0x60
		}

		vol := 0x28
		if r%4 == 0 {
			vol = 0x38
		}

		w.p.Hit(song.At(b, r), chHat, "C-5", w.k.hat, song.VolOpts(vol, song.Pan(pan))...)
	}

	if fill == 3 {
		for i, vol := range []int{0x28, 0x30, 0x38, 0x44} {
			w.p.Hit(song.At(b, 12+i), chSnare, "C-5", w.k.snare, song.Vol(vol))
		}
	}
}

// stab plays a short sample on the effects channel
func (w pattern) stab(b, r int, note string, ins *song.Instrument, vol int) {
	w.p.Hit(song.At(b, r), chFX, note, ins, song.VolOpts(vol)...)
}

// crash hits the crash cymbal, quietly unless it is accented
func (w pattern) crash(b, r, vol int) {
	w.stab(b, r, "C-5", w.k.crash, vol)
}

// shimmer scatters airy hits across the stereo field
func (w pattern) shimmer(b int, rows []int) {
	for i, r := range rows {
		pan := 0x40
		if i%2 == 0 {
			pan = 0xC0
		}

		w.p.Hit(song.At(b, r), chShimmer, "C-6", w.k.shimmer,
			song.VolOpts(0x24, song.Pan(pan))...)
	}
}

func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.pad(b, cb, 0x24)
		w.arp(b, cb.arp[:4], 4, 0x2C, k.arp, chArp)

		if b >= 1 {
			w.shimmer(b, []int{0, 8})
		}
	}

	w.stab(1, 12, "D-6", k.bell, 0x28)
	w.stab(2, 4, "A-5", k.bell, 0x24)
	w.stab(3, 0, "C-5", k.riser, 0x28)
	w.crash(3, 8, 0x30)

	return w.p
}

// verse plays a four-bar verse; heavy adds eighths, busier drums and the lead
func verse(s *song.Song, k *kit, heavy bool) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	cs := verseChords

	if heavy {
		cs = turnChords
	}

	step := 2
	if heavy {
		step = 1
	}

	for b, cb := range cs {
		w.pad(b, cb, 0x24)
		notes := cb.arp[:]
		if heavy {
			notes = append(cb.arp[:], cb.arp[:]...)
		}

		w.arp(b, notes, step, 0x30, k.arp, chArp)
		w.bass(b, cb, heavy, 0x38)

		fill := 1
		if b == 3 {
			fill = 2
		}

		w.drums(b, fill)

		if heavy || b >= 2 {
			w.melody(b, verseLead[b%len(verseChords)], 4, 0x40, k.lead, chLead)
		}

		if heavy && b%2 == 1 {
			w.stab(b, 6, "C-6", k.bell, 0x2C)
		}
	}

	if heavy {
		w.drums(3, 3)
	}

	w.crash(0, 0, 0x34)

	return w.p
}

func chorus(s *song.Song, k *kit, harmony bool) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.pad(b, cb, 0x2C)
		w.arp(b, append(cb.arp[:], cb.arp[:]...), 1, 0x34, k.arp, chArp)
		w.bass(b, cb, true, 0x3C)
		w.melody(b, chorusLead[b], 2, 0x48, k.lead, chLead)

		if harmony {
			w.melody(b, harmonyLead[b], 2, 0x30, k.harmony, chHarmony)
		} else {
			w.shimmer(b, []int{0, 8})
		}

		w.drums(b, 2)
		w.crash(b, 0, 0x38)

		ins := k.chordMaj
		if cb.minor {
			ins = k.chordMin
		}

		w.stab(b, 8, cb.pad[0], ins, 0x28)

		if b%2 == 1 {
			w.stab(b, 10, "C-5", k.bigjim, 0x24)
			w.stab(b, 14, "C-6", k.noise, 0x20)
		}
	}

	w.drums(3, 3)

	return w.p
}

func breakdown(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range breakChords {
		w.pad(b, cb, 0x28)
		w.arp(b, cb.arp[:4], 4, 0x28, k.arp, chArp)
		w.melody(b, bellLine[b], 4, 0x30, k.bell, chFX)
		w.shimmer(b, []int{0, 4, 8, 12})
	}

	w.stab(0, 0, "C-5", k.riser, 0x24)
	w.crash(0, 8, 0x28)

	for i, vol := range []int{0x1C, 0x28, 0x34, 0x40} {
		w.p.Hit(song.At(3, 8+i*2), chSnare, "C-5", k.snare, song.Vol(vol))
	}

	return w.p
}

func finale(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords {
		w.pad(b, cb, 0x30)
		w.arp(b, append(cb.arp[:], cb.arp[:]...), 1, 0x3C, k.arp, chArp)
		w.bass(b, cb, true, 0x40)
		w.melody(b, chorusLead[b], 2, 0x50, k.lead, chLead)
		w.melody(b, harmonyLead[b], 2, 0x38, k.harmony, chHarmony)
		w.drums(b, 3)
		w.crash(b, 0, 0x40)
		w.stab(b, 8, cb.pad[0], k.brass, 0x30)
	}

	w.stab(3, 12, "C-5", k.riser, 0x24)

	return w.p
}

func outro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verseChords[:2] {
		w.pad(b, cb, 0x24)
		w.arp(b, cb.arp[:4], 4, 0x2C, k.arp, chArp)
		w.shimmer(b, []int{0, 8})

		if b == 0 {
			w.melody(b, verseLead[b], 4, 0x38, k.lead, chLead)
		}
	}

	// the last D minor chord rings out over a crash
	w.pad(2, dm, 0x2C)
	w.p.Hit(song.At(2, 0), chBass, "D-4", k.bass, song.Vol(0x30))
	w.melody(2, []string{"D-6", "A-5", "F-5", "D-5"}, 4, 0x40, k.lead, chLead)
	w.crash(2, 0, 0x3C)

	for _, r := range []int{0, 6, 12} {
		w.p.Hit(song.At(3, r), chArp, "D-6", k.arp, song.Vol(0x28))
	}

	w.pad(3, dm, 0x20)
	w.stab(3, 0, "C-5", k.riser, 0x1C)

	for _, ch := range []int{chLead, chBass} {
		w.p.Hit(song.At(3, 0), ch, "###", nil)
	}

	for _, ch := range []int{chPadLo, chPadMid, chPadHi, chShimmer, chFX} {
		w.p.Hit(rows-1, ch, "###", nil)
	}

	return w.p
}

func main() {
	s := song.New("voyage to fjords", 12)
	s.Timing(6, 116)
	k := &kit{
		arp: s.Instrument(song.Instrument{
			Name: "aharp",
			WAV: sample("arp.wav"),
			Volume: 44,
			Envelope: song.Points(0, 64, 12, 0),
		}),
		shimmer: s.Instrument(song.Instrument{
			Name: "fjords",
			WAV: sample("fjords.wav"),
			Volume: 36,
			Envelope: song.Points(0, 64, 10, 0),
		}),
		lead: s.Instrument(song.Instrument{
			Name: "1voice lead",
			WAV: sample("voice.wav"),
			Volume: 52,
			Envelope: song.Points(0, 64, 16, 0),
		}),
		harmony: s.Instrument(song.Instrument{
			Name: "domu lead",
			WAV: sample("domu.wav"),
			Volume: 44,
			Envelope: song.Points(0, 64, 16, 0),
		}),
		pad: s.Instrument(song.Instrument{
			Name: "mist pad",
			WAV: sample("mist.wav"),
			Loop: true,
			Volume: 36,
			Fadeout: 1024,
		}),
		bass: s.Instrument(song.Instrument{
			Name: "sylead bass",
			WAV: sample("sylead.wav"),
			Volume: 52,
			Envelope: song.Points(0, 64, 20, 0),
		}),
		kick: s.Instrument(song.Instrument{
			Name: "thump kick",
			WAV: sample("kick.wav"),
			Envelope: song.Points(0, 64, 9, 0),
		}),
		snare: s.Instrument(song.Instrument{
			Name: "dp snare",
			WAV: sample("snare.wav"),
			Volume: 56,
			Envelope: song.Points(0, 64, 8, 0),
		}),
		hat: s.Instrument(song.Instrument{
			Name: "hihat",
			WAV: sample("hat.wav"),
			Volume: 40,
			Envelope: song.Points(0, 64, 3, 0),
		}),
		crash: s.Instrument(song.Instrument{
			Name: "big crash",
			WAV: sample("crash.wav"),
			Volume: 40,
			Envelope: song.Points(0, 64, 40, 0),
		}),
		riser: s.Instrument(song.Instrument{
			Name: "reverse cym",
			WAV: sample("riser.wav"),
			Volume: 36,
			Envelope: song.Points(0, 64, 60, 0),
		}),
		bell: s.Instrument(song.Instrument{
			Name: "bell",
			WAV: sample("bell.wav"),
			Volume: 44,
			Envelope: song.Points(0, 64, 16, 0),
		}),
		brass: s.Instrument(song.Instrument{
			Name: "brass stab",
			WAV: sample("brass.wav"),
			Volume: 40,
			Envelope: song.Points(0, 64, 20, 0),
		}),
		chordMin: s.Instrument(song.Instrument{
			Name: "cos min",
			WAV: sample("cosmin.wav"),
			Volume: 40,
			Envelope: song.Points(0, 64, 20, 0),
		}),
		chordMaj: s.Instrument(song.Instrument{
			Name: "cos maj",
			WAV: sample("cosmaj.wav"),
			Volume: 40,
			Envelope: song.Points(0, 64, 24, 0),
		}),
		bigjim: s.Instrument(song.Instrument{
			Name: "big jim",
			WAV: sample("bigjim.wav"),
			Volume: 36,
			Envelope: song.Points(0, 64, 12, 0),
		}),
		noise: s.Instrument(song.Instrument{
			Name: "valhalla",
			WAV: sample("valhalla.wav"),
			Volume: 32,
			Envelope: song.Points(0, 64, 8, 0),
		}),
	}

	patterns := []*song.Pattern{
		intro(s, k),
		verse(s, k, false),
		verse(s, k, true),
		chorus(s, k, false),
		breakdown(s, k),
		verse(s, k, true),
		chorus(s, k, true),
		breakdown(s, k),
		finale(s, k),
		outro(s, k),
	}

	s.Play(patterns...)

	if err := s.WriteBeside(); err != nil {
		log.Fatal(err)
	}
}
