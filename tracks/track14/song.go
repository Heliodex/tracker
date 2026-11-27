// Command track14 writes "drowned waltz", a slow lo-fi waltz in F# minor, as an XM module, song.xm, in this folder.
//
// Build it from anywhere inside the repository with go run ./tracks/track14, or with tracker build.
package main

// drowned waltz – Muse Spark 1.3

import (
	"log"

	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/song"
)

type (
	// hit is one lead note: the bar and row within the bar, and the note name
	hit struct {
		bar, row int
		note string
	}

	// kit holds the instruments, numbered as they are added to the song
	kit struct {
		lead, pad, bass, kick, rim, shaker, vinyl *song.Instrument
	}

	// pattern wraps a song pattern with the instruments to play it with
	pattern struct {
		p *song.Pattern
		k *kit
	}

	// section describes the layers of an eight-bar pattern. A nil phrase plays no lead
	section struct {
		padVol, bassVol, drumStyle, leadVol int
		phrase []hit
		fill bool
	}
)

// channels, in the order they appear in the tracker
const (
	chLead = iota // the music box
	chPad0        // three pad voices, chPad0 to chPad0+2
	chPad1
	chPad2
	chBass
	chKick
	chRim
	chShake
	chVinyl
	numChannels = 9
)

const (
	// a 3/4 bar is 12 rows: beats on rows 0, 4 and 8
	barLen = 12
	bars = 8
	rows = barLen * bars

	speed = 6
	// the slowest track in the repo so far is track5 at 84 BPM
	tempo = 70
)

var (
	// music-box lead, tuned to C-4
	leadRecipe = recipe.New(512).Triangle(recipe.Note(512, "C-4"), 0.22).Sine(recipe.Note(512, "C-5"), 0.09).Sine(recipe.Note(512, "C-3"), 0.06).String()

	// detuned glass pad: the second layer beats slowly against the first, so it needs OffCycle
	padRecipe = recipe.New(1024).Triangle(recipe.Note(1024, "C-4"), 0.17).Triangle(recipe.Note(1024, "C-4")+1.5, 0.14).Sine(recipe.Note(1024, "C-3"), 0.09).String()

	// bowed sub bass, tuned to C-4
	bassRecipe = recipe.New(256).Sine(recipe.Note(256, "C-4"), 0.28).Triangle(recipe.Note(256, "C-4"), 0.14).Sine(recipe.Note(256, "C-3"), 0.10).String()

	// felt kick: a 190 Hz sine heard at C-2 sounds near 48 Hz
	kickRecipe = recipe.New(2000).Sine(190, 0.55).Sine(95, 0.18).Noise(800, 0.04).String()

	// soft rim with a woody body
	rimRecipe = recipe.New(1500).Noise(1800, 0.30).Triangle(320, 0.13).String()

	// short shaker
	shakerRecipe = recipe.New(600).Noise(3500, 0.28).String()

	// vinyl crackle bed, looping under the whole song
	vinylRecipe = recipe.New(8000).Noise(1, 0.5).Noise(3000, 0.12).String()

	// progression is i - VI - III - VII in F# minor, turning to V at the end,
	// the same shape as track9's loop a whole step up
	progression = []string{"F#m", "D", "A", "E", "F#m", "D", "A", "C#"}

	// verse asks twice, climbing out of the chord and falling back
	verse = []hit{
		{0, 0, "A-4"}, {0, 6, "C#5"},
		{1, 0, "D-5"}, {1, 8, "C#5"},
		{2, 0, "E-5"}, {2, 6, "C#5"},
		{3, 0, "B-4"}, {3, 6, "A-4"},
		{4, 0, "A-4"}, {4, 6, "E-5"},
		{5, 0, "F#5"}, {5, 8, "D-5"},
		{6, 0, "C#5"}, {6, 6, "B-4"},
		{7, 0, "C#5"}, {7, 8, "G#4"},
	}

	// chorus answers higher and denser, spelling the chords out
	chorus = []hit{
		{0, 0, "C#5"}, {0, 4, "E-5"}, {0, 8, "F#5"},
		{1, 0, "D-5"}, {1, 6, "F#5"},
		{2, 0, "A-5"}, {2, 4, "G-5"}, {2, 8, "E-5"},
		{3, 0, "B-4"}, {3, 4, "E-5"}, {3, 8, "G#5"},
		{4, 0, "A-5"}, {4, 6, "E-5"},
		{5, 0, "D-5"}, {5, 4, "F#5"}, {5, 8, "A-5"},
		{6, 0, "E-5"}, {6, 4, "C#5"}, {6, 8, "A-4"},
		// F-5 is the E# that makes C# major, written enharmonically
		{7, 0, "C#5"}, {7, 4, "F-5"}, {7, 8, "G#5"},
	}

	// break drifts low and sparse, leaving room for the drums to return
	breakPhrase = []hit{
		{0, 0, "E-4"}, {0, 8, "A-4"},
		{1, 0, "F#4"},
		{2, 0, "A-4"}, {2, 8, "C#5"},
		{3, 0, "B-4"},
		{4, 4, "A-4"},
		{5, 0, "D-5"},
		{6, 0, "A-4"},
		{7, 0, "G#4"},
	}

	// outro resolves down to the root and lets go
	outroPhrase = []hit{
		{0, 0, "C#5"},
		{1, 6, "B-4"},
		{2, 0, "A-4"},
	}
)

