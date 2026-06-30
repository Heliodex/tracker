package load

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// HeaderFlags is the set of flags for an XM header
type HeaderFlags uint16

// ModuleHeader is a representation of the XM file header
type ModuleHeader struct {
	IDText          [17]uint8
	Name            [20]uint8
	Reserved1A      uint8
	TrackerName     [20]uint8
	VersionNumber   uint16
	HeaderSize      uint32
	SongLength      uint16
	RestartPosition uint16
	NumChannels     uint16
	NumPatterns     uint16
	NumInstruments  uint16
	Flags           HeaderFlags
	DefaultSpeed    uint16
	DefaultTempo    uint16
	OrderTable      [256]uint8
}

func readHeaderPartial(r io.Reader) (*ModuleHeader, error) {
	xmh := ModuleHeader{}

	if err := binary.Read(r, binary.LittleEndian, &xmh.IDText); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &xmh.Name); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &xmh.Reserved1A); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &xmh.TrackerName); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &xmh.VersionNumber); err != nil {
		return nil, err
	}

	sz := uint32(0)
	if err := binary.Read(r, binary.LittleEndian, &xmh.HeaderSize); err != nil {
		return nil, err
	}
	sz += 4

	if err := binary.Read(r, binary.LittleEndian, &xmh.SongLength); err != nil {
		return nil, err
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return &xmh, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &xmh.RestartPosition); err != nil {
		return nil, err
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return &xmh, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &xmh.NumChannels); err != nil {
		return nil, err
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return &xmh, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &xmh.NumPatterns); err != nil {
		return nil, err
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return &xmh, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &xmh.NumInstruments); err != nil {
		return nil, err
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return &xmh, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &xmh.Flags); err != nil {
		return nil, err
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return &xmh, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &xmh.DefaultSpeed); err != nil {
		return nil, err
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return &xmh, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &xmh.DefaultTempo); err != nil {
		return nil, err
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return &xmh, nil
	}

	for i := range xmh.OrderTable {
		if err := binary.Read(r, binary.LittleEndian, &xmh.OrderTable[i]); err != nil {
			return nil, err
		}
		if sz++; sz >= xmh.HeaderSize {
			return &xmh, nil
		}
	}

	return &xmh, nil
}

func readHeader(r io.Reader) (*ModuleHeader, error) {
	xmh, err := readHeaderPartial(r)
	if err != nil {
		return nil, err
	}

	if xmh.NumChannels < 1 || xmh.NumChannels > 32 {
		return nil, errors.New("invalid number of channels - possibly corrupt file")
	}

	if xmh.NumPatterns > 256 {
		return nil, errors.New("invalid number of patterns - possibly corrupt file")
	}

	if xmh.NumInstruments > 128 {
		return nil, errors.New("invalid number of instruments - possibly corrupt file")
	}

	return xmh, nil
}

// PatternHeader is the XM packed pattern header definition
type PatternHeader struct {
	PatternHeaderLength   uint32
	PackingType           uint8
	NumRows               uint16
	PackedPatternDataSize uint16
}

func readPatternHeaderPartial(r io.Reader, fileVersion uint16) (*PatternHeader, error) {
	ph := PatternHeader{}

	sz := uint32(0)
	if err := binary.Read(r, binary.LittleEndian, &ph.PatternHeaderLength); err != nil {
		return nil, err
	}
	if sz += 4; sz >= ph.PatternHeaderLength {
		return &ph, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ph.PackingType); err != nil {
		return nil, err
	}
	if sz++; sz >= ph.PatternHeaderLength {
		return &ph, nil
	}

	if fileVersion == 0x0102 {
		var rowCount uint8
		if err := binary.Read(r, binary.LittleEndian, &rowCount); err != nil {
			return nil, err
		}

		ph.NumRows = uint16(rowCount) + 1
		if sz++; sz >= ph.PatternHeaderLength {
			return &ph, nil
		}

	} else {
		if err := binary.Read(r, binary.LittleEndian, &ph.NumRows); err != nil {
			return nil, err
		}
		if sz += 2; sz >= ph.PatternHeaderLength {
			return &ph, nil
		}
	}

	if err := binary.Read(r, binary.LittleEndian, &ph.PackedPatternDataSize); err != nil {
		return nil, err
	}
	if sz += 2; sz >= ph.PatternHeaderLength {
		return &ph, nil
	}

	return &ph, nil
}

