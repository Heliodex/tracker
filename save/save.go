package save

import (
	"bytes"
	"encoding/binary"
	"io"

	. "github.com/Heliodex/tracker/types"
)

func Write(w io.Writer, f *File) (err error) {
	var buf bytes.Buffer

	// Write ModuleHeader fields in the same order as they are read
	if _, err = buf.Write(f.Head.IDText[:]); err != nil {
		return
	}
	if _, err = buf.Write(f.Head.Name[:]); err != nil {
		return
	}
	if err = binary.Write(&buf, binary.LittleEndian, f.Head.Reserved1A); err != nil {
		return
	}
	if _, err = buf.Write(f.Head.TrackerName[:]); err != nil {
		return
	}
	if err = binary.Write(&buf, binary.LittleEndian, f.Head.VersionNumber); err != nil {
		return
	}
	// HeaderSize is written as stored; we will not recompute it here
	if err = binary.Write(&buf, binary.LittleEndian, f.Head.HeaderSize); err != nil {
		return
	}
	if err = binary.Write(&buf, binary.LittleEndian, f.Head.SongLength); err != nil {
		return
	}
	if err = binary.Write(&buf, binary.LittleEndian, f.Head.RestartPosition); err != nil {
		return
	}
	if err = binary.Write(&buf, binary.LittleEndian, f.Head.NumChannels); err != nil {
		return
	}
	if err = binary.Write(&buf, binary.LittleEndian, f.Head.NumPatterns); err != nil {
		return
	}
	if err = binary.Write(&buf, binary.LittleEndian, f.Head.NumInstruments); err != nil {
		return
	}
	if err = binary.Write(&buf, binary.LittleEndian, uint16(f.Head.Flags)); err != nil {
		return
	}
	if err = binary.Write(&buf, binary.LittleEndian, f.Head.DefaultSpeed); err != nil {
		return
	}
	if err = binary.Write(&buf, binary.LittleEndian, f.Head.DefaultTempo); err != nil {
		return
	}
	// OrderTable (256 bytes)
	for i := range 256 {
		if err = binary.Write(&buf, binary.LittleEndian, f.Head.OrderTable[i]); err != nil {
			return
		}
	}

	// Write Patterns
	for _, p := range f.Patterns {
		ph := &p.Header
		if err = binary.Write(&buf, binary.LittleEndian, ph.PatternHeaderLength); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ph.PackingType); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ph.NumRows); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ph.PackedPatternDataSize); err != nil {
			return
		}
		// Pack pattern data back to the original packed form

		var packed []byte
		if packed, err = p.Pack(int(f.Head.NumChannels)); err != nil {
			return
		}
		if _, err = buf.Write(packed); err != nil {
			return
		}
	}

	// Write Instruments
	for _, ih := range f.Instruments {
		if err = binary.Write(&buf, binary.LittleEndian, ih.Size); err != nil {
			return
		}
		if _, err = buf.Write(ih.Name[:]); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.Type); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.SamplesCount); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.SampleHeaderSize); err != nil {
			return
		}
		if _, err := buf.Write(ih.SampleNumber[:]); err != nil {
			return
		}
		// VolEnv and PanEnv arrays
		for i := range ih.VolEnv {
			if err = binary.Write(&buf, binary.LittleEndian, ih.VolEnv[i].X); err != nil {
				return
			}
			if err = binary.Write(&buf, binary.LittleEndian, ih.VolEnv[i].Y); err != nil {
				return
			}
		}
		for i := range ih.PanEnv {
			if err = binary.Write(&buf, binary.LittleEndian, ih.PanEnv[i].X); err != nil {
				return
			}
			if err = binary.Write(&buf, binary.LittleEndian, ih.PanEnv[i].Y); err != nil {
				return
			}
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.VolPoints); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.PanPoints); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.VolSustainPoint); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.VolLoopStartPoint); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.VolLoopEndPoint); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.PanSustainPoint); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.PanLoopStartPoint); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.PanLoopEndPoint); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.VolFlags); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.PanFlags); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.VibratoType); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.VibratoSweep); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.VibratoDepth); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.VibratoRate); err != nil {
			return
		}
		if err = binary.Write(&buf, binary.LittleEndian, ih.VolumeFadeout); err != nil {
			return
		}
		for i := range ih.ReservedP241 {
			if err = binary.Write(&buf, binary.LittleEndian, ih.ReservedP241[i]); err != nil {
				return
			}
		}

		// Write each sample header and its data
		for _, s := range ih.Samples {
			if err = binary.Write(&buf, binary.LittleEndian, s.Length); err != nil {
				return
			}
			if err = binary.Write(&buf, binary.LittleEndian, s.LoopStart); err != nil {
				return
			}
			if err = binary.Write(&buf, binary.LittleEndian, s.LoopLength); err != nil {
				return
			}
			if err = binary.Write(&buf, binary.LittleEndian, s.Volume); err != nil {
				return
			}
			if err = binary.Write(&buf, binary.LittleEndian, s.Finetune); err != nil {
				return
			}
			if err = binary.Write(&buf, binary.LittleEndian, s.Flags); err != nil {
				return
			}
			if err = binary.Write(&buf, binary.LittleEndian, s.Panning); err != nil {
				return
			}
			if err = binary.Write(&buf, binary.LittleEndian, s.RelativeNoteNumber); err != nil {
				return
			}
			if err = binary.Write(&buf, binary.LittleEndian, s.ReservedP17); err != nil {
				return
			}
			if _, err = buf.Write(s.Name[:]); err != nil {
				return
			}
			if _, err = buf.Write(s.SampleData); err != nil {
				return
			}
		}
	}

	// Finally write the buffer to the provided writer
	_, err = w.Write(buf.Bytes())
	return
}
