// Command track2 writes "cool guy", an XM module, song.xm, in this folder.
//
// Run it from anywhere inside the repository with go run ./tracks/track2, or with tracker build.
package main

import (
	"embed"
	"log"
	"strconv"
	"strings"

	"github.com/Heliodex/tracker/song"
)

// rows is the length of every pattern
const rows = 128

// samples holds the WAV files in the samples folder, which are built into the program
//
//go:embed samples
var samples embed.FS

// fill sets every cell of channel ch to an empty note with opt, for the notes that are played over it
func fill(p *song.Pattern, ch int, opt song.Opt) {
	for r := range rows {
		p.Hit(r, ch, "...", nil, opt)
	}
}

// play hits row of channel ch with cell, a note such as "C-5", or "C-5:2" to play instrument 2, which otherwise is def
func play(p *song.Pattern, k []*song.Instrument, row, ch int, cell string, def int, opts ...song.Opt) {
	note, n, found := strings.Cut(cell, ":")
	if found {
		def, _ = strconv.Atoi(n)
	}

	p.Hit(row, ch, note, k[def-1], opts...)
}

// hits plays entries on channel ch, each written as "row:cell", with the options on every cell
func hits(p *song.Pattern, k []*song.Instrument, ch, def int, entries string, opts ...song.Opt) {
	for e := range strings.FieldsSeq(entries) {
		r, cell, _ := strings.Cut(e, ":")

		row, err := strconv.Atoi(r)
		if err != nil {
			log.Fatal(err)
		}

		play(p, k, row, ch, cell, def, opts...)
	}
}

