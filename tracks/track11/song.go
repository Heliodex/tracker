// Command track11 writes "Glass Orchard", a warm broken-beat song, as an XM module.
//
// Build it from anywhere inside the repository with go run ./tracks/track11, or with tracker build.
package main

import (
	"log"

	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/song"
)

type (
	chord struct {
		bass, fifth string
		pad [3]string
		arp [8]string
	}

	hit struct {
		row int
		note string
	}

	kit struct {
		lead, pluck, bass, kick, snare, hat, openHat, pad, bell, crash *song.Instrument
	}
)

const (
	chLead = iota
	chPluck
	chBass
	chKick
	chSnare
	chHat
	chOpenHat
	chPadLow
	chPadMid
	chPadHigh
	chBell
)

const rows = 64

var (
	progression = []chord{
		{
			bass: "D-2",
			fifth: "A-2",
			pad: [3]string{"D-3", "F-3", "A-3"},
			arp: [8]string{"D-4", "A-4", "C-5", "F-5", "A-4", "F-4", "C-5", "A-4"},
		},
		{
			bass: "A#1",
			fifth: "F-2",
			pad: [3]string{"A#2", "D-3", "F-3"},
			arp: [8]string{"A#3", "F-4", "A-4", "D-5", "F-4", "D-4", "A-4", "F-4"},
		},
		{
			bass: "F-2",
			fifth: "C-3",
			pad: [3]string{"F-3", "A-3", "C-4"},
			arp: [8]string{"F-3", "C-4", "G-4", "A-4", "C-5", "A-4", "G-4", "C-4"},
		},
		{
			bass: "C-2",
			fifth: "G-2",
			pad: [3]string{"C-3", "E-3", "G-3"},
			arp: [8]string{"C-4", "G-4", "D-5", "E-5", "G-4", "E-4", "D-5", "G-4"},
		},
	}

	verse = [][]hit{
		{{0, "A-4"}, {6, "F-4"}, {9, "D-5"}, {14, "C-5"}},
		{{0, "A-4"}, {4, "F-4"}, {10, "D-5"}, {13, "F-5"}},
		{{2, "C-5"}, {6, "A-4"}, {11, "G-4"}, {14, "A-4"}},
		{{0, "G-4"}, {5, "E-4"}, {9, "D-5"}, {14, "C-5"}},
	}
	hook = [][]hit{
		{{0, "A-4"}, {3, "D-5"}, {6, "F-5"}, {10, "A-5"}, {14, "F-5"}},
		{{0, "F-5"}, {4, "D-5"}, {7, "A-4"}, {10, "D-5"}, {14, "F-5"}},
		{{0, "A-4"}, {4, "C-5"}, {7, "F-5"}, {11, "G-5"}, {14, "F-5"}},
		{{0, "E-5"}, {3, "D-5"}, {6, "C-5"}, {10, "G-4"}, {14, "A-4"}},
	}
	hookAnswer = [][]hit{
		{{0, "D-5"}, {4, "F-5"}, {8, "A-5"}, {12, "F-5"}},
		{{0, "A-4"}, {4, "D-5"}, {8, "F-5"}, {12, "A-5"}},
		{{0, "C-5"}, {4, "A-4"}, {8, "G-4"}, {12, "A-4"}},
		{{0, "G-4"}, {4, "C-5"}, {8, "E-5"}, {12, "D-5"}},
	}
)

