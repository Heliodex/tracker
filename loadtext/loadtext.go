// Package loadtext implements a text-based format for tracker music.
package loadtext

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ChannelData struct {
	Note,
	Instrument,
	Volume,
	Effect,
	EffectParameter uint8
}

type Pattern [][]ChannelData

type TextFormat struct {
	Name        [20]uint8
	NumPatterns uint16
	OrderTable  []uint8
	Patterns    []Pattern
}

type Parser struct {
	reader io.Reader
	offset int
}

func (p *Parser) SkipString(s string) (ok bool) {
	ls := len(s)
	toRead := make([]byte, ls)
	if _, err := p.reader.Read(toRead); err != nil {
		return
	}

	if !bytes.Equal(toRead, []byte(s)) {
		return
	}

	p.offset += ls
	return true
}

func (p *Parser) ReadLine() (line string, err error) {
	var buf bytes.Buffer
	for {
		var b [1]byte
		if _, err = p.reader.Read(b[:]); err != nil {
			return "", err
		}

		if b[0] == '\n' {
			break
		}

		buf.WriteByte(b[0])
		p.offset++
	}

	return buf.String(), nil
}

func parsePatternRows(p *Parser, numRows int) (Pattern, error) {
	rows := make(Pattern, numRows)
	for i := range numRows {
		if !p.SkipString("Row ") {
			return nil, errors.New("expected 'Row '")
		}

		
	}

	return
}

func parsePatterns(p *Parser, tf *TextFormat) error {
	for len(tf.OrderTable) > len(tf.Patterns) {
		if !p.SkipString("\nPattern ") {
			return errors.New("expected newline and 'Pattern '")
		}

		patternIndexLine, err := p.ReadLine()
		if err != nil {
			return fmt.Errorf("failed to read pattern index: %w", err)
		}

		patternIndex, err := strconv.Atoi(patternIndexLine)
		if err != nil {
			return fmt.Errorf("invalid pattern index: %s", patternIndexLine)
		}

		if patternIndex != len(tf.Patterns) {
			return fmt.Errorf("pattern index mismatch: expected %d, got %d", len(tf.Patterns), patternIndex)
		}

		if !p.SkipString("Rows: ") {
			return errors.New("expected 'Rows: '")
		}

		rowsLine, err := p.ReadLine()
		if err != nil {
			return fmt.Errorf("failed to read number of rows: %w", err)
		}

		numRows, err := strconv.Atoi(rowsLine)
		if err != nil {
			return fmt.Errorf("invalid number of rows: %s", rowsLine)
		}

		rows, err := parsePatternRows(p, numRows)
		if err != nil {
			return fmt.Errorf("failed to parse pattern rows: %w", err)
		}

		tf.Patterns = append(tf.Patterns, rows)
	}

	return nil
}

func parseName(p *Parser, tf *TextFormat) error {
	if !p.SkipString("Name: ") {
		return errors.New("expected 'Name: '")
	}

	name, err := p.ReadLine()
	if err != nil {
		return fmt.Errorf("failed to read Name: %w", err)
	}
	if len(name) > 20 {
		return fmt.Errorf("Name too long: %s", name)
	}

	copy(tf.Name[:], name)
	return nil
}

func parseNumPatterns(p *Parser, tf *TextFormat) error {
	if !p.SkipString("NumPatterns: ") {
		return errors.New("expected 'NumPatterns: '")
	}

	numPatternsLine, err := p.ReadLine()
	if err != nil {
		return fmt.Errorf("failed to read NumPatterns: %w", err)
	}

	numPatterns, err := strconv.Atoi(numPatternsLine)
	if err != nil {
		return fmt.Errorf("invalid NumPatterns value: %s", numPatternsLine)
	}

	tf.NumPatterns = uint16(numPatterns)
	return nil
}

func parseOrderTable(p *Parser, tf *TextFormat) error {
	if !p.SkipString("Order: ") {
		return errors.New("expected 'Order: '")
	}

	orderLine, err := p.ReadLine()
	if err != nil {
		return fmt.Errorf("failed to read Order: %w", err)
	}

	order := strings.SplitSeq(orderLine, " ")
	for o := range order {
		i, err := strconv.Atoi(o)
		if err != nil {
			return fmt.Errorf("invalid order value: %s", o)
		}
		tf.OrderTable = append(tf.OrderTable, uint8(i))
	}

	return nil
}

func parseText(p *Parser, tf *TextFormat) (err error) {
	if err = parseName(p, tf); err != nil {
		return
	}
	if err = parseNumPatterns(p, tf); err != nil {
		return
	}
	if err = parseOrderTable(p, tf); err != nil {
		return
	}

	return parsePatterns(p, tf)
}

func ReadText(r io.Reader) (tf *TextFormat, err error) {
	p := &Parser{reader: r}
	tf = &TextFormat{}

	if err = parseText(p, tf); err != nil {
		return nil, fmt.Errorf("offset %d: %w", p.offset, err)
	}

	return
}