func readPatternHeader(r io.Reader, fileVersion uint16) (*PatternHeader, error) {
	ph, err := readPatternHeaderPartial(r, fileVersion)
	if err != nil {
		return nil, err
	}

	//if ph.NumRows == 0 {
	//	ph.NumRows = 64
	//}

	if ph.PackingType != 0 {
		return nil, errors.New("unexpected pattern packing type - possibly corrupt file")
	}

	if ph.NumRows < 1 || ph.NumRows > 256 {
		return nil, errors.New("pattern row count out of range - possibly corrupt file")
	}

	return ph, nil
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
	Flags           ChannelFlags
	Note            uint8
	Instrument      uint8
	Volume          uint8
	Effect          uint8
	EffectParameter uint8
}

// HasNote returns true when the channel includes note data
func (f ChannelData) HasNote() bool {
	return f.Flags.HasNote()
}

// HasInstrument returns true when the channel includes instrument data
func (f ChannelData) HasInstrument() bool {
	return f.Flags.HasInstrument()
}

// HasVolume returns true when the channel includes volume data
func (f ChannelData) HasVolume() bool {
	return f.Flags.HasVolume()
}

// HasEffect returns true when the channel includes effect data
func (f ChannelData) HasEffect() bool {
	return f.Flags.HasEffect()
}

// HasEffectParameter returns true when the channel includes effect parameter data
func (f ChannelData) HasEffectParameter() bool {
	return f.Flags.HasEffectParameter()
}

// IsValid returns true when the channel flags are valid
func (f ChannelData) IsValid() bool {
	return f.Flags.IsValid()
}

// PatternFileFormat is the XM pattern definition in file format
type PatternFileFormat struct {
	Header     PatternHeader
	PackedData []byte
}

// PatternRow is the XM unpacked pattern channel data list for a single pattern row
type PatternRow []ChannelData

// Pattern is an XM internal file representation and converted/unpacked pattern set
type Pattern struct {
	PatternFileFormat

	Data []PatternRow
}

