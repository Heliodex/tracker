// Package loadtext implements a text-based format for tracker music.
package loadtext

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type TextFormat struct {
	Name [20]uint8
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

func parseText(parser *Parser, tf *TextFormat) error {
	if !parser.SkipString("Name: ") {
		return errors.New("expected 'Name: '")
	}

	name, err := parser.ReadLine()
	if err != nil {
		return fmt.Errorf("failed to read Name: %w", err)
	}
	if len(name) > 20 {
		return fmt.Errorf("Name too long: %s", name)
	}

	copy(tf.Name[:], name)
	return nil
}

func ReadText(r io.Reader) (tf *TextFormat, err error) {
	parser := &Parser{reader: r}
	tf = &TextFormat{}

	if err = parseText(parser, tf); err != nil {
		return nil, fmt.Errorf("offset %d: %w", parser.offset, err)
	}

	return
}
