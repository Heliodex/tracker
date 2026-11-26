// Package loadtext implements a text-based format for tracker music.
//
// A file looks like this:
//
//	Name: cool guy
//	NumChannels: 2
//	NumPatterns: 1
//	NumInstruments: 1
//	Speed: 6
//	Tempo: 125
//	LinearSlides: 1
//	Order: 0
//
//	Pattern 0
//	Rows: 2
//	Row 000: C-4 01 .. .. .. | ... .. .. .. ..
//	Row 001: ### .. .. .. .. | A#5 02 .. .. ..
//
//	Instrument 0
//	Name: kick
//	Type: 0
//	Keymap: 0 0 0 ... (96 values)
//	VolumeEnvelope: 0,64 10,32
//	VolumeSustain: 0
//	VolumeLoop: 0 0
//	VolumeFlags: 0
//	PanEnvelope:
//	PanSustain: 0
//	PanLoop: 0 0
//	PanFlags: 0
//	Vibrato: 0 0 0 0
//	VolumeFadeout: 0
//	Samples: 1
//
//	Sample 0
//	Name: kick
//	File: samples/00-00-kick.wav
//	Volume: 64
//	Finetune: 0
//	Panning: 128
//	RelativeNote: 0
//	LoopMode: 0
//	LoopStart: 0
//	LoopLength: 0
//
// Each channel cell is a note followed by four hex columns: instrument, volume, effect and effect parameter.
// "..." is an empty note, "###" is note off, and ".." is a zero value.
//
// Sample data is not part of the text. File is the path to a sample, relative to the text file, without its extension.
// It resolves to a .wav file, or to a .txt recipe, see package project.
// Loop points are in frames (samples), not bytes.
package loadtext

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Heliodex/tracker/types"
)

type (
	// ChannelData is a single cell of a pattern
	ChannelData struct {
		Note, Instrument, Volume, Effect, EffectParameter uint8
	}

	// Pattern is a list of rows, each holding one ChannelData per channel
	Pattern [][]ChannelData

	// Sample is one sample of an instrument
	Sample struct {
		Name [22]uint8
		File string
		Volume uint8
		Finetune int8
		Panning uint8
		RelativeNote int8
		LoopMode uint8
		LoopStart uint32
		LoopLength uint32

		// Sixteen is set from the WAV file when it is read, and is not stored in the text
		Sixteen bool
		// Data is the sample data in XM form, set from the WAV file when it is read, and is not stored in the text
		Data []uint8
	}

	// Instrument is an instrument, with its envelopes, vibrato and samples
	Instrument struct {
		Name [22]uint8
		Type uint8
		Keymap [96]uint8
		VolPoints, PanPoints []types.EnvPoint
		VolSustain, VolLoopStart, VolLoopEnd, VolFlags uint8
		PanSustain, PanLoopStart, PanLoopEnd, PanFlags uint8
		VibratoType, VibratoSweep, VibratoDepth, VibratoRate uint8
		VolumeFadeout uint16
		Samples []Sample
	}

	// TextFormat is the text representation of a module
	TextFormat struct {
		Name [20]uint8
		NumChannels, NumPatterns, NumInstruments uint16
		Speed, Tempo uint16
		LinearSlides bool
		OrderTable []uint8
		Patterns []Pattern
		Instruments []Instrument
	}

	parser struct {
		r *bufio.Reader
		line int
	}
)

const (
	maxChannels = 32
	maxPatterns = 256
	maxRows = 256
	maxOrder = 256
	maxInstruments = 128
	maxEnvPoints = 12
	maxSamples = 255

	// Order entries 254 and 255 are skip and end markers in XM, not patterns
	orderSkip = 254
	orderEnd = 255
)

// readLine returns the next line without its line ending
func (p *parser) readLine() (string, error) {
	line, err := p.r.ReadString('\n')
	if err != nil && (err != io.EOF || line == "") {
		return "", err
	}

	p.line++

	return strings.TrimRight(line, "\r\n"), nil
}

// header reads a line of the form "Kind index", such as "Pattern 0"
func (p *parser) header(kind string, index int) error {
	line, err := p.readLine()
	if err != nil {
		return fmt.Errorf("expected '%s %d': %w", kind, index, err)
	}

	rest, ok := strings.CutPrefix(line, kind+" ")
	if n, err := strconv.Atoi(rest); !ok || err != nil || n != index {
		return fmt.Errorf("expected '%s %d', got %q", kind, index, line)
	}

	return nil
}