// sample returns the WAV file called name from the samples folder
func sample(name string) []byte {
	b, err := samples.ReadFile("samples/" + name)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

// instruments adds the instruments to the song. They are numbered from 1 in the order they are listed here
func instruments(s *song.Song) []*song.Instrument {
	return []*song.Instrument{
		s.Instrument(song.Instrument{
			Name: "main",
			WAV: sample("00-00-m1-yn.wav"),
			Loop: true,
			Volume: 32,
			Envelope: song.Points(0, 64, 16, 0),
		}),
		s.Instrument(song.Instrument{
			Name: "cymbal",
			WAV: sample("01-00-noise.wav"),
			Loop: true,
			Envelope: song.Points(0, 0, 1, 64, 10, 15, 20, 0),
		}),
		s.Instrument(song.Instrument{
			Name: "boom",
			WAV: sample("02-00-boom.wav"),
			Envelope: song.Points(0, 64, 32, 64),
		}),
		s.Instrument(song.Instrument{
			Name: "beep",
			WAV: sample("03-00-m1-yn.wav"),
			Loop: true,
			Envelope: song.Points(0, 64, 16, 0),
		}),
		s.Instrument(song.Instrument{
			Name: "cymbal2",
			WAV: sample("04-00-noise2.wav"),
			Envelope: song.Points(0, 64, 6, 0),
		}),
		s.Instrument(song.Instrument{
			Name: "boom2",
			WAV: sample("05-00-m1-y-x-7m-ap.wav"),
			Envelope: song.Points(0, 64, 16, 0),
		}),
	}
}

// slots plays entries on channel ch, one every step rows from row 0. An entry is a cell as play takes it,
// "..." for an empty note, or "." for no cell
func slots(p *song.Pattern, k []*song.Instrument, ch, step, def int, entries string) {
	for i, e := range strings.Fields(entries) {
		if e != "." {
			play(p, k, i*step, ch, e, def)
		}
	}
}

func pattern0(s *song.Song, k []*song.Instrument) *song.Pattern {
	p := s.Pattern(rows)

	slots(p, k, 0, 4, 1, "F-5 F-4 ... F-5 F-4 ... F-5 G-5 F-5 F-4 ... F-5 F-4 ... E-5 D-5 C-5 ... ... G-5 C-5 ... G-5 A-5 C-5 ... ... G-5 C-5 ... F-5 E-5")

	return p
}

func pattern1(s *song.Song, k []*song.Instrument) *song.Pattern {
	p := s.Pattern(rows)

	slots(p, k, 0, 4, 1, "D-5 ... ... A-5 D-5 ... G-5 A-5 D-5 ... ... A-5 D-5 ... A#5 A#4 A-5 A#4 ... A-5 A#4 ... A-5 A#5 A-5 A#4 ... A-5 A#4 F-5 G-5 F-4")

	return p
}

func pattern2(s *song.Song, k []*song.Instrument) *song.Pattern {
	p := s.Pattern(rows)

	slots(p, k, 0, 4, 1, "D-5 ... ... A-5 D-5 ... G-5 A-5 D-5 ... ... A-5 D-5 ... A#5 A#4 A-5 A#4 ... A-5 A#4 ... A-5 A#5 A-5 A#4 ... A-5 A#4 F-5 G-5 F-4")
	hits(p, k, 1, 3, "0:C-5 16:E-5:6 32:C-5 48:E-5:6 64:C-5 80:E-5:6 96:C-5 116:C-5 120:E-5:6")
	slots(p, k, 2, 4, 5, "C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 . . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 .")
	p.Hit(56, 2, "C-5", k[1], song.Vol(32))
	fill(p, 3, song.Pan(0))
	hits(p, k, 3, 4, "0:D-6 8:D-6 12:D-6 24:D-6 26:E-6 28:D-6 32:G-6 34:A-6 36:G-6 40:F-6 42:G-6 44:F-6 48:E-6 50:F-6 52:G-6 56:E-6 58:F-6 60:G-6 62:A-5 64:F-6 66:F-5 68:F-6 70:F-5 72:F-6 74:F-5 76:F-6 78:F-5 80:E-6 82:F-5 84:E-6 86:F-5 88:E-6 90:F-5 92:E-6 94:F-5 96:G-6 100:A-6 104:E-6 108:D-6 112:F-6 116:E-6 120:C-6 124:A-5", song.Pan(0))
	fill(p, 4, song.Pan(255))
	hits(p, k, 4, 4, "0:D-5 4:D-5 8:D-5 12:D-5 16:D-5 20:D-5 24:D-5 26:E-5 28:D-5 32:G-5 34:A-5 36:G-5 40:F-5 42:G-5 44:F-5 48:E-5 50:F-5 52:G-5 56:E-5 57:... 58:F-5 60:G-5 62:A-4 64:F-5 66:F-4 68:F-5 70:F-4 72:F-5 74:F-4 76:F-5 78:F-4 80:E-5 82:F-4 84:E-5 86:F-4 88:E-5 90:F-4 92:E-5 94:F-4 96:G-5 102:A-5 106:E-5 110:D-5 114:F-5 118:E-5 122:C-5 126:A-4", song.Pan(255))
	slots(p, k, 5, 4, 1, "D-4 ... ... A-4 D-4 ... G-4 A-4 D-4 ... ... A-4 D-4 ... A#4 A#3 A-4 A#3 ... A-4 A#3 ... A-4 A#4 A-4 A#3 ... A-4 A#3 F-4 G-4 F-3")

	return p
}

func pattern3(s *song.Song, k []*song.Instrument) *song.Pattern {
	p := s.Pattern(rows)

	slots(p, k, 0, 4, 1, "F-5 F-4 ... F-5 F-4 ... F-5 G-5 F-5 F-4 ... F-5 F-4 ... E-5 D-5 C-5 ... ... G-5 C-5 ... G-5 A-5 C-5 ... ... G-5 C-5 ... F-5 E-5")
	hits(p, k, 1, 3, "0:C-5 16:E-5:6 32:C-5 48:E-5:6 64:C-5 80:E-5:6 96:C-5 112:E-5:6")
	slots(p, k, 2, 4, 5, "C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 .")
	fill(p, 3, song.Pan(0))
	hits(p, k, 3, 4, "0:F-5 12:F-5 20:F-5 24:G-5 28:A-5 32:F-5 44:F-5 52:F-5 56:G-5 60:A-5 64:A-5 76:A-5 84:A-5 88:A-5 90:A#5 92:A-5 96:A-5 108:A-5 116:A-5 120:A#5 124:A-5", song.Pan(0))
	fill(p, 4, song.Pan(255))
	hits(p, k, 4, 4, "0:F-4 8:F-4 12:F-4 16:F-4 20:G-4 24:A-4 28:F-4 32:F-4 40:F-4 44:F-4 48:F-4 52:F-4 56:G-4 60:A-4 64:A-4 72:A-4 76:A-4 80:A-4 84:A-4 88:A-4 90:A#4 92:A-4 96:A-4 104:A-4 108:A-4 112:A-4 116:A-4 120:A#4 124:A-4", song.Pan(255))
	slots(p, k, 5, 4, 1, "F-4 F-3 ... F-4 F-3 ... F-4 G-4 F-4 F-3 ... F-4 F-3 ... E-4 D-4 C-4 ... ... G-4 C-4 ... G-4 A-4 C-4 ... ... G-4 C-4 ... F-4 E-4")

	return p
}

func pattern4(s *song.Song, k []*song.Instrument) *song.Pattern {
	p := s.Pattern(rows)

	slots(p, k, 0, 4, 1, "D-5 ... ... A-5 D-5 ... G-5 A-5 D-5 ... ... A-5 D-5 ... A#5 A#4 A-5 A#4 ... A-5 A#4 ... A-5 A#5 A-5 A#4 ... A-5 A#4 F-5 G-5 F-4")
	hits(p, k, 1, 3, "0:C-5 16:E-5:6 32:C-5 48:E-5:6 64:C-5 80:E-5:6 96:C-5 116:C-5 120:E-5:6")
	slots(p, k, 2, 4, 5, "C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 . . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 .")
	p.Hit(56, 2, "C-5", k[1], song.Vol(32))
	fill(p, 3, song.Pan(0))
	hits(p, k, 3, 4, "0:D-6 8:D-6 12:D-6 24:D-6 26:E-6 28:D-6 32:D-6 34:E-6 36:D-6 40:C-6 42:D-6 44:C-6 48:B-5 50:C-6 52:B-5 56:A-5 58:B-5 60:A-5 64:D-6 68:C-6 72:B-5 76:A-5 80:G-5 84:F-5 88:G-5 92:E-5 96:D-5 100:D-4 104:D-5 108:D-4 112:D-5 116:D-4 120:D-5 124:E-5", song.Pan(0))
	fill(p, 4, song.Pan(255))
	hits(p, k, 4, 4, "0:D-5 4:D-5 8:D-5 12:D-5 16:D-5 20:D-5 24:D-5 26:E-5 28:D-5 32:D-5 34:E-5 36:D-5 40:C-5 42:D-5 44:C-5 48:B-4 50:C-5 52:B-4 56:A-4 58:B-4 60:A-4 64:D-5 68:C-5 72:B-4 76:A-4 80:G-4 84:F-4 88:G-4 92:E-4 96:D-4 100:D-3 104:D-4 108:D-3 112:D-4 116:D-3 120:D-4 124:E-4", song.Pan(255))
	slots(p, k, 5, 4, 1, "D-4 ... ... A-4 D-4 ... G-4 A-4 D-4 ... ... A-4 D-4 ... A#4 A#3 A-4 A#3 ... A-4 A#3 ... A-4 A#4 A-4 A#3 ... A-4 A#3 F-4 G-4 F-3")

	return p
}

func pattern5(s *song.Song, k []*song.Instrument) *song.Pattern {
	p := s.Pattern(rows)

	slots(p, k, 0, 4, 1, "F-5 F-4 ... F-5 F-4 ... F-5 G-5 F-5 F-4 ... F-5 F-4 ... E-5 D-5 C-5 ... ... G-5 C-5 ... G-5 A-5 C-5 ... ... G-5 C-5 ... F-5 G-5")
	slots(p, k, 2, 4, 5, "C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 .")

	return p
}

func pattern6(s *song.Song, k []*song.Instrument) *song.Pattern {
	p := s.Pattern(rows)

	slots(p, k, 0, 4, 1, "D-5 ... ... A-5 D-5 ... G-5 A-5 D-5 ... ... A-5 D-5 ... A#5 A#4 A-5 A#4 ... A-5 A#4 ... A-5 A#5 A-5 A#4 ... A-5 A#4 F-5 G-5 F-4")
	hits(p, k, 1, 3, "116:C-5 120:C-4:6")
	slots(p, k, 2, 4, 5, "C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-6 C-6 C-6 .")

	return p
}

func pattern7(s *song.Song, k []*song.Instrument) *song.Pattern {
	p := s.Pattern(rows)

	slots(p, k, 0, 4, 1, "F-5 F-4 ... F-5 F-4 ... F-5 G-5 F-5 F-4 ... F-5 F-4 ... E-5 D-5 C-5 ... ... G-5 C-5 ... G-5 A-5 C-5 ... ... G-5 C-5 ... F-5 G-5")
	slots(p, k, 2, 4, 5, "C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 .")
	slots(p, k, 5, 4, 1, "F-4 F-3 ... F-4 F-3 ... F-4 G-4 F-4 F-3 ... F-4 F-3 ... E-4 D-4 C-4 ... ... G-4 C-4 ... G-4 A-4 C-4 ... ... G-4 C-4 ... F-4 E-4")

	return p
}

func pattern8(s *song.Song, k []*song.Instrument) *song.Pattern {
	p := s.Pattern(rows)

	slots(p, k, 0, 4, 1, "D-5 ... ... A-5 D-5 ... G-5 A-5 D-5 ... ... A-5 D-5 ... A#5 A#4 A-5 A#4 ... A-5 A#4 ... A-5 A#5 A-5 A#4 ... A-5 A#4 F-5 G-5 F-4")
	hits(p, k, 1, 3, "116:C-5 120:C-4:6")
	slots(p, k, 2, 4, 5, "C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-7 C-7 C-5:2 . C-6 C-6 C-6 .")
	slots(p, k, 5, 4, 1, "D-4 ... ... A-4 D-4 ... G-4 A-4 D-4 ... ... A-4 D-4 ... A#4 A#3 A-4 A#3 ... A-4 A#3 ... A-4 A#4 A-4 A#3 ... A-4 A#3 F-4 G-4 F-3")

	return p
}

func main() {
	s := song.New("cool guy", 6)

	s.Timing(3, 190)

	k := instruments(s)
	patterns := []*song.Pattern{
		pattern0(s, k),
		pattern1(s, k),
		pattern2(s, k),
		pattern3(s, k),
		pattern4(s, k),
		pattern5(s, k),
		pattern6(s, k),
		pattern7(s, k),
		pattern8(s, k),
	}

	for _, i := range []int{0, 1, 5, 6, 7, 8, 3, 4, 3, 2} {
		s.Play(patterns[i])
	}

	// write next to this file, wherever the program is run from
	if err := s.WriteBeside(); err != nil {
		log.Fatal(err)
	}
}
