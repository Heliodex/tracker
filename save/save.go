package save

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	. "github.com/Heliodex/tracker/types"
)

func writeHeaderPartial(w *bytes.Buffer, xmh *ModuleHeader) (err error) {
	var sz uint32
	if err = binary.Write(w, binary.LittleEndian, xmh.ModuleHeader1); err != nil {
		return
	}
	if sz += 4 + 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, xmh.RestartPosition); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, xmh.NumChannels); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, xmh.NumPatterns); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, xmh.NumInstruments); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, uint16(xmh.Flags)); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, xmh.DefaultSpeed); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, xmh.DefaultTempo); err != nil {
		return
	}
	if sz += 2; sz >= xmh.HeaderSize {
		return
	}

	for i := range xmh.OrderTable {
		if err = binary.Write(w, binary.LittleEndian, xmh.OrderTable[i]); err != nil {
			return
		}
		if sz++; sz >= xmh.HeaderSize {
			return
		}
	}

	return
}

func writeHeader(w *bytes.Buffer, xmh *ModuleHeader) error {
	if xmh.NumChannels < 1 || xmh.NumChannels > 32 {
		return errors.New("invalid number of channels")
	}

	if xmh.NumPatterns > 256 {
		return errors.New("invalid number of patterns")
	}

	if xmh.NumInstruments > 128 {
		return errors.New("invalid number of instruments")
	}

	return writeHeaderPartial(w, xmh)
}

func writePatternHeaderPartial(w *bytes.Buffer, ph *PatternHeader) (err error) {
	var sz uint32
	if err = binary.Write(w, binary.LittleEndian, ph.PatternHeaderLength); err != nil {
		return
	}
	if sz += 4; sz >= ph.PatternHeaderLength {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, ph.PackingType); err != nil {
		return
	}
	if sz++; sz >= ph.PatternHeaderLength {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, ph.NumRows); err != nil {
		return
	}
	if sz += 2; sz >= ph.PatternHeaderLength {
		return
	}

	return binary.Write(w, binary.LittleEndian, ph.PackedPatternDataSize)
}

func writePatternHeader(w *bytes.Buffer, ph *PatternHeader) (err error) {
	if ph.PackingType != 0 {
		return errors.New("unexpected pattern packing type")
	}

	if ph.NumRows < 1 || ph.NumRows > 256 {
		return errors.New("pattern row count out of range")
	}

	return writePatternHeaderPartial(w, ph)
}

func writeInstrumentHeaderPartial(w *bytes.Buffer, ih *InstrumentHeader) (err error) {
	var sz uint32
	if err = binary.Write(w, binary.LittleEndian, &ih.InstrumentHeader1); err != nil {
		return
	}
	if sz += 4 + 22 + 1 + 2; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.SampleHeaderSize); err != nil {
		return
	}
	if sz += 4; sz >= ih.Size {
		return
	}

	for i := range ih.SampleNumber {
		if err = binary.Write(w, binary.LittleEndian, &ih.SampleNumber[i]); err != nil {
			return
		}
		if sz++; sz >= ih.Size {
			return
		}
	}

	for i := range ih.VolEnv {
		if err = binary.Write(w, binary.LittleEndian, &ih.VolEnv[i].X); err != nil {
			return
		}
		if sz += 2; sz >= ih.Size {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &ih.VolEnv[i].Y); err != nil {
			return
		}
		if sz += 2; sz >= ih.Size {
			return
		}
	}

	for i := range ih.PanEnv {
		if err = binary.Write(w, binary.LittleEndian, &ih.PanEnv[i].X); err != nil {
			return
		}
		if sz += 2; sz >= ih.Size {
			return
		}
		if err = binary.Write(w, binary.LittleEndian, &ih.PanEnv[i].Y); err != nil {
			return
		}
		if sz += 2; sz >= ih.Size {
			return
		}
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.VolPoints); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.PanPoints); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.VolSustainPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.VolLoopStartPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.VolLoopEndPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.PanSustainPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.PanLoopStartPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.PanLoopEndPoint); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.VolFlags); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.PanFlags); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.VibratoType); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.VibratoSweep); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.VibratoDepth); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.VibratoRate); err != nil {
		return
	}
	if sz++; sz >= ih.Size {
		return
	}

	if err = binary.Write(w, binary.LittleEndian, &ih.VolumeFadeout); err != nil {
		return
	}
	if sz += 2; sz >= ih.Size {
		return
	}

	for i := range ih.ReservedP241 {
		if err = binary.Write(w, binary.LittleEndian, &ih.ReservedP241[i]); err != nil {
			return
		}
		if sz += 2; sz >= ih.Size {
			return
		}
	}

	return
}

func unconvertSample16Bit(converted []uint8) []uint8 {
	data := make([]uint8, len(converted))

	var new int16
	for i := 0; i < len(converted); i += 2 {
		s := binary.LittleEndian.Uint16(converted[i:])
		old := int16(s) - new
		binary.LittleEndian.PutUint16(data[i:], uint16(old))
		new = old
	}

	return data
}

func unconvertSample8Bit(converted []uint8) []uint8 {
	data := make([]uint8, len(converted))

	var new int8
	for i, s := range converted {
		old := int8(s) - new
		data[i] = uint8(old)
		new = old
	}

	return data
}

func writeInstrumentHeader(w *bytes.Buffer, ih *InstrumentHeader) (err error) {
	if ih.Size < 29 {
		return errors.New("unusually small instrument header size")
	}

	if err = writeInstrumentHeaderPartial(w, ih); err != nil {
		return
	}

	for range ih.SamplesCount {
		var s SampleHeader
		if err = binary.Write(w, binary.LittleEndian, &s.SampleHeader1); err != nil {
			return
		}

		s.SampleData = make([]uint8, s.Length)

		ih.Samples = append(ih.Samples, s)
	}

	for _, s := range ih.Samples {
		if err = binary.Write(w, binary.LittleEndian, &s.SampleData); err != nil {
			return
		}

		// convert the sample in the background
		if (s.Flags & SampleFlag16Bit) != 0 {
			unconvertSample16Bit(s.SampleData)
		} else {
			unconvertSample8Bit(s.SampleData)
		}
	}

	return
}

// Write writes an XM file from the File `f` to the writer `w`
func Write(w io.Writer, f *File) (err error) {
	var buf bytes.Buffer

	if err = writeHeader(&buf, &f.Head); err != nil {
		return
	}

	for i := range f.Patterns {
		p := &f.Patterns[i]
		if err = writePatternHeaderPartial(&buf, &p.Header); err != nil {
			return
		}

		var packed []byte
		if packed, err = p.Pack(int(f.Head.NumChannels)); err != nil {
			return
		}
		if _, err = buf.Write(packed); err != nil {
			return
		}
	}

	for i := range f.Instruments {
		ih := &f.Instruments[i]
		if err = writeInstrumentHeader(&buf, ih); err != nil {
			return
		}
		for j := range ih.Samples {
			if err = writeSampleHeader(&buf, &ih.Samples[j]); err != nil {
				return
			}
		}
	}

	_, err = w.Write(buf.Bytes())
	return
}
