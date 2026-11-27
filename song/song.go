// Package song is a Go API for writing tracker modules as code.
//
// A song is built from instruments, patterns and an order. Instruments take their samples from recipes
// (see package recipe), so a song folder needs no other files. A small program writes the module:
//
//	s := song.New("demo", 2)
//	lead := s.Instrument(song.Instrument{Name: "lead", Recipe: "Frames: 256\nLayer: square 261.34375 0.5\n"})
//	p := s.Pattern(64)
//	p.Line(0, 0, 4, song.Triad("C", 4), lead)
//	s.Play(p)
//	if err := s.Write("song.xm"); err != nil { log.Fatal(err) }
//
// Notes are written as tracker note names, such as "C-4" or "A#3". "###" is note off.
// Rows are numbered from 0, and a bar is RowsPerBar rows.
package song

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Heliodex/tracker/loadtext"
	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/save"
	"github.com/Heliodex/tracker/types"
)

type (
	// Point is a point on an envelope: a tick, and a volume from 0 to 64
	Point struct {
		X, Y int
	}

	// Instrument is a sound: a recipe for its sample, and how it is played.
	Instrument struct {
		// Name is at most 22 characters
		Name string
		// Recipe is the sample, as recipe text (see package recipe)
		Recipe string
		// Loop repeats the whole sample. Without it, the sample plays once
		Loop bool
		// OffCycle allows a looping sample to have layers that are not whole cycles, such as a detuned unison that beats.
		// The loop then has a small click or phase jump at its end. Without it, such a layer is an error (see recipe.Loopable)
		OffCycle bool
		// Volume is the sample volume from 0 to 64. Zero means 64
		Volume int
		// Pan is the sample panning from 0 (left) to 255 (right). Zero means centre, 128
		Pan int
		// Envelope is the volume envelope, as points of tick and volume (0 to 64).
		// Empty means a flat envelope at full volume, which keeps the note-off fade working
		Envelope []Point
		// Fadeout is how fast the volume falls after a note off, from 0 to 65535. Zero means 256
		Fadeout, id int
		recipe *recipe.Recipe
		song *Song
	}

	// Song is a module under construction
	Song struct {
		name string
		channels, speed, tempo int
		linear bool
		instruments []*Instrument
		patterns, order []*Pattern
		err error
	}
)

// RowsPerBar is the number of rows in a bar, for laying out a song in bars
const RowsPerBar = 16

// fail records the first error, which Build returns
func (s *Song) fail(format string, args ...any) {
	if s.err == nil {
		s.err = fmt.Errorf(format, args...)
	}
}

// New starts a song with the given name, which is at most 20 characters, and number of channels from 1 to 32.
// Speed and tempo start at 6 and 125, with linear slides.
func New(name string, channels int) *Song {
	s := &Song{
		name: name,
		channels: channels,
		speed: 6,
		tempo: 125,
		linear: true,
	}

	if len(name) > 20 {
		s.fail("song name is longer than 20 characters: %q", name)
	}

	if channels < 1 || channels > 32 {
		s.fail("channels must be between 1 and 32, got %d", channels)
	}

	return s
}

// Instrument adds an instrument to the song, and returns it for use in patterns
func (s *Song) Instrument(ins Instrument) *Instrument {
	in := ins
	in.id = len(s.instruments)
	in.song = s

	if len(in.Name) > 22 {
		s.fail("instrument name is longer than 22 characters: %q", in.Name)
	}

	if in.Volume < 0 || in.Volume > 64 {
		s.fail("instrument %q: volume must be between 0 and 64, got %d", in.Name, in.Volume)
	}

	if in.Pan < 0 || in.Pan > 255 {
		s.fail("instrument %q: pan must be between 0 and 255, got %d", in.Name, in.Pan)
	}

	if in.Fadeout < 0 || in.Fadeout > 65535 {
		s.fail("instrument %q: fadeout must be between 0 and 65535, got %d", in.Name, in.Fadeout)
	}

	rc, err := recipe.Parse(strings.NewReader(in.Recipe))
	if err != nil {
		s.fail("instrument %q: %v", in.Name, err)
	}

	if in.Loop && !in.OffCycle {
		if err := rc.Loopable(); err != nil {
			s.fail("instrument %q: %v (set OffCycle if the beating is intended)", in.Name, err)
		}
	}

	in.recipe = rc
	s.instruments = append(s.instruments, &in)

	return &in
}

