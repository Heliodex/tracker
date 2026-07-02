package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
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

func indent(s string, level int) string {
	parts := strings.Split(s, "\n")
	for i, v := range parts {
		parts[i] = strings.Repeat("  ", level) + v
	}

	return strings.Join(parts, "\n")
}

type ModuleHeader1 struct {
	IDText        [17]uint8
	Name          [20]uint8
	Reserved1A    uint8
	TrackerName   [20]uint8
	VersionNumber uint16
	HeaderSize    uint32
	SongLength    uint16
}

func (m *ModuleHeader1) String() string {
	var b bytes.Buffer
	b.WriteString("IDText: ")
	b.WriteString(string(m.IDText[:]))
	b.WriteString("\nName: ")
	b.WriteString(string(m.Name[:]))
	b.WriteString("\nReserved1A: ")
	b.WriteString(string(m.Reserved1A))
	b.WriteString("\nTrackerName: ")
	b.WriteString(string(m.TrackerName[:]))
	b.WriteString("\nVersionNumber: ")
	b.WriteString(strconv.Itoa(int(m.VersionNumber)))
	b.WriteString("\nHeaderSize: ")
	b.WriteString(strconv.Itoa(int(m.HeaderSize)))
	b.WriteString("\nSongLength: ")
	b.WriteString(strconv.Itoa(int(m.SongLength)))
	return fmt.Sprintf("ModuleHeader1 {\n%s\n}", indent(b.String(), 1))
}

// ModuleHeader is a representation of the XM file header
type ModuleHeader struct {
	ModuleHeader1
	RestartPosition,
	NumChannels,
	NumPatterns,
	NumInstruments uint16
	Flags HeaderFlags
	DefaultSpeed,
	DefaultTempo uint16
	OrderTable [256]uint8
}