func (p *Pattern) unpack(numChannels int) error {
	numRows := int(p.Header.NumRows)

	if len(p.PackedData) == 0 {
		// empty pattern
		p.Data = make([]PatternRow, numRows)
		for i := range p.Data {
			p.Data[i] = make(PatternRow, numChannels)
		}
		return nil
	}

	// it's not empty, so let's unpack it!

	p.Data = make([]PatternRow, numRows)
	packed := bytes.NewReader(p.PackedData)
	for i := range p.Data {
		row := make(PatternRow, numChannels)
		p.Data[i] = row
		for c := 0; c < numChannels; c++ {
			ch := &row[c]
			if err := binary.Read(packed, binary.LittleEndian, &ch.Flags); err != nil {
				return err
			}

			// is the first byte a bitfield instead of note?
			if ch.IsValid() {
				// it is!
				// note present?
				if ch.HasNote() {
					if err := binary.Read(packed, binary.LittleEndian, &ch.Note); err != nil {
						return err
					}
				}
			} else {
				// it isn't... assume it's a note and that we have everything present
				ch.Note = uint8(ch.Flags)
				ch.Flags = ChannelFlagsAll
			}

			// instrument present?
			if ch.HasInstrument() {
				if err := binary.Read(packed, binary.LittleEndian, &ch.Instrument); err != nil {
					return err
				}
			}

			// volume present?
			if ch.HasVolume() {
				if err := binary.Read(packed, binary.LittleEndian, &ch.Volume); err != nil {
					return err
				}
			}

			// effect present?
			if ch.HasEffect() {
				if err := binary.Read(packed, binary.LittleEndian, &ch.Effect); err != nil {
					return err
				}
			}

			// effect parameter present?
			if ch.HasEffectParameter() {
				if err := binary.Read(packed, binary.LittleEndian, &ch.EffectParameter); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// EnvPoint is a representation of an XM file envelope point
type EnvPoint struct {
	X uint16
	Y uint16
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

// SampleHeader is a representation of the XM file sample header
type SampleHeader struct {
	Length             uint32
	LoopStart          uint32
	LoopLength         uint32
	Volume             uint8
	Finetune           int8
	Flags              SampleFlags
	Panning            uint8
	RelativeNoteNumber int8
	ReservedP17        uint8
	Name               [22]uint8
	SampleData         []uint8
}

// InstrumentHeader is a representation of the XM file instrument header
type InstrumentHeader struct {
	Size         uint32
	Name         [22]uint8
	Type         uint8
	SamplesCount uint16

	SampleHeaderSize uint32
	SampleNumber     [96]uint8
	VolEnv           [12]EnvPoint
	PanEnv           [12]EnvPoint

	VolPoints         uint8
	PanPoints         uint8
	VolSustainPoint   uint8
	VolLoopStartPoint uint8
	VolLoopEndPoint   uint8
	PanSustainPoint   uint8
	PanLoopStartPoint uint8
	PanLoopEndPoint   uint8
	VolFlags          EnvelopeFlags
	PanFlags          EnvelopeFlags
	VibratoType       uint8
	VibratoSweep      uint8
	VibratoDepth      uint8
	VibratoRate       uint8
	VolumeFadeout     uint16
	ReservedP241      [11]uint16

	Samples []SampleHeader
}

func readInstrumentHeaderPartial(r io.Reader) (*InstrumentHeader, error) {
	ih := InstrumentHeader{}

	sz := uint32(0)
	if err := binary.Read(r, binary.LittleEndian, &ih.Size); err != nil {
		return nil, err
	}
	sz += 4

	if err := binary.Read(r, binary.LittleEndian, &ih.Name); err != nil {
		return nil, err
	}
	sz += 22

	if err := binary.Read(r, binary.LittleEndian, &ih.Type); err != nil {
		return nil, err
	}
	sz++

	if err := binary.Read(r, binary.LittleEndian, &ih.SamplesCount); err != nil {
		return nil, err
	}
	if sz += 2; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.SampleHeaderSize); err != nil {
		return nil, err
	}
	if sz += 4; sz >= ih.Size {
		return &ih, nil
	}

	for i := range ih.SampleNumber {
		if err := binary.Read(r, binary.LittleEndian, &ih.SampleNumber[i]); err != nil {
			return nil, err
		}
		if sz++; sz >= ih.Size {
			return &ih, nil
		}
	}

	for i := range ih.VolEnv {
		if err := binary.Read(r, binary.LittleEndian, &ih.VolEnv[i].X); err != nil {
			return nil, err
		}
		if sz += 2; sz >= ih.Size {
			return &ih, nil
		}
		if err := binary.Read(r, binary.LittleEndian, &ih.VolEnv[i].Y); err != nil {
			return nil, err
		}
		if sz += 2; sz >= ih.Size {
			return &ih, nil
		}
	}

	for i := range ih.PanEnv {
		if err := binary.Read(r, binary.LittleEndian, &ih.PanEnv[i].X); err != nil {
			return nil, err
		}
		if sz += 2; sz >= ih.Size {
			return &ih, nil
		}
		if err := binary.Read(r, binary.LittleEndian, &ih.PanEnv[i].Y); err != nil {
			return nil, err
		}
		if sz += 2; sz >= ih.Size {
			return &ih, nil
		}
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.VolPoints); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.PanPoints); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.VolSustainPoint); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.VolLoopStartPoint); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.VolLoopEndPoint); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.PanSustainPoint); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.PanLoopStartPoint); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.PanLoopEndPoint); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.VolFlags); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.PanFlags); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.VibratoType); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.VibratoSweep); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.VibratoDepth); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.VibratoRate); err != nil {
		return nil, err
	}
	if sz++; sz >= ih.Size {
		return &ih, nil
	}

	if err := binary.Read(r, binary.LittleEndian, &ih.VolumeFadeout); err != nil {
		return nil, err
	}
	if sz += 2; sz >= ih.Size {
		return &ih, nil
	}

	for i := range ih.ReservedP241 {
		if err := binary.Read(r, binary.LittleEndian, &ih.ReservedP241[i]); err != nil {
			return nil, err
		}
		if sz += 2; sz >= ih.Size {
			return &ih, nil
		}
	}

	return &ih, nil
}