// at returns the row of row r in bar b, counting 12 rows to the bar
func at(bar, row int) int {
	return bar*barLen + row
}

// pads holds the chord of bar b on the three pad voices, spread across the stereo field
func (w pattern) pads(b int, chord string, vol int) {
	if vol == 0 {
		return
	}

	pans := [3]int{0x50, 0x80, 0xB0}

	for i, note := range song.Triad(chord, 3) {
		w.p.Hit(at(b, 0), chPad0+i, note, w.k.pad, song.VolOpts(vol, song.Pan(pans[i]))...)
	}
}

// bass bows the root on beat one, answering with the fifth on beat three when full
func (w pattern) bass(b int, chord string, vol int, full bool) {
	if vol == 0 {
		return
	}

	w.p.Hit(at(b, 0), chBass, song.Triad(chord, 1)[0], w.k.bass, song.VolOpts(vol)...)

	if full {
		w.p.Hit(at(b, 8), chBass, song.Triad(chord, 1)[2], w.k.bass, song.VolOpts(vol-8)...)
	}
}

// drums plays the waltz: felt kick on one, shaker on two and three, rim answering on three.
// Style 0 is silent, 1 is sparse and 2 is full
func (w pattern) drums(b, style int, pickup bool) {
	if style == 0 {
		return
	}

	w.p.Hit(at(b, 0), chKick, "C-2", w.k.kick, song.VolOpts(0x44)...)

	if pickup {
		w.p.Hit(at(b, 10), chKick, "C-2", w.k.kick, song.VolOpts(0x36)...)
	}

	beats := []int{4, 8}

	if style == 2 {
		beats = []int{2, 4, 6, 8, 10}
	}

	for i, r := range beats {
		vol := 0x22

		if r%4 == 0 {
			vol = 0x2A
		}

		pan := 0x90

		if i%2 == 0 {
			pan = 0x70
		}

		w.p.Hit(at(b, r), chShake, "C-5", w.k.shaker, song.VolOpts(vol, song.Pan(pan))...)
	}

	if style == 1 {
		w.p.Hit(at(b, 8), chRim, "C-4", w.k.rim, song.VolOpts(0x2C, song.Pan(0x98))...)

		return
	}

	w.p.Hit(at(b, 8), chRim, "C-4", w.k.rim, song.VolOpts(0x38, song.Pan(0x98))...)

	if b%2 == 1 {
		w.p.Hit(at(b, 4), chRim, "C-4", w.k.rim, song.VolOpts(0x2E, song.Pan(0x98))...)
	}
}

// fill rolls the rim into the next pattern over beat three of bar b
func (w pattern) fill(b int) {
	for i, r := range []int{8, 9, 10, 11} {
		w.p.Hit(at(b, r), chRim, "C-4", w.k.rim, song.VolOpts(0x28+i*5)...)
	}
}

// melody plays a phrase of music-box notes, each held until the next note
func (w pattern) melody(phrase []hit, vol int) {
	if vol == 0 {
		return
	}

	for i, h := range phrase {
		row := at(h.bar, h.row)

		w.p.Hit(row, chLead, h.note, w.k.lead, song.VolOpts(vol, song.Vibrato(4, 2))...)

		off := rows - 1

		if i+1 < len(phrase) {
			next := phrase[i+1]
			off = at(next.bar, next.row) - 1
		}

		if off > row {
			w.p.Hit(off, chLead, "###", nil)
		}
	}
}

