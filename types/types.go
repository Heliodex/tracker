package types

import (
	"bytes"
	"encoding/binary"
)

// HeaderFlags is the set of flags for an XM header
type HeaderFlags uint16

/*
const (
	// HeaderFlagLinearSlides activates the linear frequency table (off = Amiga frequency table)
	HeaderFlagLinearSlides = HeaderFlags(0x0001)
	// HeaderFlagExtendedFilterRange activates the extended filter range
	HeaderFlagExtendedFilterRange = HeaderFlags(0x1000)
)

// IsLinearSlides returns true if the song plays with linear note slides (or if false, with Amiga note slides)
func (f HeaderFlags) IsLinearSlides() bool {
	return (f & HeaderFlagLinearSlides) != 0
}

// IsExtendedFilterRange returns true if the song has extended filter ranges enabled
func (f HeaderFlags) IsExtendedFilterRange() bool {
	return (f & HeaderFlagExtendedFilterRange) != 0
}
*/

type ModuleHeader1 struct {
	IDText        [17]uint8
	Name          [20]uint8
	Reserved1A    uint8
	TrackerName   [20]uint8
	VersionNumber uint16
}

// ModuleHeader is a representation of the XM file header
type ModuleHeader struct {
	ModuleHeader1
	HeaderSize uint32
	SongLength,
	RestartPosition,
	NumChannels,
	NumPatterns,
	NumInstruments uint16
	Flags HeaderFlags
	DefaultSpeed,
	DefaultTempo uint16
	OrderTable [256]uint8
}

// ChannelFlags describes what is valid in a channel
type ChannelFlags uint8

const (
	// ChannelFlagHasNote signifies that the channel data includes a note
	ChannelFlagHasNote = ChannelFlags(0x01)
	// ChannelFlagHasInstrument signifies that the channel data includes an instrument
	ChannelFlagHasInstrument = ChannelFlags(0x02)
	// ChannelFlagHasVolume signifies that the channel data includes a volume
	ChannelFlagHasVolume = ChannelFlags(0x04)
	// ChannelFlagHasEffect signifies that the channel data includes an effect
	ChannelFlagHasEffect = ChannelFlags(0x08)
	// ChannelFlagHasEffectParameter signifies that the channel data includes an effect parameter
	ChannelFlagHasEffectParameter = ChannelFlags(0x10)
	// ChannelFlagValid signifies that the channel flags are valid
	ChannelFlagValid = ChannelFlags(0x80)

	// ChannelFlagsAll is all channel flags at once
	ChannelFlagsAll = ChannelFlags(0xFF)
)

// HasNote returns true when the channel includes note data
func (f ChannelFlags) HasNote() bool {
	return (f & ChannelFlagHasNote) != 0
}

// HasInstrument returns true when the channel includes instrument data
func (f ChannelFlags) HasInstrument() bool {
	return (f & ChannelFlagHasInstrument) != 0
}

// HasVolume returns true when the channel includes volume data
func (f ChannelFlags) HasVolume() bool {
	return (f & ChannelFlagHasVolume) != 0
}

// HasEffect returns true when the channel includes effect data
func (f ChannelFlags) HasEffect() bool {
	return (f & ChannelFlagHasEffect) != 0
}

// HasEffectParameter returns true when the channel includes effect parameter data
func (f ChannelFlags) HasEffectParameter() bool {
	return (f & ChannelFlagHasEffectParameter) != 0
}

// IsValid returns true when the channel flags are valid
func (f ChannelFlags) IsValid() bool {
	return (f & ChannelFlagValid) != 0
}

// ChannelData is the XM unpacked pattern channel data definition
type ChannelData struct {
	ChannelFlags
	Note,
	Instrument,
	Volume,
	Effect,
	EffectParameter uint8
}

// PatternHeader is the XM packed pattern header definition
type PatternHeader struct {
	PatternHeaderLength   uint32
	PackingType           uint8
	NumRows               uint16
	PackedPatternDataSize uint16
}

// PatternFileFormat is the XM pattern definition in file format
type PatternFileFormat struct {
	Header PatternHeader
}

// PatternRow is the XM unpacked pattern channel data list for a single pattern row
type PatternRow []ChannelData

// Pattern is an XM internal file representation and converted/unpacked pattern set
type Pattern struct {
	PatternFileFormat

	Data []PatternRow
}