func convertSample16Bit(data []uint8) {
	old := int16(0)
	for i := 0; i < len(data); i += 2 {
		s := binary.LittleEndian.Uint16(data[i:])
		new := int16(s) + old
		binary.LittleEndian.PutUint16(data[i:], uint16(new))
		old = new
	}
}

func convertSample8Bit(data []uint8) {
	old := int8(0)
	for i, s := range data {
		new := int8(s) + old
		data[i] = uint8(new)
		old = new
	}
}

func readInstrumentHeader(r io.Reader) (*InstrumentHeader, error) {
	ih, err := readInstrumentHeaderPartial(r)
	if err != nil {
		return nil, err
	}

	if ih.Size < 29 {
		return nil, errors.New("unusually small instrument header size - possibly corrupt file")
	}

	for i := uint16(0); i < ih.SamplesCount; i++ {
		s := SampleHeader{}

		if err := binary.Read(r, binary.LittleEndian, &s.Length); err != nil {
			return nil, err
		}

		if err := binary.Read(r, binary.LittleEndian, &s.LoopStart); err != nil {
			return nil, err
		}

		if err := binary.Read(r, binary.LittleEndian, &s.LoopLength); err != nil {
			return nil, err
		}

		if err := binary.Read(r, binary.LittleEndian, &s.Volume); err != nil {
			return nil, err
		}

		if err := binary.Read(r, binary.LittleEndian, &s.Finetune); err != nil {
			return nil, err
		}

		if err := binary.Read(r, binary.LittleEndian, &s.Flags); err != nil {
			return nil, err
		}

		if err := binary.Read(r, binary.LittleEndian, &s.Panning); err != nil {
			return nil, err
		}

		if err := binary.Read(r, binary.LittleEndian, &s.RelativeNoteNumber); err != nil {
			return nil, err
		}

		if err := binary.Read(r, binary.LittleEndian, &s.ReservedP17); err != nil {
			return nil, err
		}

		if err := binary.Read(r, binary.LittleEndian, &s.Name); err != nil {
			return nil, err
		}

		s.SampleData = make([]uint8, int(s.Length))

		ih.Samples = append(ih.Samples, s)
	}

	for _, s := range ih.Samples {
		if err := binary.Read(r, binary.LittleEndian, &s.SampleData); err != nil {
			return nil, err
		}

		// convert the sample in the background
		if (s.Flags & SampleFlag16Bit) != 0 {
			convertSample16Bit(s.SampleData)
		} else {
			convertSample8Bit(s.SampleData)
		}
	}
	return ih, nil
}

// File is an XM internal file representation
type File struct {
	Head        ModuleHeader
	Patterns    []Pattern
	Instruments []InstrumentHeader
}

// Read reads an XM file from the reader `r` and creates an internal File representation
func Read(r io.Reader) (*File, error) {
	xmh, err := readHeader(r)
	if err != nil {
		return nil, err
	}

	f := File{
		Head: *xmh,
	}

	for i := uint16(0); i < xmh.NumPatterns; i++ {
		p := Pattern{}

		ph, err := readPatternHeader(r, xmh.VersionNumber)
		if err != nil {
			return nil, err
		}

		p.Header = *ph

		ppd := make([]byte, int(ph.PackedPatternDataSize))
		if err := binary.Read(r, binary.LittleEndian, &ppd); err != nil {
			return nil, err
		}

		p.PackedData = ppd

		if err := p.unpack(int(xmh.NumChannels)); err != nil {
			return nil, err
		}

		f.Patterns = append(f.Patterns, p)
	}

	for i := uint16(0); i < xmh.NumInstruments; i++ {
		ih, err := readInstrumentHeader(r)
		if err != nil {
			return nil, err
		}

		f.Instruments = append(f.Instruments, *ih)
	}

	return &f, err
}
