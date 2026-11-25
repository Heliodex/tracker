package loadtext

import (
	"errors"
	"fmt"

	"github.com/Heliodex/tracker/types"
)

const (
	// xmHeaderSize is the value of the header size field for XM version 1.04 headers
	xmHeaderSize = 276
	// xmInstrumentSize is the instrument header size that covers every field, including the sample map
	xmInstrumentSize = 263
	// xmSampleHeaderSize is the size of one sample header
	xmSampleHeaderSize = 40
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

// sampleToXM converts a sample to its XM header, with loop points converted from frames to bytes
func sampleToXM(s *Sample) (types.SampleHeader, error) {
	var bytesPerFrame uint32 = 1

	flags := types.SampleFlags(s.LoopMode)

	if s.Sixteen {
		bytesPerFrame = 2
		flags |= types.SampleFlag16Bit

		if len(s.Data)%2 != 0 {
			return types.SampleHeader{}, errors.New("16-bit sample data has an odd length")
		}
	}

	return types.SampleHeader{
		SampleHeader1: types.SampleHeader1{
			Length: uint32(len(s.Data)),
			LoopStart: s.LoopStart * bytesPerFrame,
			LoopLength: s.LoopLength * bytesPerFrame,
			Volume: s.Volume,
			Finetune: s.Finetune,
			Flags: flags,
			Panning: s.Panning,
			RelativeNoteNumber: s.RelativeNote,
			Name: s.Name,
		},
		SampleData: s.Data,
	}, nil
}

// instrumentToXM converts an instrument to its XM header and samples
func instrumentToXM(ins *Instrument) (types.InstrumentHeader, error) {
	h := types.InstrumentHeader{
		InstrumentHeader1: types.InstrumentHeader1{
			Size: xmInstrumentSize,
			Name: ins.Name,
			Type: ins.Type,
			SamplesCount: uint16(len(ins.Samples)),
		},
		SampleHeaderSize: xmSampleHeaderSize,
		SampleNumber: ins.Keymap,
		VolPoints: uint8(len(ins.VolPoints)),
		PanPoints: uint8(len(ins.PanPoints)),
		VolSustainPoint: ins.VolSustain,
		VolLoopStartPoint: ins.VolLoopStart,
		VolLoopEndPoint: ins.VolLoopEnd,
		PanSustainPoint: ins.PanSustain,
		PanLoopStartPoint: ins.PanLoopStart,
		PanLoopEndPoint: ins.PanLoopEnd,
		VolFlags: types.EnvelopeFlags(ins.VolFlags),
		PanFlags: types.EnvelopeFlags(ins.PanFlags),
		VibratoType: ins.VibratoType,
		VibratoSweep: ins.VibratoSweep,
		VibratoDepth: ins.VibratoDepth,
		VibratoRate: ins.VibratoRate,
		VolumeFadeout: ins.VolumeFadeout,
	}

	copy(h.VolEnv[:], ins.VolPoints)
	copy(h.PanEnv[:], ins.PanPoints)

	for i := range ins.Samples {
		sh, err := sampleToXM(&ins.Samples[i])
		if err != nil {
			return h, fmt.Errorf("sample %d: %w", i, err)
		}

		h.Samples = append(h.Samples, sh)
	}

	return h, nil
}

// ToFile converts the text format into an XM file representation, ready for save.Write.
// Sample data must already be set on each sample, see Sample.Data.
func (tf *TextFormat) ToFile() (*types.File, error) {
	if int(tf.NumPatterns) != len(tf.Patterns) {
		return nil, fmt.Errorf("NumPatterns is %d but %d patterns are present", tf.NumPatterns, len(tf.Patterns))
	}

	if int(tf.NumInstruments) != len(tf.Instruments) {
		return nil, fmt.Errorf("NumInstruments is %d but %d instruments are present", tf.NumInstruments, len(tf.Instruments))
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
	h.NumInstruments = tf.NumInstruments

	if tf.LinearSlides {
		h.Flags = linearSlides
	}

	h.DefaultSpeed = tf.Speed
	h.DefaultTempo = tf.Tempo
	copy(h.OrderTable[:], tf.OrderTable)

	for _, pattern := range tf.Patterns {
		rows := make([]types.PatternRow, len(pattern))

		for r, row := range pattern {
			cells := make(types.PatternRow, len(row))

			for c, cd := range row {
				cells[c] = types.ChannelData{
					ChannelFlags: xmCellFlags(cd),
					Note: types.Note(cd.Note),
					Instrument: cd.Instrument,
					Volume: cd.Volume,
					Effect: cd.Effect,
					EffectParameter: cd.EffectParameter,
				}
			}

			rows[r] = cells
		}

		f.Patterns = append(f.Patterns, types.Pattern{
			PatternFileFormat: types.PatternFileFormat{
				Header: types.PatternHeader{
					PatternHeaderLength: 9,
					PackingType: 0,
					NumRows: uint16(len(pattern)),
				},
			},
			Data: rows,
		})
	}

	for i := range tf.Instruments {
		ih, err := instrumentToXM(&tf.Instruments[i])
		if err != nil {
			return nil, fmt.Errorf("instrument %d: %w", i, err)
		}

		f.Instruments = append(f.Instruments, ih)
	}

	return f, nil
}

// sampleFromXM converts an XM sample header to a sample, with loop points converted from bytes to frames
func sampleFromXM(sh *types.SampleHeader) (Sample, error) {
	if sh.Flags&types.SampleFlagStereo != 0 {
		return Sample{}, errors.New("stereo samples are not supported")
	}

	var bytesPerFrame uint32 = 1

	sixteen := sh.Flags&types.SampleFlag16Bit != 0
	if sixteen {
		bytesPerFrame = 2
	}

	if sh.LoopStart%bytesPerFrame != 0 || sh.LoopLength%bytesPerFrame != 0 {
		return Sample{}, errors.New("loop point is not aligned to a sample frame")
	}

	return Sample{
		Name: sh.Name,
		Volume: sh.Volume,
		Finetune: sh.Finetune,
		Panning: sh.Panning,
		RelativeNote: sh.RelativeNoteNumber,
		LoopMode: uint8(sh.Flags & 0x03),
		LoopStart: sh.LoopStart / bytesPerFrame,
		LoopLength: sh.LoopLength / bytesPerFrame,
		Sixteen: sixteen,
		Data: sh.SampleData,
	}, nil
}

// instrumentFromXM converts an XM instrument header to an instrument
func instrumentFromXM(ih *types.InstrumentHeader) (Instrument, error) {
	if ih.VolPoints > maxEnvPoints || ih.PanPoints > maxEnvPoints {
		return Instrument{}, fmt.Errorf("more than %d envelope points", maxEnvPoints)
	}

	ins := Instrument{
		Name: ih.Name,
		Type: ih.Type,
		Keymap: ih.SampleNumber,
		VolPoints: append([]types.EnvPoint(nil), ih.VolEnv[:ih.VolPoints]...),
		PanPoints: append([]types.EnvPoint(nil), ih.PanEnv[:ih.PanPoints]...),
		VolSustain: ih.VolSustainPoint,
		VolLoopStart: ih.VolLoopStartPoint,
		VolLoopEnd: ih.VolLoopEndPoint,
		VolFlags: uint8(ih.VolFlags),
		PanSustain: ih.PanSustainPoint,
		PanLoopStart: ih.PanLoopStartPoint,
		PanLoopEnd: ih.PanLoopEndPoint,
		PanFlags: uint8(ih.PanFlags),
		VibratoType: ih.VibratoType,
		VibratoSweep: ih.VibratoSweep,
		VibratoDepth: ih.VibratoDepth,
		VibratoRate: ih.VibratoRate,
		VolumeFadeout: ih.VolumeFadeout,
	}

	for i := range ih.Samples {
		s, err := sampleFromXM(&ih.Samples[i])
		if err != nil {
			return ins, fmt.Errorf("sample %d: %w", i, err)
		}

		ins.Samples = append(ins.Samples, s)
	}

	return ins, nil
}

// FromFile converts an XM file representation into the text format.
// Sample data is copied into each sample, ready for writing to WAV files.
func FromFile(f *types.File) (*TextFormat, error) {
	h := &f.Head

	if h.SongLength > 256 {
		return nil, fmt.Errorf("song length out of range: %d", h.SongLength)
	}

	tf := &TextFormat{
		Name: h.Name,
		NumChannels: h.NumChannels,
		NumPatterns: uint16(len(f.Patterns)),
		NumInstruments: uint16(len(f.Instruments)),
		Speed: h.DefaultSpeed,
		Tempo: h.DefaultTempo,
		LinearSlides: h.Flags&linearSlides != 0,
		OrderTable: append([]uint8(nil), h.OrderTable[:h.SongLength]...),
	}

	for _, p := range f.Patterns {
		pattern := make(Pattern, len(p.Data))

		for r, row := range p.Data {
			cells := make([]ChannelData, len(row))

			for c, cd := range row {
				cells[c] = ChannelData{
					Note: uint8(cd.Note),
					Instrument: cd.Instrument,
					Volume: cd.Volume,
					Effect: cd.Effect,
					EffectParameter: cd.EffectParameter,
				}
			}

			pattern[r] = cells
		}

		tf.Patterns = append(tf.Patterns, pattern)
	}

	for i := range f.Instruments {
		ins, err := instrumentFromXM(&f.Instruments[i])
		if err != nil {
			return nil, fmt.Errorf("instrument %d: %w", i, err)
		}

		tf.Instruments = append(tf.Instruments, ins)
	}

	return tf, nil
}
