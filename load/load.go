package load

import (
	"bytes"
	"encoding/binary"
	"errors"

	. "github.com/Heliodex/tracker/types"
)

func readHeaderPartial(r *bytes.Buffer) (xmh *ModuleHeader, err error) {
	xmh = &ModuleHeader{}

	var sz uint32
	if err = binary.Read(r, binary.LittleEndian, &xmh.ModuleHeader1); err != nil {
		return
	}
	if sz += 4 + 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &xmh.RestartPosition); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &xmh.NumChannels); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &xmh.NumPatterns); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &xmh.NumInstruments); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &xmh.Flags); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &xmh.DefaultSpeed); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &xmh.DefaultTempo); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	for i := range xmh.OrderTable {
		if err = binary.Read(r, binary.LittleEndian, &xmh.OrderTable[i]); err != nil {
			return
		}
		if sz++; sz >= xmh.HeaderSize {
			return
		}
	}

	return
}

func readHeader(r *bytes.Buffer) (xmh *ModuleHeader, err error) {
	if xmh, err = readHeaderPartial(r); err != nil {
		return
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

	return
}

func readPatternHeaderPartial(r *bytes.Buffer) (ph *PatternHeader, err error) {
	ph = &PatternHeader{}

	var sz uint32
	if err = binary.Read(r, binary.LittleEndian, &ph.PatternHeaderLength); err != nil {
		return
	}
	if sz += 4; sz >= ph.PatternHeaderLength {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ph.PackingType); err != nil {
		return
	}
	if sz++; sz >= ph.PatternHeaderLength {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ph.NumRows); err != nil {
		return
	}
	if sz += 2; sz >= ph.PatternHeaderLength {
		return
	}

	return ph, binary.Read(r, binary.LittleEndian, &ph.PackedPatternDataSize)
}

func readPatternHeader(r *bytes.Buffer) (ph *PatternHeader, err error) {
	if ph, err = readPatternHeaderPartial(r); err != nil {
		return
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

	return
}

func readInstrumentHeaderPartial(r *bytes.Buffer) (ih *InstrumentHeader, err error) {
	// var ih InstrumentHeader
	ih = &InstrumentHeader{}

	var sz uint32
	if err = binary.Read(r, binary.LittleEndian, &ih.InstrumentHeader1); err != nil {
		return
	}
	if sz += 4 + 22 + 1 + 2; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.SampleHeaderSize); err != nil {
		return
	}
	if sz += 4; sz >= ih.Size {
		return
	}

	for i := range ih.SampleNumber {
		if err = binary.Read(r, binary.LittleEndian, &ih.SampleNumber[i]); err != nil {
			return
		}
		if sz++; sz >= ih.Size {
			return
		}
	}

	for i := range ih.VolEnv {
		if err = binary.Read(r, binary.LittleEndian, &ih.VolEnv[i].X); err != nil {
			return
		}
		if sz += 2; sz >= ih.Size {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &ih.VolEnv[i].Y); err != nil {
			return
		}
		if sz += 2; sz >= ih.Size {
			return
		}
	}

	for i := range ih.PanEnv {
		if err = binary.Read(r, binary.LittleEndian, &ih.PanEnv[i].X); err != nil {
			return
		}
		if sz += 2; sz >= ih.Size {
			return
		}
		if err = binary.Read(r, binary.LittleEndian, &ih.PanEnv[i].Y); err != nil {
			return
		}
		if sz += 2; sz >= ih.Size {
			return
		}
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.VolPoints); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.PanPoints); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.VolSustainPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.VolLoopStartPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.VolLoopEndPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.PanSustainPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.PanLoopStartPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.PanLoopEndPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.VolFlags); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.PanFlags); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.VibratoType); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.VibratoSweep); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.VibratoDepth); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.VibratoRate); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Read(r, binary.LittleEndian, &ih.VolumeFadeout); err != nil {
		return
	}
	if sz += 2; sz >= ih.Size {
		return
	}

	for i := range ih.ReservedP241 {
		if err = binary.Read(r, binary.LittleEndian, &ih.ReservedP241[i]); err != nil {
			return
		}
		if sz += 2; sz >= ih.Size {
			return
		}
	}

	return
}

func ConvertSample16Bit(data []uint8) []uint8 {
	converted := make([]uint8, len(data))

	var old int16
	for i := 0; i < len(data); i += 2 {
		s := binary.LittleEndian.Uint16(data[i:])
		new := int16(s) + old
		binary.LittleEndian.PutUint16(converted[i:], uint16(new))
		old = new
	}

	return converted
}

func ConvertSample8Bit(data []uint8) []uint8 {
	converted := make([]uint8, len(data))

	var old int8
	for i, s := range data {
		new := int8(s) + old
		converted[i] = uint8(new)
		old = new
	}

	return converted
}

// func bufoffset(b *bytes.Buffer) int {
// 	// WILD.
// 	return -(b.Available() - b.Cap() + b.Len())
// }

func readInstrumentHeader(r *bytes.Buffer) (ih *InstrumentHeader, err error) {
	if ih, err = readInstrumentHeaderPartial(r); err != nil {
		return
	}

	if ih.Size < 29 {
		return nil, errors.New("unusually small instrument header size - possibly corrupt file")
	}

	ih.Samples = make([]SampleHeader, ih.SamplesCount)
	for i := range ih.SamplesCount {
		if err = binary.Read(r, binary.LittleEndian, &ih.Samples[i].SampleHeader1); err != nil {
			return
		}
	}

	for i, s := range ih.Samples {
		// fmt.Println("reading4", bufoffset(r), s.Length)

		sd := make([]uint8, s.Length)
		if err = binary.Read(r, binary.LittleEndian, &sd); err != nil {
			return
		}

		// convert the sample in the background
		if (s.Flags & SampleFlag16Bit) != 0 {
			ih.Samples[i].SampleData = ConvertSample16Bit(sd)
		} else {
			ih.Samples[i].SampleData = ConvertSample8Bit(sd)
		}
	}

	return
}

// Read reads an XM file from the reader `r` and creates an internal File representation
func Read(r *bytes.Buffer) (f *File, err error) {
	xmh, err := readHeader(r)
	if err != nil {
		return
	}

	if xmh.VersionNumber != 0x0104 {
		return nil, errors.New("unsupported XM file version")
	}

	// fmt.Println("reading0", bufoffset(r))

	f = &File{
		Head: *xmh,
	}

	for range xmh.NumPatterns {
		var p Pattern

		var ph *PatternHeader
		if ph, err = readPatternHeader(r); err != nil {
			return
		}
		p.Header = *ph

		ppd := make([]byte, ph.PackedPatternDataSize)
		if err = binary.Read(r, binary.LittleEndian, &ppd); err != nil {
			return
		}
		// p.PackedData = ppd

		if err = p.Unpack(int(xmh.NumChannels), ppd); err != nil {
			return
		}
		f.Patterns = append(f.Patterns, p)
	}

	// fmt.Println("reading1", bufoffset(r))

	for range xmh.NumInstruments {
		var ih *InstrumentHeader
		if ih, err = readInstrumentHeader(r); err != nil {
			return
		}

		f.Instruments = append(f.Instruments, *ih)
	}

	return
}
