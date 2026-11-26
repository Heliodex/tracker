// Package wav reads and writes the mono PCM WAV files that hold XM sample data.
//
// Sample data is passed around in XM form: signed 8-bit samples, or signed little-endian 16-bit samples.
// WAV stores 8-bit samples unsigned, so those are converted on the way in and out.
package wav

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// SampleRate is written to every file. XM stores no sample rate, so this is the rate at which C-4 plays
const SampleRate = 8363

// Decode reads a mono 8-bit or 16-bit PCM WAV file and returns its samples in XM form.
// The sample rate is ignored, as XM has no equivalent.
func Decode(file []byte) (pcm []uint8, sixteen bool, err error) {
	if len(file) < 12 || string(file[0:4]) != "RIFF" || string(file[8:12]) != "WAVE" {
		return nil, false, errors.New("not a WAV file")
	}

	var haveFmt bool

	rest := file[12:]

	var bits uint16

	for len(rest) >= 8 {
		id := string(rest[0:4])
		size := int(binary.LittleEndian.Uint32(rest[4:8]))
		rest = rest[8:]

		if size > len(rest) {
			return nil, false, fmt.Errorf("%q chunk is truncated", id)
		}

		chunk := rest[:size]

		// chunks are padded to an even length
		next := min(size+size%2, len(rest))
		rest = rest[next:]

		switch id {
		case "fmt ":
			if size < 16 {
				return nil, false, errors.New("fmt chunk is too short")
			}

			format := binary.LittleEndian.Uint16(chunk[0:2])
			channels := binary.LittleEndian.Uint16(chunk[2:4])
			bits = binary.LittleEndian.Uint16(chunk[14:16])

			if format != 1 {
				return nil, false, fmt.Errorf("unsupported WAV format %d, want uncompressed PCM", format)
			}

			if channels != 1 {
				return nil, false, fmt.Errorf("%d channels, only mono samples are supported", channels)
			}

			if bits != 8 && bits != 16 {
				return nil, false, fmt.Errorf("%d-bit samples, only 8-bit and 16-bit are supported", bits)
			}

			haveFmt = true
		case "data":
			if !haveFmt {
				return nil, false, errors.New("data chunk before fmt chunk")
			}

			if bits == 16 {
				if size%2 != 0 {
					return nil, false, errors.New("16-bit sample data has an odd length")
				}

				return append([]uint8(nil), chunk...), true, nil
			}

			pcm = make([]uint8, len(chunk))

			for i, s := range chunk {
				pcm[i] = s ^ 0x80
			}

			return pcm, false, nil
		}
	}

	return nil, false, errors.New("no data chunk")
}

// Encode returns a mono PCM WAV file holding pcm. sixteen selects 16-bit samples, otherwise 8-bit
func Encode(pcm []uint8, sixteen bool) ([]byte, error) {
	bytesPerSample := 1
	data := make([]uint8, len(pcm))

	if sixteen {
		bytesPerSample = 2

		if len(pcm)%2 != 0 {
			return nil, errors.New("16-bit sample data has an odd length")
		}

		copy(data, pcm)
	} else {
		for i, s := range pcm {
			data[i] = s ^ 0x80
		}
	}

	var b bytes.Buffer

	pad := len(data) % 2

	b.WriteString("RIFF")
	binary.Write(&b, binary.LittleEndian, uint32(4+8+16+8+len(data)+pad))
	b.WriteString("WAVE")
	b.WriteString("fmt ")
	binary.Write(&b, binary.LittleEndian, uint32(16))
	binary.Write(&b, binary.LittleEndian, uint16(1)) // uncompressed PCM
	binary.Write(&b, binary.LittleEndian, uint16(1)) // mono
	binary.Write(&b, binary.LittleEndian, uint32(SampleRate))
	binary.Write(&b, binary.LittleEndian, uint32(SampleRate*bytesPerSample))
	binary.Write(&b, binary.LittleEndian, uint16(bytesPerSample))
	binary.Write(&b, binary.LittleEndian, uint16(8*bytesPerSample))
	b.WriteString("data")
	binary.Write(&b, binary.LittleEndian, uint32(len(data)))
	b.Write(data)

	if pad == 1 {
		b.WriteByte(0)
	}

	return b.Bytes(), nil
}