// field reads a line of the form "Name: value"
func (p *parser) field(name string) (string, error) {
	line, err := p.readLine()
	if err != nil {
		return "", fmt.Errorf("expected '%s:': %w", name, err)
	}

	rest, ok := strings.CutPrefix(line, name+":")
	if !ok {
		return "", fmt.Errorf("expected '%s:', got %q", name, line)
	}

	return strings.TrimPrefix(rest, " "), nil
}

// numbers reads a "Name: value value ..." line of exactly count integers, each between lo and hi inclusive
func (p *parser) numbers(name string, count, lo, hi int) ([]int, error) {
	v, err := p.field(name)
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(v)
	if len(fields) != count {
		return nil, fmt.Errorf("%s: expected %d values, got %d", name, count, len(fields))
	}

	out := make([]int, count)

	for i, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil || n < lo || n > hi {
			return nil, fmt.Errorf("invalid %s value: %q", name, f)
		}

		out[i] = n
	}

	return out, nil
}

// intField reads a "Name: value" line holding one integer between lo and hi inclusive
func (p *parser) intField(name string, lo, hi int) (int, error) {
	v, err := p.numbers(name, 1, lo, hi)
	if err != nil {
		return 0, err
	}

	return v[0], nil
}

// points reads a "Name: x,y x,y ..." line of envelope points
func (p *parser) points(name string) ([]types.EnvPoint, error) {
	v, err := p.field(name)
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(v)
	if len(fields) > maxEnvPoints {
		return nil, fmt.Errorf("%s: too many points: %d", name, len(fields))
	}

	points := make([]types.EnvPoint, len(fields))

	for i, f := range fields {
		x, y, ok := strings.Cut(f, ",")

		xn, errX := strconv.ParseUint(x, 10, 16)
		yn, errY := strconv.ParseUint(y, 10, 16)

		if !ok || errX != nil || errY != nil {
			return nil, fmt.Errorf("invalid %s point: %q", name, f)
		}

		points[i] = types.EnvPoint{X: uint16(xn), Y: uint16(yn)}
	}

	return points, nil
}

// copyName copies a name into dst, failing if it does not fit
func copyName(dst []uint8, name string) error {
	if len(name) > len(dst) {
		return fmt.Errorf("name too long: %q", name)
	}

	copy(dst, name)

	return nil
}

func parseHex(s string) (uint8, error) {
	if s == ".." {
		return 0, nil
	}

	v, err := strconv.ParseUint(s, 16, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid hex value: %q", s)
	}

	return uint8(v), nil
}

func parseCell(s string) (ChannelData, error) {
	fields := strings.Fields(s)
	if len(fields) != 5 {
		return ChannelData{}, fmt.Errorf("expected 5 columns, got %q", s)
	}

	note, err := types.ParseNote(fields[0])
	if err != nil {
		return ChannelData{}, err
	}

	var hex [4]uint8

	for i, f := range fields[1:] {
		if hex[i], err = parseHex(f); err != nil {
			return ChannelData{}, err
		}
	}

	return ChannelData{
		Note: uint8(note),
		Instrument: hex[0],
		Volume: hex[1],
		Effect: hex[2],
		EffectParameter: hex[3],
	}, nil
}

// parseRow parses the cells of a row, separated by " | ", one per channel
func parseRow(s string, numChannels int) ([]ChannelData, error) {
	cells := strings.Split(s, " | ")
	if len(cells) != numChannels {
		return nil, fmt.Errorf("expected %d channels, got %d", numChannels, len(cells))
	}

	row := make([]ChannelData, numChannels)

	for c, cell := range cells {
		cd, err := parseCell(cell)
		if err != nil {
			return nil, fmt.Errorf("channel %d: %w", c+1, err)
		}

		row[c] = cd
	}

	return row, nil
}

