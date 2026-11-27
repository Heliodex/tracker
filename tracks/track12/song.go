// Command track12 writes "track12", an original song in B minor, as an XM module, song.xm, in this folder.
//
// Run it from anywhere inside the repository with go run ./tracks/track12, or with tracker build.
//
// The piece is a syncopated synth-pop track in B minor at 124 BPM: an intro, two verses, a chorus,
// a verse, a chorus, a bridge, a big chorus, a chorus, and a short outro resolving back to B minor.
package main

// track12 – Seaglass Alpha

import (
	"log"

	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/song"
	"github.com/Heliodex/tracker/types"
)

type (
	// chordBar is one bar of the progression
	chordBar struct {
		chord string // chord name for song.Triad, such as "Bm" or "F#"
		oct int // octave of the triad root for the pad
		bass string // bass note for the whole bar
		lead []string // four lead notes on rows 0, 4, 8, 12
	}

	// kit holds the instruments
	kit struct {
		lead, bass, kick, snare, hat, crash, pad, bell *song.Instrument
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
	// instrument recipes, all tuned to C-4 so note names pick the pitch

	// triangle lead with a soft sine octave, rounder than a square
	leadRecipe = recipe.New(256).Triangle(recipe.Note(256, "C-4"), 0.35).Sine(recipe.Note(256, "C-5"), 0.1).Sine(recipe.Note(256, "G-5"), 0.04).String()

	// square bass with a triangle body an octave down
	bassRecipe = recipe.New(256).Square(recipe.Note(256, "C-4"), 0.2).Duty(0.5).Triangle(recipe.Note(256, "C-3"), 0.4).String()

	// kick: a 220 Hz sine played at C-2 sounds at 55 Hz, with a triangle click on top
	kickRecipe = recipe.New(2000).Sine(220, 1).Triangle(440, 0.1).Noise(2500, 0.05).String()

	// snare: bright noise with a sine body
	snareRecipe = recipe.New(3000).Noise(1800, 0.8).Noise(4000, 0.35).Sine(240, 0.2).String()

	// short, bright hat
	hatRecipe = recipe.New(600).Noise(4000, 0.7).String()

	// long noise crash, darker than a hat
	crashRecipe = recipe.New(9000).Noise(4000, 0.5).Noise(1500, 0.3).String()

	// sine pad with a faint saw under it
	padRecipe = recipe.New(512).Sine(recipe.Note(512, "C-4"), 0.3).Triangle(recipe.Note(512, "C-5"), 0.06).Saw(recipe.Note(512, "C-3"), 0.05).String()

	// plucked triangle bell for the arpeggio
	bellRecipe = recipe.New(1024).Triangle(recipe.Note(1024, "C-4"), 0.4).Sine(recipe.Note(1024, "C-5"), 0.06).String()

	// bridge progression: Em A G F#
	bridgeChords = []chordBar{
		{
			chord: "Em",
			oct: 3,
			bass: "E-2",
			lead: []string{"B-4", "E-5", "G-5", "E-5"},
		},
		{
			chord: "A",
			oct: 3,
			bass: "A-2",
			lead: []string{"A-4", "C#5", "E-5", "A-5"},
		},
		{
			chord: "G",
			oct: 3,
			bass: "G-2",
			lead: []string{"B-4", "D-5", "G-5", "D-5"},
		},
		{
			chord: "F#",
			oct: 3,
			bass: "F#2",
			lead: []string{"C#5", "A#4", "F#4", "A#4"},
		},
	}

	// chorus progression: G D Em A
	chorusChords = []chordBar{
		{
			chord: "G",
			oct: 3,
			bass: "G-2",
			lead: []string{"D-5", "G-5", "B-5", "G-5"},
		},
		{
			chord: "D",
			oct: 3,
			bass: "D-2",
			lead: []string{"F#5", "D-5", "A-4", "D-5"},
		},
		{
			chord: "Em",
			oct: 3,
			bass: "E-2",
			lead: []string{"E-5", "G-5", "B-5", "G-5"},
		},
		{
			chord: "A",
			oct: 3,
			bass: "A-2",
			lead: []string{"C#5", "E-5", "A-5", "E-5"},
		},
	}

	// verse progression: Bm G D A
	verse = []chordBar{
		{
			chord: "Bm",
			oct: 3,
			bass: "B-2",
			lead: []string{"B-4", "D-5", "F#5", "D-5"},
		},
		{
			chord: "G",
			oct: 3,
			bass: "G-2",
			lead: []string{"G-4", "B-4", "D-5", "B-4"},
		},
		{
			chord: "D",
			oct: 3,
			bass: "D-2",
			lead: []string{"F#4", "A-4", "D-5", "A-4"},
		},
		{
			chord: "A",
			oct: 3,
			bass: "A-2",
			lead: []string{"A-4", "C#5", "E-5", "C#5"},
		},
	}

	// verseB progression, same chords with a different lead
	verseB = []chordBar{
		{
			chord: "Bm",
			oct: 3,
			bass: "B-2",
			lead: []string{"F#5", "D-5", "B-4", "D-5"},
		},
		{
			chord: "G",
			oct: 3,
			bass: "G-2",
			lead: []string{"D-5", "B-4", "G-4", "B-4"},
		},
		{
			chord: "D",
			oct: 3,
			bass: "D-2",
			lead: []string{"A-4", "F#4", "D-4", "F#4"},
		},
		{
			chord: "A",
			oct: 3,
			bass: "A-2",
			lead: []string{"E-5", "C#5", "A-4", "C#5"},
		},
	}
)

// bass plays the bar's bass note on the downbeat
func (w pattern) bass(b int, note string) {
	w.p.Hit(song.At(b, 0), chBass, note, w.k.bass)
}

// crash plays a crash on the hat channel at the bar's downbeat
func (w pattern) crash(b, vol int) {
	w.p.Hit(song.At(b, 0), chHat, "C-4", w.k.crash, song.VolOpts(vol)...)
}

// drums plays a syncopated kick on rows 0, 7 and 10, snare on 4 and 12; fill adds a snare roll at the bar's end
func (w pattern) drums(b int, fill bool) {
	w.p.Hit(song.At(b, 0), chDrum, "C-2", w.k.kick)
	w.p.Hit(song.At(b, 7), chDrum, "C-2", w.k.kick)
	w.p.Hit(song.At(b, 10), chDrum, "C-2", w.k.kick)
	w.p.Hit(song.At(b, 4), chDrum, "C-4", w.k.snare)
	w.p.Hit(song.At(b, 12), chDrum, "C-4", w.k.snare)

	if fill {
		for i, r := range []int{12, 13, 14, 15} {
			w.p.Hit(song.At(b, r), chDrum, "C-4", w.k.snare, song.VolOpts(0x30+i*4)...)
		}
	}
}

// hats plays closed hats from row start every step rows through the bar, so start 2 puts them on the offbeats
func (w pattern) hats(b, start, step, vol int) {
	for _, r := range song.Steps(start, song.RowsPerBar, step) {
		w.p.Hit(song.At(b, r), chHat, "C-5", w.k.hat, song.VolOpts(vol)...)
	}
}

// lead plays the bar's four-note lead on rows 0, 4, 8, 12
func (w pattern) lead(b int, notes []string, vol int) {
	w.p.Line(chLead, song.At(b, 0), 4, notes, w.k.lead, song.VolOpts(vol)...)
}

// off silences the sustained channels at the end of a pattern
func (w pattern) off() {
	for _, ch := range []int{chLead, chBass, chPadLo, chPadMid, chPadHi} {
		w.p.Hit(rows-1, ch, "###", nil)
	}
}

// pad plays the bar's triad across the three pad channels with panning
func (w pattern) pad(b int, cb chordBar, vol int) {
	pans := [3]int{0x60, 0x80, 0xA0}

	for i, note := range song.Triad(cb.chord, cb.oct) {
		w.p.Hit(song.At(b, 0), chPadLo+i, note, w.k.pad, song.VolOpts(vol, song.Pan(pans[i]))...)
	}
}

// shift returns note moved by semis semitones
func shift(note string, semis int) string {
	n, err := types.ParseNote(note)
	if err != nil {
		panic(err)
	}

	v := int(n) + semis
	if v < 1 || v > 96 {
		panic("note out of range")
	}

	return types.Note(v).String()
}

// arp8 returns eight eighth-note arpeggio steps from a triad: the root, then a dip and a leap up an octave,
// falling back through the top of the chord
func arp8(tri []string) []string {
	up := shift(tri[0], 12)

	return []string{tri[0], tri[2], tri[1], tri[2], up, tri[2], up, tri[1]}
}

// arp plays the bar's arpeggio, one note every step rows
func (w pattern) arp(b int, cb chordBar, step, vol int) {
	notes := arp8(song.Triad(cb.chord, cb.oct+1))

	w.p.Line(chArp, song.At(b, 0), step, notes, w.k.bell, song.VolOpts(vol)...)
}

// bridge: stripped back, no kick, lead on pads, building into the final chorus
func bridge(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range bridgeChords {
		w.pad(b, cb, 0x24)
		w.bass(b, cb.bass)
		w.arp(b, cb, 2, 0x24)
		w.lead(b, cb.lead, 0x40)
		w.hats(b, 2, 4, 0x20)
		w.p.Hit(song.At(b, 4), chDrum, "C-4", k.snare, song.VolOpts(0x28)...)
		w.p.Hit(song.At(b, 12), chDrum, "C-4", k.snare, song.VolOpts(0x28)...)
	}

	// snare build into the chorus
	for i, r := range []int{12, 13, 14, 15} {
		w.p.Hit(song.At(3, r), chDrum, "C-4", k.snare, song.VolOpts(0x30+i*4)...)
	}

	w.crash(0, 0x28)
	w.off()

	return w.p
}

// chorus lays out a chorus; big turns up the volume, thickens the arpeggio, and adds hat work
func chorus(s *song.Song, k *kit, big bool) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	arpVol := 0x30
	hatStep := 4
	leadVol := 0x48
	padVol := 0x28

	if big {
		arpVol = 0x38
		leadVol = 0x50
		padVol = 0x2C
		hatStep = 2
	}

	for b, cb := range chorusChords {
		w.pad(b, cb, padVol)
		w.bass(b, cb.bass)
		w.arp(b, cb, 2, arpVol)
		w.lead(b, cb.lead, leadVol)
		w.hats(b, 2, hatStep, 0x30)
		w.drums(b, big && b == len(chorusChords)-1)

		if b == 0 || b == 2 {
			w.crash(b, 0x38)
		}
	}

	w.off()

	return w.p
}