// toText converts the instrument to the form the text format and XM writer use
func (ins *Instrument) toText() (loadtext.Instrument, error) {
	rc := ins.recipe
	if rc == nil {
		return loadtext.Instrument{}, fmt.Errorf("instrument %q has no recipe", ins.Name)
	}

	volume, pan, fadeout := ins.Volume, ins.Pan, ins.Fadeout
	if volume == 0 {
		volume = 64
	}

	if pan == 0 {
		pan = 128
	}

	if fadeout == 0 {
		fadeout = 256
	}

	li := loadtext.Instrument{
		VolumeFadeout: uint16(fadeout),
	}

	copy(li.Name[:], ins.Name)

	// the envelope is always on, so a note off fades out, as it does in the original trackers
	points := ins.Envelope
	if len(points) == 0 {
		points = []Point{
			{
				X: 0,
				Y: 64,
			},
		}
	}

	if len(points) > 12 {
		return loadtext.Instrument{}, fmt.Errorf("instrument %q has more than 12 envelope points", ins.Name)
	}

	for _, pt := range points {
		if pt.X < 0 || pt.X > 65535 || pt.Y < 0 || pt.Y > 64 {
			return loadtext.Instrument{}, fmt.Errorf("instrument %q: envelope point %v is out of range", ins.Name, pt)
		}

		li.VolPoints = append(li.VolPoints, types.EnvPoint{
			X: uint16(pt.X),
			Y: uint16(pt.Y),
		})
	}

	li.VolFlags = 1
	sample := loadtext.Sample{
		Volume: uint8(volume),
		Panning: uint8(pan),
		Data: rc.Render(),
		Sixteen: rc.Sixteen,
	}

	copy(sample.Name[:], ins.Name)

	// the loop length is the whole sample, even for a one-shot, where the loop mode ignores it
	sample.LoopLength = uint32(rc.Frames)

	if ins.Loop {
		sample.LoopMode = 1
	}

	li.Samples = []loadtext.Sample{sample}

	return li, nil
}

// Build converts the song to an XM file representation, ready to save
func (s *Song) Build() (*types.File, error) {
	if s.err != nil {
		return nil, s.err
	}

	if len(s.order) > 256 {
		return nil, errors.New("order is longer than 256 entries")
	}

	tf := &loadtext.TextFormat{
		NumChannels: uint16(s.channels),
		NumPatterns: uint16(len(s.patterns)),
		NumInstruments: uint16(len(s.instruments)),
		Speed: uint16(s.speed),
		Tempo: uint16(s.tempo),
		LinearSlides: s.linear,
	}

	copy(tf.Name[:], s.name)

	for _, p := range s.order {
		tf.OrderTable = append(tf.OrderTable, uint8(p.index))
	}

	for _, p := range s.patterns {
		tf.Patterns = append(tf.Patterns, p.rows)
	}

	for _, ins := range s.instruments {
		li, err := ins.toText()
		if err != nil {
			return nil, err
		}

		tf.Instruments = append(tf.Instruments, li)
	}

	return tf.ToFile()
}

// Pattern adds an empty pattern of rows rows to the song. Patterns are played with Play
func (s *Song) Pattern(rows int) *Pattern {
	if rows < 1 || rows > 256 {
		s.fail("pattern rows must be between 1 and 256, got %d", rows)

		rows = 1
	}

	p := &Pattern{
		song: s,
		index: len(s.patterns),
		rows: make(loadtext.Pattern, rows),
	}

	for r := range p.rows {
		p.rows[r] = make([]loadtext.ChannelData, s.channels)
	}

	s.patterns = append(s.patterns, p)

	return p
}

// Play appends patterns to the order, so they play in this sequence
func (s *Song) Play(patterns ...*Pattern) {
	for _, p := range patterns {
		if p.song != s {
			s.fail("pattern %d belongs to another song", p.index)
			continue
		}

		s.order = append(s.order, p)
	}
}

// Timing sets the speed (ticks per row) and tempo (beats per minute)
func (s *Song) Timing(speed, tempo int) {
	if speed < 1 || speed > 255 {
		s.fail("speed must be between 1 and 255, got %d", speed)
		return
	}

	if tempo < 32 || tempo > 255 {
		s.fail("tempo must be between 32 and 255, got %d", tempo)
		return
	}

	s.speed = speed
	s.tempo = tempo
}

// Write builds the song and writes it to path as an XM file
func (s *Song) Write(path string) error {
	f, err := s.Build()
	if err != nil {
		return err
	}

	var b bytes.Buffer

	if err := save.Write(&b, f); err != nil {
		return err
	}

	return os.WriteFile(path, b.Bytes(), 0o644)
}
