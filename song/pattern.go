package song

import (
	"fmt"

	"github.com/Heliodex/tracker/loadtext"
	"github.com/Heliodex/tracker/types"
)

// Pattern is a grid of rows, by channels
type Pattern struct {
	song *Song
	index int
	rows loadtext.Pattern
}

// Opt changes a cell as it is written. Options are applied after the note and instrument.
// An Opt returns an error when its value is out of range, which the song reports
type Opt func(c *loadtext.ChannelData) error

// Vol sets the volume column, from 0 to 255. The volume column takes raw values, where 0x10 to 0x50 is volume 0 to 64
func Vol(v int) Opt {
	return func(c *loadtext.ChannelData) error {
		if v < 0 || v > 255 {
			return fmt.Errorf("volume column must be between 0 and 255, got %d", v)
		}

		c.Volume = uint8(v)

		return nil
	}
}

// Pan sets the panning, from 0 (left) to 255 (right), with the 8xx effect
func Pan(v int) Opt {
	return func(c *loadtext.ChannelData) error {
		if v < 0 || v > 255 {
			return fmt.Errorf("pan must be between 0 and 255, got %d", v)
		}

		c.Effect = 0x08
		c.EffectParameter = uint8(v)

		return nil
	}
}

// Vibrato sets vibrato with the 4xy effect, with a speed and depth from 0 to 15
func Vibrato(speed, depth int) Opt {
	return func(c *loadtext.ChannelData) error {
		if speed < 0 || speed > 15 || depth < 0 || depth > 15 {
			return fmt.Errorf("vibrato speed and depth must be between 0 and 15, got %d and %d", speed, depth)
		}

		c.Effect = 0x04
		c.EffectParameter = uint8(speed<<4 | depth)

		return nil
	}
}

// Hit plays note on channel ch at row row, replacing whatever the cell held before.
// note is a tracker note name such as "C-4", or "###" for note off.
// ins is the instrument to play, or nil for none
func (p *Pattern) Hit(row, ch int, note string, ins *Instrument, opts ...Opt) {
	if row < 0 || row >= len(p.rows) || ch < 0 || ch >= p.song.channels {
		p.song.fail("pattern %d: row %d channel %d is out of range", p.index, row, ch)
		return
	}

	n, err := types.ParseNote(note)
	if err != nil {
		p.song.fail("pattern %d row %d channel %d: %v", p.index, row, ch, err)
		return
	}

	c := &p.rows[row][ch]
	*c = loadtext.ChannelData{Note: uint8(n)}

	if ins != nil {
		if ins.song != p.song {
			p.song.fail("pattern %d: instrument %q belongs to another song", p.index, ins.Name)
			return
		}

		c.Instrument = uint8(ins.id + 1)
	}

	for _, o := range opts {
		if err := o(c); err != nil {
			p.song.fail("pattern %d row %d channel %d: %v", p.index, row, ch, err)
		}
	}
}

// Line plays notes on channel ch, one every step rows, starting at row start
func (p *Pattern) Line(ch, start, step int, notes []string, ins *Instrument, opts ...Opt) {
	for i, note := range notes {
		p.Hit(start+i*step, ch, note, ins, opts...)
	}
}

// Rows plays the same note at each of the given rows
func (p *Pattern) Rows(ch int, rows []int, note string, ins *Instrument, opts ...Opt) {
	for _, r := range rows {
		p.Hit(r, ch, note, ins, opts...)
	}
}
