// Package loadtext implements a text-based format for tracker music.
//
// A file looks like this:
//
//	Name: cool guy
//	NumChannels: 2
//	NumPatterns: 1
//	Order: 0
//
//	Pattern 0
//	Rows: 2
//	Row 000: C-4 01 .. .. .. | ... .. .. .. ..
//	Row 001: ### .. .. .. .. | A#5 02 .. .. ..
//
// Each channel cell is a note followed by four hex columns: instrument, volume, effect and effect parameter.
// "..." is an empty note, "###" is note off, and ".." is a zero value.
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

	// TextFormat is the text representation of a module
	TextFormat struct {
		Name [20]uint8
		NumPatterns, NumChannels uint16
		OrderTable []uint8
		Patterns []Pattern
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

	// Order entries 254 and 255 are skip and end markers in XM, not patterns
	orderSkip = 254
)

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

// readLine returns the next line without its line ending
func (p *parser) readLine() (string, error) {
	line, err := p.r.ReadString('\n')
	if err != nil && (err != io.EOF || line == "") {
		return "", err
	}

	p.line++

	return strings.TrimRight(line, "\r\n"), nil
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

func parseName(p *parser, tf *TextFormat) error {
	name, err := p.field("Name")
	if err != nil {
		return err
	}

	if len(name) > len(tf.Name) {
		return fmt.Errorf("name too long: %q", name)
	}

	copy(tf.Name[:], name)

	return nil
}

func parseOrder(p *parser, tf *TextFormat) error {
	order, err := p.field("Order")
	if err != nil {
		return err
	}

	for s := range strings.FieldsSeq(order) {
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

// intField reads a "Name: value" line holding an integer between min and max inclusive
func (p *parser) intField(name string, min, max int) (int, error) {
	v, err := p.field(name)
	if err != nil {
		return 0, err
	}

	n, err := strconv.Atoi(v)
	if err != nil || n < min || n > max {
		return 0, fmt.Errorf("invalid %s value: %q", name, v)
	}

	return n, nil
}

func parsePattern(p *parser, index, numChannels int) (Pattern, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, fmt.Errorf("expected 'Pattern %d': %w", index, err)
	}

	indexStr, ok := strings.CutPrefix(line, "Pattern ")
	if n, err := strconv.Atoi(indexStr); !ok || err != nil || n != index {
		return nil, fmt.Errorf("expected 'Pattern %d', got %q", index, line)
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

func parseText(p *parser) (*TextFormat, error) {
	tf := &TextFormat{}

	if err := parseName(p, tf); err != nil {
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

	if err := parseOrder(p, tf); err != nil {
		return nil, err
	}

	for _, o := range tf.OrderTable {
		if o != orderSkip && o != 255 && int(o) >= numPatterns {
			return nil, fmt.Errorf("order refers to missing pattern %d", o)
		}
	}

	if err := parsePatterns(p, tf); err != nil {
		return nil, err
	}

	return tf, nil
}

// ReadText reads a text-format module from r
func ReadText(r io.Reader) (*TextFormat, error) {
	p := &parser{
		r: bufio.NewReader(r),
	}

	tf, err := parseText(p)
	if err != nil {
		return nil, fmt.Errorf("line %d: %w", p.line, err)
	}

	return tf, nil
}