func parsePattern(p *parser, index, numChannels int) (Pattern, error) {
	if err := p.header("Pattern", index); err != nil {
		return nil, err
	}

	numRows, err := p.intField("Rows", 1, maxRows)
	if err != nil {
		return nil, err
	}

	pattern := make(Pattern, numRows)

	for i := range pattern {
		line, err := p.readLine()
		if err != nil {
			return nil, fmt.Errorf("expected 'Row %03d': %w", i, err)
		}

		before, cells, ok := strings.Cut(line, ": ")
		rowStr, isRow := strings.CutPrefix(before, "Row ")

		if n, err := strconv.Atoi(rowStr); !ok || !isRow || err != nil || n != i {
			return nil, fmt.Errorf("expected 'Row %03d: ', got %q", i, line)
		}

		row, err := parseRow(cells, numChannels)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i, err)
		}

		pattern[i] = row
	}

	return pattern, nil
}

func parsePatterns(p *parser, tf *TextFormat) error {
	for i := range int(tf.NumPatterns) {
		blank, err := p.readLine()
		if err != nil || blank != "" {
			return errors.New("expected blank line before pattern")
		}

		pattern, err := parsePattern(p, i, int(tf.NumChannels))
		if err != nil {
			return err
		}

		tf.Patterns = append(tf.Patterns, pattern)
	}

	return nil
}

func parseSample(p *parser, index int) (Sample, error) {
	var s Sample

	if err := p.header("Sample", index); err != nil {
		return s, err
	}

	name, err := p.field("Name")
	if err != nil {
		return s, err
	}

	if err := copyName(s.Name[:], name); err != nil {
		return s, err
	}

	if s.File, err = p.field("File"); err != nil {
		return s, err
	}

	if s.File == "" {
		return s, errors.New("sample has no File")
	}

	volume, err := p.intField("Volume", 0, 64)
	if err != nil {
		return s, err
	}

	s.Volume = uint8(volume)

	finetune, err := p.intField("Finetune", -128, 127)
	if err != nil {
		return s, err
	}

	s.Finetune = int8(finetune)

	panning, err := p.intField("Panning", 0, 255)
	if err != nil {
		return s, err
	}

	s.Panning = uint8(panning)

	relative, err := p.intField("RelativeNote", -128, 127)
	if err != nil {
		return s, err
	}

	s.RelativeNote = int8(relative)

	loopMode, err := p.intField("LoopMode", 0, 3)
	if err != nil {
		return s, err
	}

	s.LoopMode = uint8(loopMode)

	loopStart, err := p.intField("LoopStart", 0, math32Max)
	if err != nil {
		return s, err
	}

	s.LoopStart = uint32(loopStart)

	loopLength, err := p.intField("LoopLength", 0, math32Max)
	if err != nil {
		return s, err
	}

	s.LoopLength = uint32(loopLength)

	return s, nil
}

// math32Max is the largest value a 32-bit field can hold
const math32Max = 1<<32 - 1

func parseInstrument(p *parser, index int) (Instrument, error) {
	var ins Instrument

	if err := p.header("Instrument", index); err != nil {
		return ins, err
	}

	name, err := p.field("Name")
	if err != nil {
		return ins, err
	}

	if err := copyName(ins.Name[:], name); err != nil {
		return ins, err
	}

	typ, err := p.intField("Type", 0, 255)
	if err != nil {
		return ins, err
	}

	ins.Type = uint8(typ)

	keymap, err := p.numbers("Keymap", 96, 0, 255)
	if err != nil {
		return ins, err
	}

	for i, k := range keymap {
		ins.Keymap[i] = uint8(k)
	}

	if ins.VolPoints, err = p.points("VolumeEnvelope"); err != nil {
		return ins, err
	}

	if ins.VolSustain, err = readUint8(p, "VolumeSustain"); err != nil {
		return ins, err
	}

	loop, err := p.numbers("VolumeLoop", 2, 0, 255)
	if err != nil {
		return ins, err
	}

	ins.VolLoopStart, ins.VolLoopEnd = uint8(loop[0]), uint8(loop[1])

	if ins.VolFlags, err = readUint8(p, "VolumeFlags"); err != nil {
		return ins, err
	}

	if ins.PanPoints, err = p.points("PanEnvelope"); err != nil {
		return ins, err
	}

	if ins.PanSustain, err = readUint8(p, "PanSustain"); err != nil {
		return ins, err
	}

	loop, err = p.numbers("PanLoop", 2, 0, 255)
	if err != nil {
		return ins, err
	}

	ins.PanLoopStart, ins.PanLoopEnd = uint8(loop[0]), uint8(loop[1])

	if ins.PanFlags, err = readUint8(p, "PanFlags"); err != nil {
		return ins, err
	}

	vibrato, err := p.numbers("Vibrato", 4, 0, 255)
	if err != nil {
		return ins, err
	}

	ins.VibratoType, ins.VibratoSweep = uint8(vibrato[0]), uint8(vibrato[1])
	ins.VibratoDepth, ins.VibratoRate = uint8(vibrato[2]), uint8(vibrato[3])

	fadeout, err := p.intField("VolumeFadeout", 0, 65535)
	if err != nil {
		return ins, err
	}

	ins.VolumeFadeout = uint16(fadeout)

	numSamples, err := p.intField("Samples", 0, maxSamples)
	if err != nil {
		return ins, err
	}

	for i := range numSamples {
		blank, err := p.readLine()
		if err != nil || blank != "" {
			return ins, errors.New("expected blank line before sample")
		}

		s, err := parseSample(p, i)
		if err != nil {
			return ins, err
		}

		ins.Samples = append(ins.Samples, s)
	}

	return ins, nil
}