func (p *Pattern) Unpack(numChannels int, pd []byte) (err error) {
	numRows := int(p.Header.NumRows)

	if len(pd) == 0 {
		// empty pattern
		p.Data = make([]PatternRow, numRows)
		for i := range p.Data {
			p.Data[i] = make(PatternRow, numChannels)
		}
		return
	}

	// it's not empty, so let's unpack it!

	p.Data = make([]PatternRow, numRows)
	packed := bytes.NewReader(pd)
	for i := range p.Data {
		row := make(PatternRow, numChannels)
		p.Data[i] = row
		for c := range numChannels {
			ch := &row[c]
			if err = binary.Read(packed, binary.LittleEndian, &ch.ChannelFlags); err != nil {
				return
			}

			// is the first byte a bitfield instead of note?
			if ch.IsValid() {
				// it is!
				// note present?
				if ch.HasNote() {
					if err = binary.Read(packed, binary.LittleEndian, &ch.Note); err != nil {
						return
					}
				}
			} else {
				// it isn't... assume it's a note and that we have everything present
				ch.Note = uint8(ch.ChannelFlags)
				ch.ChannelFlags = ChannelFlagsAll
			}

			// instrument present?
			if ch.HasInstrument() {
				if err = binary.Read(packed, binary.LittleEndian, &ch.Instrument); err != nil {
					return
				}
			}

			// volume present?
			if ch.HasVolume() {
				if err = binary.Read(packed, binary.LittleEndian, &ch.Volume); err != nil {
					return
				}
			}

			// effect present?
			if ch.HasEffect() {
				if err = binary.Read(packed, binary.LittleEndian, &ch.Effect); err != nil {
					return
				}
			}

			// effect parameter present?
			if ch.HasEffectParameter() {
				if err = binary.Read(packed, binary.LittleEndian, &ch.EffectParameter); err != nil {
					return
				}
			}
		}
	}

	return
}

// Pack converts the unpacked pattern data back to packed format
func (p *Pattern) Pack(numChannels int) ([]byte, error) {
	var buf bytes.Buffer
	numRows := int(p.Header.NumRows)

	for i := range numRows {
		row := p.Data[i]
		for c := range numChannels {
			ch := &row[c]

			// write channel flags
			if err := binary.Write(&buf, binary.LittleEndian, ch.ChannelFlags); err != nil {
				return nil, err
			}

			// note present?
			if ch.HasNote() {
				if err := binary.Write(&buf, binary.LittleEndian, ch.Note); err != nil {
					return nil, err
				}
			}

			// instrument present?
			if ch.HasInstrument() {
				if err := binary.Write(&buf, binary.LittleEndian, ch.Instrument); err != nil {
					return nil, err
				}
			}

			// volume present?
			if ch.HasVolume() {
				if err := binary.Write(&buf, binary.LittleEndian, ch.Volume); err != nil {
					return nil, err
				}
			}

			// effect present?
			if ch.HasEffect() {
				if err := binary.Write(&buf, binary.LittleEndian, ch.Effect); err != nil {
					return nil, err
				}
			}

			// effect parameter present?
			if ch.HasEffectParameter() {
				if err := binary.Write(&buf, binary.LittleEndian, ch.EffectParameter); err != nil {
					return nil, err
				}
			}
		}
	}

	return buf.Bytes(), nil
}

// EnvPoint is a representation of an XM file envelope point
type EnvPoint struct {
	X, Y uint16
}

// EnvelopeFlags is a representation of the XM file instrument envelope flags (vol/pan)
type EnvelopeFlags uint8

// SampleFlags is a representation of the XM file sample flags
type SampleFlags uint8

const (
	// sampleFlagLoopModeMask is the mask to pull the loop mode from the sample flags
	sampleFlagLoopModeMask = SampleFlags(0x03)
	// SampleFlag16Bit designates that the sample is 16-bit
	SampleFlag16Bit = SampleFlags(0x10)
	// SampleFlagStereo designates that the sample is stereo
	SampleFlagStereo = SampleFlags(0x20)
)

type SampleHeader1 struct {
	Length,
	LoopStart,
	LoopLength uint32
	Volume             uint8
	Finetune           int8
	Flags              SampleFlags
	Panning            uint8
	RelativeNoteNumber int8
	ReservedP17        uint8
	Name               [22]uint8
}

// SampleHeader is a representation of the XM file sample header
type SampleHeader struct {
	SampleHeader1
	SampleData []uint8
}

// InstrumentHeader is a representation of the XM file instrument header
type InstrumentHeader struct {
	Size         uint32
	Name         [22]uint8
	Type         uint8
	SamplesCount uint16

	SampleHeaderSize uint32
	SampleNumber     [96]uint8
	VolEnv,
	PanEnv [12]EnvPoint

	VolPoints uint8
	PanPoints,
	VolSustainPoint,
	VolLoopStartPoint,
	VolLoopEndPoint,
	PanSustainPoint,
	PanLoopStartPoint,
	PanLoopEndPoint uint8
	VolFlags,
	PanFlags EnvelopeFlags
	VibratoType,
	VibratoSweep,
	VibratoDepth,
	VibratoRate uint8
	VolumeFadeout uint16
	ReservedP241  [11]uint16

	Samples []SampleHeader
}

// File is an XM internal file representation
type File struct {
	Head        ModuleHeader
	Patterns    []Pattern
	Instruments []InstrumentHeader
}