func makeKit(s *song.Song) *kit {
	return &kit{
		lead: s.Instrument(song.Instrument{
			Name: "glass lead",
			Recipe: recipe.New(256).Triangle(recipe.Note(256, "C-4"), 0.28).Sine(recipe.Note(256, "C-4"), 0.12).Sine(recipe.Note(256, "C-5"), 0.08).String(),
			Loop: true,
			Volume: 34,
			Pan: 116,
			Envelope: song.Points(0, 0, 2, 64),
			Fadeout: 384,
		}),
		pluck: s.Instrument(song.Instrument{
			Name: "wood pluck",
			Recipe: recipe.New(1200).Triangle(recipe.Note(1200, "C-4"), 0.25).Sine(recipe.Note(1200, "C-5"), 0.1).Sine(recipe.Note(1200, "G-5"), 0.04).String(),
			Volume: 28,
			Pan: 146,
			Envelope: song.Points(0, 64, 5, 28, 16, 0),
		}),
		bass: s.Instrument(song.Instrument{
			Name: "round bass",
			Recipe: recipe.New(256).Sine(recipe.Note(256, "C-4"), 0.42).Triangle(recipe.Note(256, "C-3"), 0.16).String(),
			Loop: true,
			Volume: 42,
			Pan: 128,
			Envelope: song.Points(0, 0, 2, 64),
		}),
		kick: s.Instrument(song.Instrument{
			Name: "soft kick",
			Recipe: recipe.New(2200).Sine(55, 0.72).Sine(110, 0.18).Noise(1800, 0.025).String(),
			Volume: 52,
			Pan: 128,
			Envelope: song.Points(0, 64, 18, 0),
		}),
		snare: s.Instrument(song.Instrument{
			Name: "brush snare",
			Recipe: recipe.New(2800).Noise(1200, 0.42).Noise(3200, 0.12).Triangle(190, 0.18).String(),
			Volume: 38,
			Pan: 138,
			Envelope: song.Points(0, 64, 16, 0),
		}),
		hat: s.Instrument(song.Instrument{
			Name: "shimmer hat",
			Recipe: recipe.New(500).Noise(3600, 0.34).String(),
			Volume: 24,
			Pan: 176,
			Envelope: song.Points(0, 64, 4, 0),
		}),
		openHat: s.Instrument(song.Instrument{
			Name: "open shimmer",
			Recipe: recipe.New(1500).Noise(3000, 0.28).Noise(4000, 0.08).String(),
			Volume: 22,
			Pan: 76,
			Envelope: song.Points(0, 64, 12, 0),
		}),
		pad: s.Instrument(song.Instrument{
			Name: "orchard pad",
			Recipe: recipe.New(512).Triangle(recipe.Note(512, "C-4"), 0.23).Sine(recipe.Note(512, "C-3"), 0.18).Saw(recipe.Note(512, "C-4"), 0.06).String(),
			Loop: true,
			Volume: 27,
			Pan: 128,
			Envelope: song.Points(0, 0, 10, 64),
			Fadeout: 900,
		}),
		bell: s.Instrument(song.Instrument{
			Name: "dew bell",
			Recipe: recipe.New(1800).Sine(recipe.Note(1800, "C-4"), 0.3).Sine(recipe.Note(1800, "C-5"), 0.13).Sine(recipe.Note(1800, "G-5"), 0.07).String(),
			Volume: 26,
			Pan: 202,
			Envelope: song.Points(0, 64, 8, 24, 28, 0),
		}),
		crash: s.Instrument(song.Instrument{
			Name: "leaf wash",
			Recipe: recipe.New(6000).Noise(1300, 0.34).Noise(4000, 0.18).String(),
			Volume: 20,
			Pan: 128,
			Envelope: song.Points(0, 64, 48, 0),
		}),
	}
}

func addPads(p *song.Pattern, k *kit, bar int, c chord, volume int) {
	channels := [3]int{chPadLow, chPadMid, chPadHigh}
	pans := [3]int{0x58, 0x80, 0xA8}

	for i, note := range c.pad {
		p.Hit(song.At(bar, 0), channels[i], note, k.pad, song.Vol(volume), song.Pan(pans[i]))
	}
}

