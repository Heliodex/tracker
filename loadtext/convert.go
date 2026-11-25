package loadtext

import (
	"errors"
	"fmt"

	"github.com/Heliodex/tracker/types"
)

const (
	// xmHeaderSize is the value of the header size field for XM version 1.04 headers
	xmHeaderSize = 276
	// defaults for the header fields the text format does not hold
	defaultSpeed = 6
	defaultTempo = 125
	// linearSlides selects the linear frequency table, which is the default for XM modules
	linearSlides = 1
)

// xmCellFlags returns the packed-data flags for a cell, so only the columns in use are stored
func xmCellFlags(c ChannelData) types.ChannelFlags {
	f := types.ChannelFlagValid

	if c.Note != 0 {
		f |= types.ChannelFlagHasNote
	}

	if c.Instrument != 0 {
		f |= types.ChannelFlagHasInstrument
	}

	if c.Volume != 0 {
		f |= types.ChannelFlagHasVolume
	}

	if c.Effect != 0 {
		f |= types.ChannelFlagHasEffect
	}

	if c.EffectParameter != 0 {
		f |= types.ChannelFlagHasEffectParameter
	}

	return f
}

// ToFile converts the text format into an XM file representation, ready for save.Write.
// The text format holds no instruments or samples, so the file has none.
func (tf *TextFormat) ToFile() (*types.File, error) {
	if int(tf.NumPatterns) != len(tf.Patterns) {
		return nil, fmt.Errorf("NumPatterns is %d but %d patterns are present", tf.NumPatterns, len(tf.Patterns))
	}

	if len(tf.OrderTable) > 256 {
		return nil, errors.New("order table too long")
	}

	f := &types.File{}
	h := &f.Head

	copy(h.IDText[:], "Extended Module: ")
	h.Name = tf.Name
	h.Reserved1A = 0x1A
	copy(h.TrackerName[:], "tracker")
	h.VersionNumber = 0x0104
	h.HeaderSize = xmHeaderSize
	h.SongLength = uint16(len(tf.OrderTable))
	h.NumChannels = tf.NumChannels
	h.NumPatterns = tf.NumPatterns
	h.Flags = linearSlides
	h.DefaultSpeed = defaultSpeed
	h.DefaultTempo = defaultTempo
	copy(h.OrderTable[:], tf.OrderTable)

	for _, pattern := range tf.Patterns {
		rows := make([]types.PatternRow, len(pattern))

		for r, row := range pattern {
			cells := make(types.PatternRow, len(row))

			for c, cd := range row {
				cells[c] = types.ChannelData{
					ChannelFlags:    xmCellFlags(cd),
					Note:            types.Note(cd.Note),
					Instrument:      cd.Instrument,
					Volume:          cd.Volume,
					Effect:          cd.Effect,
					EffectParameter: cd.EffectParameter,
				}
			}

			rows[r] = cells
		}

		f.Patterns = append(f.Patterns, types.Pattern{
			PatternFileFormat: types.PatternFileFormat{
				Header: types.PatternHeader{
					PatternHeaderLength: 9,
					PackingType:         0,
					NumRows:             uint16(len(pattern)),
				},
			},
			Data: rows,
		})
	}

	return f, nil
}

// FromFile converts an XM file representation into the text format.
// Instruments are dropped, as the text format does not hold them.
func FromFile(f *types.File) (*TextFormat, error) {
	h := &f.Head

	if h.SongLength > 256 {
		return nil, fmt.Errorf("song length out of range: %d", h.SongLength)
	}

	tf := &TextFormat{
		Name:        h.Name,
		NumChannels: h.NumChannels,
		NumPatterns: uint16(len(f.Patterns)),
		OrderTable:  append([]uint8(nil), h.OrderTable[:h.SongLength]...),
	}

	for _, p := range f.Patterns {
		pattern := make(Pattern, len(p.Data))

		for r, row := range p.Data {
			cells := make([]ChannelData, len(row))

			for c, cd := range row {
				cells[c] = ChannelData{
					Note:            uint8(cd.Note),
					Instrument:      cd.Instrument,
					Volume:          cd.Volume,
					Effect:          cd.Effect,
					EffectParameter: cd.EffectParameter,
				}
			}

			pattern[r] = cells
		}

		tf.Patterns = append(tf.Patterns, pattern)
	}

	return tf, nil
}