// readUint8 reads a "Name: value" line holding one byte
func readUint8(p *parser, name string) (uint8, error) {
	v, err := p.intField(name, 0, 255)

	return uint8(v), err
}

func parseInstruments(p *parser, tf *TextFormat) error {
	for i := range int(tf.NumInstruments) {
		blank, err := p.readLine()
		if err != nil || blank != "" {
			return errors.New("expected blank line before instrument")
		}

		ins, err := parseInstrument(p, i)
		if err != nil {
			return err
		}

		tf.Instruments = append(tf.Instruments, ins)
	}

	return nil
}

func parseOrder(p *parser, tf *TextFormat) error {
	order, err := p.field("Order")
	if err != nil {
		return err
	}

	for _, s := range strings.Fields(order) {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 || n > 255 {
			return fmt.Errorf("invalid order value: %q", s)
		}

		tf.OrderTable = append(tf.OrderTable, uint8(n))
	}

	if len(tf.OrderTable) > maxOrder {
		return fmt.Errorf("order table too long: %d entries", len(tf.OrderTable))
	}

	return nil
}

func parseText(p *parser) (*TextFormat, error) {
	tf := &TextFormat{}

	name, err := p.field("Name")
	if err != nil {
		return nil, err
	}

	if err := copyName(tf.Name[:], name); err != nil {
		return nil, err
	}

	numChannels, err := p.intField("NumChannels", 1, maxChannels)
	if err != nil {
		return nil, err
	}

	tf.NumChannels = uint16(numChannels)

	numPatterns, err := p.intField("NumPatterns", 0, maxPatterns)
	if err != nil {
		return nil, err
	}

	tf.NumPatterns = uint16(numPatterns)

	numInstruments, err := p.intField("NumInstruments", 0, maxInstruments)
	if err != nil {
		return nil, err
	}

	tf.NumInstruments = uint16(numInstruments)

	speed, err := p.intField("Speed", 0, 65535)
	if err != nil {
		return nil, err
	}

	tf.Speed = uint16(speed)

	tempo, err := p.intField("Tempo", 0, 65535)
	if err != nil {
		return nil, err
	}

	tf.Tempo = uint16(tempo)

	linear, err := p.intField("LinearSlides", 0, 1)
	if err != nil {
		return nil, err
	}

	tf.LinearSlides = linear == 1

	if err := parseOrder(p, tf); err != nil {
		return nil, err
	}

	for _, o := range tf.OrderTable {
		if o != orderSkip && o != orderEnd && int(o) >= numPatterns {
			return nil, fmt.Errorf("order refers to missing pattern %d", o)
		}
	}

	if err := parsePatterns(p, tf); err != nil {
		return nil, err
	}

	if err := parseInstruments(p, tf); err != nil {
		return nil, err
	}

	return tf, nil
}

// ReadText reads a text-format module from r
func ReadText(r io.Reader) (*TextFormat, error) {
	p := &parser{r: bufio.NewReader(r)}

	tf, err := parseText(p)
	if err != nil {
		return nil, fmt.Errorf("line %d: %w", p.line, err)
	}

	return tf, nil
}