func addDrums(p *song.Pattern, k *kit, bar, level int) {
	base := song.At(bar, 0)
	p.Hit(base, chKick, "C-2", k.kick, song.Vol(0x38+level*2))
	p.Hit(base+7, chKick, "C-2", k.kick, song.Vol(0x26+level))
	p.Hit(base+10, chKick, "C-2", k.kick, song.Vol(0x30+level))
	p.Hit(base+4, chSnare, "C-4", k.snare, song.Vol(0x30+level*2))
	p.Hit(base+12, chSnare, "C-4", k.snare, song.Vol(0x34+level*2))

	for _, row := range []int{2, 6, 10, 14} {
		volume := 0x22 + level
		if row == 6 || row == 14 {
			volume += 4
		}

		p.Hit(base+row, chHat, "C-5", k.hat, song.Vol(volume))
	}

	if level >= 2 {
		p.Hit(base+6, chOpenHat, "C-5", k.openHat, song.Vol(0x24+level))
		p.Hit(base+14, chOpenHat, "C-5", k.openHat, song.Vol(0x28+level))
	}

	if level >= 3 && bar%2 == 1 {
		p.Hit(base+15, chSnare, "C-4", k.snare, song.Vol(0x20+level))
	}
}

func buildPattern(s *song.Song, k *kit, level, rotation int, melody [][]hit, answer bool) *song.Pattern {
	p := s.Pattern(rows)
	for bar := 0; bar < 4; bar++ {
		c := progression[(bar+rotation)%len(progression)]
		padVol := []int{0x1C, 0x24, 0x28, 0x2C, 0x30}[level]
		addPads(p, k, bar, c, padVol)

		if level >= 1 {
			p.Hit(song.At(bar, 0), chBass, c.bass, k.bass, song.Vol(0x30+level*2))
			if level >= 2 {
				p.Hit(song.At(bar, 8), chBass, c.fifth, k.bass, song.Vol(0x2C+level*2))
			}
		}

		if level >= 2 {
			step := 2
			count := len(c.arp)
			if level == 2 {
				step = 4
				count = 4
			}

			for i := 0; i < count; i++ {
				p.Hit(song.At(bar, i*step), chPluck, c.arp[i*len(c.arp)/count], k.pluck, song.Vol(0x24+level*2))
			}
		}

		if melody != nil {
			phrase := melody[(bar+rotation)%len(melody)]
			for _, note := range phrase {
				p.Hit(song.At(bar, note.row), chLead, note.note, k.lead, song.Vol(0x30+level*3))
				if answer && level >= 3 && note.row%2 == 0 {
					p.Hit(song.At(bar, note.row), chBell, note.note, k.bell, song.Vol(0x24+level))
				}
			}
		}

		if level >= 1 {
			addDrums(p, k, bar, level)
		}

		if level == 0 && bar%2 == 0 {
			p.Hit(song.At(bar, 12), chBell, c.arp[3], k.bell, song.Vol(0x28))
		}

		if level >= 2 && bar == 0 {
			p.Hit(song.At(bar, 0), chBell, "C-5", k.crash, song.Vol(0x2C+level))
		}
	}

	return p
}

func main() {
	s := song.New("Glass Orchard", 11)
	s.Timing(6, 104)
	k := makeKit(s)

	intro := buildPattern(s, k, 0, 0, nil, false)
	verseA := buildPattern(s, k, 1, 0, verse, false)
	verseB := buildPattern(s, k, 2, 1, verse, false)
	lift := buildPattern(s, k, 2, 2, hookAnswer, false)
	chorusA := buildPattern(s, k, 3, 0, hook, false)
	chorusB := buildPattern(s, k, 3, 1, hookAnswer, false)
	interlude := buildPattern(s, k, 1, 1, nil, false)
	bridge := buildPattern(s, k, 2, 3, hookAnswer, false)
	finalA := buildPattern(s, k, 4, 0, hook, true)
	finalB := buildPattern(s, k, 4, 1, hookAnswer, true)
	outro := buildPattern(s, k, 0, 2, nil, false)

	s.Play(intro, verseA, verseB, lift, chorusA, chorusB, interlude, verseA, lift, chorusA, chorusB, bridge, finalA, finalB, outro)

	if err := s.WriteBeside(); err != nil {
		log.Fatal(err)
	}
}