func (m *ModuleHeader) String() string {
	var b bytes.Buffer
	b.WriteString("ModuleHeader1: ")
	b.WriteString(m.ModuleHeader1.String())
	b.WriteString("\nRestartPosition: ")
	b.WriteString(strconv.Itoa(int(m.RestartPosition)))
	b.WriteString("\nNumChannels: ")
	b.WriteString(strconv.Itoa(int(m.NumChannels)))
	b.WriteString("\nNumPatterns: ")
	b.WriteString(strconv.Itoa(int(m.NumPatterns)))
	b.WriteString("\nNumInstruments: ")
	b.WriteString(strconv.Itoa(int(m.NumInstruments)))
	b.WriteString("\nFlags: ")
	b.WriteString(strconv.Itoa(int(m.Flags)))
	b.WriteString("\nDefaultSpeed: ")
	b.WriteString(strconv.Itoa(int(m.DefaultSpeed)))
	b.WriteString("\nDefaultTempo: ")
	b.WriteString(strconv.Itoa(int(m.DefaultTempo)))
	b.WriteString("\nOrderTable: ")
	for i := range 256 {
		b.WriteString(strconv.Itoa(int(m.OrderTable[i])))
		if i < 255 {
			b.WriteString(", ")
		}
	}
	return fmt.Sprintf("ModuleHeader {\n%s\n}", indent(b.String(), 1))
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

func (c *ChannelData) String() string {
	var b bytes.Buffer
	b.WriteString("ChannelFlags: ")
	b.WriteString(strconv.Itoa(int(c.ChannelFlags)))
	b.WriteString("\nNote: ")
	b.WriteString(strconv.Itoa(int(c.Note)))
	b.WriteString("\nInstrument: ")
	b.WriteString(strconv.Itoa(int(c.Instrument)))
	b.WriteString("\nVolume: ")
	b.WriteString(strconv.Itoa(int(c.Volume)))
	b.WriteString("\nEffect: ")
	b.WriteString(strconv.Itoa(int(c.Effect)))
	b.WriteString("\nEffectParameter: ")
	b.WriteString(strconv.Itoa(int(c.EffectParameter)))
	return fmt.Sprintf("ChannelData {\n%s\n}", indent(b.String(), 1))
}

// PatternHeader is the XM packed pattern header definition
type PatternHeader struct {
	PatternHeaderLength   uint32
	PackingType           uint8
	NumRows               uint16
	PackedPatternDataSize uint16
}

func (p *PatternHeader) String() string {
	var b bytes.Buffer
	b.WriteString("PatternHeaderLength: ")
	b.WriteString(strconv.Itoa(int(p.PatternHeaderLength)))
	b.WriteString("\nPackingType: ")
	b.WriteString(strconv.Itoa(int(p.PackingType)))
	b.WriteString("\nNumRows: ")
	b.WriteString(strconv.Itoa(int(p.NumRows)))
	b.WriteString("\nPackedPatternDataSize: ")
	b.WriteString(strconv.Itoa(int(p.PackedPatternDataSize)))
	return fmt.Sprintf("PatternHeader {\n%s\n}", indent(b.String(), 1))
}

// PatternFileFormat is the XM pattern definition in file format
type PatternFileFormat struct {
	Header PatternHeader
}

func (p *PatternFileFormat) String() string {
	var b bytes.Buffer
	b.WriteString("Header: ")
	b.WriteString(p.Header.String())
	return fmt.Sprintf("PatternFileFormat {\n%s\n}", indent(b.String(), 1))
}

// PatternRow is the XM unpacked pattern channel data list for a single pattern row
type PatternRow []ChannelData

func (p PatternRow) String() string {
	var b bytes.Buffer
	b.WriteString("[\n")
	for i := range p {
		b.WriteString(indent(p[i].String(), 1))
		if i < len(p)-1 {
			b.WriteString(",\n")
		}
	}
	b.WriteString("\n]")
	return fmt.Sprintf("PatternRow {\n%s\n}", indent(b.String(), 1))
}

// Pattern is an XM internal file representation and converted/unpacked pattern set
type Pattern struct {
	PatternFileFormat

	Data []PatternRow
}

func (p *Pattern) String() string {
	var b bytes.Buffer
	b.WriteString("PatternFileFormat: ")
	b.WriteString(p.PatternFileFormat.String())
	b.WriteString("\nData: [\n")
	for i := range p.Data {
		b.WriteString(indent(p.Data[i].String(), 1))
		if i < len(p.Data)-1 {
			b.WriteString(",\n")
		}
	}
	b.WriteString("\n]")
	return fmt.Sprintf("Pattern {\n%s\n}", indent(b.String(), 1))
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

func (e *EnvPoint) String() string {
	var b bytes.Buffer
	b.WriteString("X: ")
	b.WriteString(strconv.Itoa(int(e.X)))
	b.WriteString("\nY: ")
	b.WriteString(strconv.Itoa(int(e.Y)))
	return fmt.Sprintf("EnvPoint {\n%s\n}", indent(b.String(), 1))
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

func (s *SampleHeader1) String() string {
	var b bytes.Buffer
	b.WriteString("Length: ")
	b.WriteString(strconv.Itoa(int(s.Length)))
	b.WriteString("\nLoopStart: ")
	b.WriteString(strconv.Itoa(int(s.LoopStart)))
	b.WriteString("\nLoopLength: ")
	b.WriteString(strconv.Itoa(int(s.LoopLength)))
	b.WriteString("\nVolume: ")
	b.WriteString(strconv.Itoa(int(s.Volume)))
	b.WriteString("\nFinetune: ")
	b.WriteString(strconv.Itoa(int(s.Finetune)))
	b.WriteString("\nFlags: ")
	b.WriteString(strconv.Itoa(int(s.Flags)))
	b.WriteString("\nPanning: ")
	b.WriteString(strconv.Itoa(int(s.Panning)))
	b.WriteString("\nRelativeNoteNumber: ")
	b.WriteString(strconv.Itoa(int(s.RelativeNoteNumber)))
	b.WriteString("\nReservedP17: ")
	b.WriteString(strconv.Itoa(int(s.ReservedP17)))
	b.WriteString("\nName: ")
	b.WriteString(string(s.Name[:]))
	return fmt.Sprintf("SampleHeader1 {\n%s\n}", indent(b.String(), 1))
}

// SampleHeader is a representation of the XM file sample header
type SampleHeader struct {
	SampleHeader1
	SampleData []uint8
}

func (s *SampleHeader) String() string {
	var b bytes.Buffer
	b.WriteString("SampleHeader1: ")
	b.WriteString(s.SampleHeader1.String())
	b.WriteString("\nSampleData: [")
	for i := range s.SampleData {
		b.WriteString(strconv.Itoa(int(s.SampleData[i])))
		if i < len(s.SampleData)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("]")
	return fmt.Sprintf("SampleHeader {\n%s\n}", indent(b.String(), 1))
}

type InstrumentHeader1 struct {
	Size         uint32
	Name         [22]uint8
	Type         uint8
	SamplesCount uint16
}

func (i *InstrumentHeader1) String() string {
	var b bytes.Buffer
	b.WriteString("Size: ")
	b.WriteString(strconv.Itoa(int(i.Size)))
	b.WriteString("\nName: ")
	b.WriteString(string(i.Name[:]))
	b.WriteString("\nType: ")
	b.WriteString(strconv.Itoa(int(i.Type)))
	b.WriteString("\nSamplesCount: ")
	b.WriteString(strconv.Itoa(int(i.SamplesCount)))
	return fmt.Sprintf("InstrumentHeader1 {\n%s\n}", indent(b.String(), 1))
}

// InstrumentHeader is a representation of the XM file instrument header
type InstrumentHeader struct {
	InstrumentHeader1

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

func (i *InstrumentHeader) String() string {
	var b bytes.Buffer
	b.WriteString("InstrumentHeader1: ")
	b.WriteString(i.InstrumentHeader1.String())
	b.WriteString("\nSampleHeaderSize: ")
	b.WriteString(strconv.Itoa(int(i.SampleHeaderSize)))
	b.WriteString("\nSampleNumber: [")
	for j := range i.SampleNumber {
		b.WriteString(strconv.Itoa(int(i.SampleNumber[j])))
		if j < len(i.SampleNumber)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("]\nVolEnv: [")
	for j := range i.VolEnv {
		b.WriteString(i.VolEnv[j].String())
		if j < len(i.VolEnv)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("]\nPanEnv: [")
	for j := range i.PanEnv {
		b.WriteString(i.PanEnv[j].String())
		if j < len(i.PanEnv)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("]\nVolPoints: ")
	b.WriteString(strconv.Itoa(int(i.VolPoints)))
	b.WriteString("\nPanPoints: ")
	b.WriteString(strconv.Itoa(int(i.PanPoints)))
	b.WriteString("\nVolSustainPoint: ")
	b.WriteString(strconv.Itoa(int(i.VolSustainPoint)))
	b.WriteString("\nVolLoopStartPoint: ")
	b.WriteString(strconv.Itoa(int(i.VolLoopStartPoint)))
	b.WriteString("\nVolLoopEndPoint: ")
	b.WriteString(strconv.Itoa(int(i.VolLoopEndPoint)))
	b.WriteString("\nPanSustainPoint: ")
	b.WriteString(strconv.Itoa(int(i.PanSustainPoint)))
	b.WriteString("\nPanLoopStartPoint: ")
	b.WriteString(strconv.Itoa(int(i.PanLoopStartPoint)))
	b.WriteString("\nPanLoopEndPoint: ")
	b.WriteString(strconv.Itoa(int(i.PanLoopEndPoint)))
	b.WriteString("\nVolFlags: ")
	b.WriteString(strconv.Itoa(int(i.VolFlags)))
	b.WriteString("\nPanFlags: ")
	b.WriteString(strconv.Itoa(int(i.PanFlags)))
	b.WriteString("\nVibratoType: ")
	b.WriteString(strconv.Itoa(int(i.VibratoType)))
	b.WriteString("\nVibratoSweep: ")
	b.WriteString(strconv.Itoa(int(i.VibratoSweep)))
	b.WriteString("\nVibratoDepth: ")
	b.WriteString(strconv.Itoa(int(i.VibratoDepth)))
	b.WriteString("\nVibratoRate: ")
	b.WriteString(strconv.Itoa(int(i.VibratoRate)))
	b.WriteString("\nVolumeFadeout: ")
	b.WriteString(strconv.Itoa(int(i.VolumeFadeout)))
	b.WriteString("\nReservedP241: [")
	for j := range i.ReservedP241 {
		b.WriteString(strconv.Itoa(int(i.ReservedP241[j])))
		if j < len(i.ReservedP241)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("]\nSamples: [")
	for j := range i.Samples {
		b.WriteString(i.Samples[j].String())
		if j < len(i.Samples)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("]")
	return fmt.Sprintf("InstrumentHeader {\n%s\n}", indent(b.String(), 1))
}

// File is an XM internal file representation
type File struct {
	Head        ModuleHeader
	Patterns    []Pattern
	Instruments []InstrumentHeader
}

func (f *File) String() string {
	var b bytes.Buffer
	b.WriteString("Head: ")
	b.WriteString(f.Head.String())
	b.WriteString("\nPatterns: [")
	for i := range f.Patterns {
		b.WriteString(f.Patterns[i].String())
		if i < len(f.Patterns)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("]\nInstruments: [")
	for i := range f.Instruments {
		b.WriteString(f.Instruments[i].String())
		if i < len(f.Instruments)-1 {
			b.WriteString(", ")
		}
	}
	b.WriteString("]")
	return fmt.Sprintf("File {\n%s\n}", indent(b.String(), 1))
}