// arrange lays out an eight-bar section, with the layers that section asks for
func (w pattern) arrange(sec section) {
	for b, chord := range progression {
		w.pads(b, chord, sec.padVol)
		w.bass(b, chord, sec.bassVol, sec.drumStyle == 2)
		w.drums(b, sec.drumStyle, sec.fill && b == bars-1)
	}

	w.melody(sec.phrase, sec.leadVol)

	if sec.fill {
		w.fill(bars - 1)
	}
}

// newSection starts a pattern of eight 3/4 bars and lays out the section in it
func newSection(s *song.Song, k *kit, sec section) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.arrange(sec)

	return w.p
}

// vinyl starts the crackle bed, which loops until the outro lets it ring out alone
func (w pattern) vinyl() {
	w.p.Hit(0, chVinyl, "C-4", w.k.vinyl, song.VolOpts(0x1A)...)
}

// intro opens with pads and vinyl alone
func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.vinyl()

	for b, chord := range progression {
		w.pads(b, chord, 0x26)
	}

	return w.p
}

// outro thins out to a final held chord, then cuts everything but the vinyl
func outro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b, chord := range progression[:3] {
		w.pads(b, chord, 0x2A)
		w.bass(b, chord, 0x34, false)
		w.drums(b, 1, false)
	}

	w.melody(outroPhrase, 0x38)

	// the last chord, held while the lead resolves
	last := progression[3]
	w.pads(3, last, 0x28)
	w.bass(3, last, 0x34, false)

	for _, ch := range []int{chLead, chPad0, chPad1, chPad2, chBass, chKick, chRim, chShake} {
		w.p.Hit(at(4, 0), ch, "###", nil)
	}

	return w.p
}

func main() {
	s := song.New("drowned waltz", numChannels)

	s.Timing(speed, tempo)

	k := &kit{
		lead: s.Instrument(song.Instrument{
			Name: "music box",
			Recipe: leadRecipe,
			Loop: true,
			Volume: 40,
			Pan: 96,
			Envelope: song.Points(0, 64, 36, 12),
			Fadeout: 1024,
		}),
		pad: s.Instrument(song.Instrument{
			Name: "glass pad",
			Recipe: padRecipe,
			OffCycle: true,
			Loop: true,
			Volume: 30,
			Pan: 128,
			Envelope: song.Points(0, 0, 48, 64),
			Fadeout: 2048,
		}),
		bass: s.Instrument(song.Instrument{
			Name: "bowed sub",
			Recipe: bassRecipe,
			Loop: true,
			Volume: 56,
			Pan: 128,
			Envelope: song.Points(0, 0, 10, 64),
			Fadeout: 1024,
		}),
		kick: s.Instrument(song.Instrument{
			Name: "felt kick",
			Recipe: kickRecipe,
			Volume: 58,
			Pan: 128,
			Envelope: song.Points(0, 64, 12, 0),
		}),
		rim: s.Instrument(song.Instrument{
			Name: "soft rim",
			Recipe: rimRecipe,
			Volume: 44,
			Pan: 152,
			Envelope: song.Points(0, 64, 10, 0),
		}),
		shaker: s.Instrument(song.Instrument{
			Name: "shaker",
			Recipe: shakerRecipe,
			Volume: 30,
			Pan: 128,
			Envelope: song.Points(0, 64, 2, 0),
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
	verseAP := newSection(s, k, section{
		padVol: 0x2C,
		bassVol: 0x38,
		drumStyle: 1,
	})
	verseBP := newSection(s, k, section{
		padVol: 0x2E,
		bassVol: 0x3A,
		drumStyle: 1,
		leadVol: 0x40,
		phrase: verse,
	})
	chorusP := newSection(s, k, section{
		padVol: 0x34,
		bassVol: 0x40,
		drumStyle: 2,
		leadVol: 0x48,
		phrase: chorus,
		fill: true,
	})
	breakP := newSection(s, k, section{
		padVol: 0x2A,
		bassVol: 0x34,
		drumStyle: 1,
		leadVol: 0x38,
		phrase: breakPhrase,
	})
	verseCP := newSection(s, k, section{
		padVol: 0x2E,
		bassVol: 0x3A,
		drumStyle: 1,
		leadVol: 0x3C,
		phrase: chorus,
	})
	outroP := outro(s, k)

	s.Play(introP, verseAP, verseBP, chorusP, breakP, verseCP, chorusP, outroP)

	// write next to this file, wherever the program is run from
	if err := s.WriteBeside(); err != nil {
		log.Fatal(err)
	}
}
