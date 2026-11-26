package song

import (
	"github.com/Heliodex/tracker/loadtext"
	"github.com/Heliodex/tracker/types"
)

// Pattern is a grid of rows, by channels
type Pattern struct {
	song *Song
	index int
	rows loadtext.Pattern
}

// Opt changes a cell as it is written. Options are applied after the note and instrument
type Opt func(c *loadtext.ChannelData)

// Vol sets the volume column. The volume column takes raw values, where 0x10 to 0x50 is volume 0 to 64
func Vol(v int) Opt {
	return func(c *loadtext.ChannelData) {
		c.Volume = uint8(v)
	}
}

// Pan sets the panning, from 0 (left) to 255 (right), with the 8xx effect
func Pan(v int) Opt {
	return func(c *loadtext.ChannelData) {
		c.Effect = 0x08
		c.EffectParameter = uint8(v)
	}
}

// Vibrato sets vibrato with the 4xy effect, with a speed and depth from 0 to 15
func Vibrato(speed, depth int) Opt {
	return func(c *loadtext.ChannelData) {
		c.Effect = 0x04
		c.EffectParameter = uint8(speed<<4 | depth)
	}
}

// Hit plays note on channel ch at row row. note is a tracker note name such as "C-4", or "###" for note off.
// ins is the instrument to play, or nil to keep the instrument already playing
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
	c.Note = uint8(n)

	if ins != nil {
		if ins.song != p.song {
			p.song.fail("pattern %d: instrument %q belongs to another song", p.index, ins.Name)
			return
		}

		c.Instrument = uint8(ins.id + 1)
	}

	for _, o := range opts {
		o(c)
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