// intro: pads and arpeggio first, then hats, bass, and drums come in
func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range verse {
		w.pad(b, cb, 0x24)
		w.arp(b, cb, 2, 0x30)

		if b == 0 {
			w.crash(b, 0x28)
		}

		if b >= 1 {
			w.hats(b, 2, 4, 0x28)
		}

		if b >= 2 {
			w.bass(b, cb.bass)
		}

		if b == 3 {
			w.drums(b, false)
		}
	}

	w.off()

	return w.p
}

// outro: three bars resolving back to B minor, ending on a single note with a crash
func outro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}
	bars := []chordBar{
		{
			chord: "Bm",
			oct: 3,
			bass: "B-2",
			lead: []string{"B-4", "D-5", "F#5", "D-5"},
		},
		{
			chord: "A",
			oct: 3,
			bass: "A-2",
			lead: []string{"A-4", "C#5", "E-5", "C#5"},
		},
		{
			chord: "G",
			oct: 3,
			bass: "G-2",
			lead: []string{"G-4", "B-4", "D-5", "B-4"},
		},
	}

	for b, cb := range bars {
		w.pad(b, cb, 0x24)
		w.bass(b, cb.bass)
		w.arp(b, cb, 2, 0x24)
		w.lead(b, cb.lead, 0x40)
		w.hats(b, 2, 4, 0x24)
		w.drums(b, false)
	}

	// final B minor chord, no drums, note-offs to let it ring out
	bm := chordBar{
		chord: "Bm",
		oct: 3,
		bass: "B-2",
	}

	w.pad(3, bm, 0x28)
	w.bass(3, bm.bass)
	w.p.Hit(song.At(3, 0), chLead, "B-4", k.lead, song.Vol(0x50))
	w.p.Hit(song.At(3, 0), chDrum, "C-2", k.kick)
	w.crash(3, 0x48)

	for _, ch := range []int{chLead, chBass, chPadLo, chPadMid, chPadHi} {
		w.p.Hit(rows-1, ch, "###", nil)
	}

	return w.p
}

// versePattern lays out a full verse bar loop; arpStep and fill vary the repeats
func versePattern(s *song.Song, k *kit, bars []chordBar, arpStep, leadVol int, fill bool) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, cb := range bars {
		w.pad(b, cb, 0x24)
		w.bass(b, cb.bass)
		w.arp(b, cb, arpStep, 0x28)
		w.lead(b, cb.lead, leadVol)
		w.hats(b, 2, 4, 0x28)
		w.drums(b, fill && b == len(bars)-1)

		if b == 0 {
			w.crash(b, 0x28)
		}
	}

	w.off()

	return w.p
}

func main() {
	s := song.New("track12", 8)

	s.Timing(6, 124)

	k := &kit{
		lead: s.Instrument(song.Instrument{
			Name: "lead",
			Recipe: leadRecipe,
			Loop: true,
			Volume: 40,
			Pan: 160,
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
			Envelope: song.Points(0, 64, 10, 0),
		}),
		snare: s.Instrument(song.Instrument{
			Name: "snare",
			Recipe: snareRecipe,
			Volume: 48,
			Pan: 136,
			Envelope: song.Points(0, 64, 18, 0),
		}),
		hat: s.Instrument(song.Instrument{
			Name: "hat",
			Recipe: hatRecipe,
			Volume: 40,
			Pan: 120,
			Envelope: song.Points(0, 64, 3, 0),
		}),
		crash: s.Instrument(song.Instrument{
			Name: "crash",
			Recipe: crashRecipe,
			Volume: 40,
			Pan: 128,
			Envelope: song.Points(0, 64, 40, 0),
		}),
		pad: s.Instrument(song.Instrument{
			Name: "pad",
			Recipe: padRecipe,
			Loop: true,
			Volume: 36,
			Pan: 128,
			Envelope: song.Points(0, 0, 16, 64),
			Fadeout: 512,
		}),
		bell: s.Instrument(song.Instrument{
			Name: "bell",
			Recipe: bellRecipe,
			Volume: 40,
			Pan: 160,
			Envelope: song.Points(0, 64, 12, 0),
		}),
	}
	patterns := []*song.Pattern{
		intro(s, k),
		versePattern(s, k, verse, 2, 0x44, false),
		versePattern(s, k, verseB, 2, 0x48, true),
		chorus(s, k, false),
		bridge(s, k),
		chorus(s, k, true),
		outro(s, k),
	}
	order := []int{0, 1, 2, 3, 1, 2, 3, 4, 5, 3, 5, 6}

	for _, i := range order {
		s.Play(patterns[i])
	}

	// write next to this file, wherever the program is run from
	if err := s.WriteBeside(); err != nil {
		log.Fatal(err)
	}
}
