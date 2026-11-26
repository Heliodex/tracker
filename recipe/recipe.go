// Package recipe renders sample recipes: plain text descriptions of waveforms, turned into XM sample data.
//
// A recipe looks like this:
//
//	# a square lead with a sine sub
//	Frames: 256
//	Bits: 8
//	Layer: square 220 0.6 duty=0.25
//	Layer: sine 110 0.4
//
// Frames is the length of the sample in frames at 8363 Hz, which must be given. Bits is 8 or 16, and defaults to 8.
// Each layer is a wave (sine, square, triangle, saw or noise), a frequency in Hz, and an amplitude from 0 to 1.
// Options after those are duty (square only, default 0.5) and phase (in cycles, default 0).
// Layers are summed, and if the sum goes past full scale the whole sample is scaled down to fit.
//
// A frequency of rate / Frames gives exactly one cycle over the sample, which loops cleanly.
// Noise is deterministic, so a recipe always renders the same samples.
package recipe

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

type (
	wave int

	layer struct {
		wave wave
		freq, amp, duty, phase float64
	}

	// Recipe is a parsed recipe, ready to render
	Recipe struct {
		Frames int
		Sixteen bool
		Layers []layer
	}

	// xorshift is a small deterministic random generator, so noise is the same every time a recipe is rendered
	xorshift uint32
)

const (
	// Rate is the sample rate of every rendered sample. XM stores no sample rate, so this is the rate at which C-4 plays
	Rate = 8363

	// maxFrames is the longest sample a recipe may describe, about two minutes
	maxFrames = 1 << 20
)

const (
	sine wave = iota
	square
	triangle
	saw
	noise
)

var waveNames = map[string]wave{"noise": noise, "saw": saw, "sine": sine, "square": square, "triangle": triangle}

// parseLayer parses "wave freq amp [duty=x] [phase=x]"
func parseLayer(value string) (layer, error) {
	fields := strings.Fields(value)
	if len(fields) < 3 {
		return layer{}, fmt.Errorf("layer needs a wave, frequency and amplitude, got %q", value)
	}

	var l layer

	w, ok := waveNames[fields[0]]
	if !ok {
		return layer{}, fmt.Errorf("unknown wave %q", fields[0])
	}

	l.wave = w
	freq, err := strconv.ParseFloat(fields[1], 64)

	if err != nil || freq <= 0 || freq > Rate/2 {
		return layer{}, fmt.Errorf("frequency must be above 0 and at most %d Hz, got %q", Rate/2, fields[1])
	}

	l.freq = freq
	amp, err := strconv.ParseFloat(fields[2], 64)

	if err != nil || amp < 0 || amp > 1 {
		return layer{}, fmt.Errorf("amplitude must be between 0 and 1, got %q", fields[2])
	}

	l.amp = amp
	l.duty = 0.5

	for _, opt := range fields[3:] {
		key, v, ok := strings.Cut(opt, "=")
		if !ok {
			return layer{}, fmt.Errorf("expected option 'key=value', got %q", opt)
		}

		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return layer{}, fmt.Errorf("invalid %s value %q", key, v)
		}

		switch key {
		case "duty":
			if w != square {
				return layer{}, errors.New("duty only applies to square waves")
			}

			if f < 0 || f > 1 {
				return layer{}, fmt.Errorf("duty must be between 0 and 1, got %q", v)
			}

			l.duty = f
		case "phase":
			l.phase = f
		default:
			return layer{}, fmt.Errorf("unknown option %q", key)
		}
	}

	return l, nil
}

// Parse reads a recipe from r
func Parse(r io.Reader) (*Recipe, error) {
	rc := &Recipe{}

	var seenFrames bool

	sc := bufio.NewScanner(r)

	for line := 0; sc.Scan(); {
		line++
		text := strings.TrimSpace(sc.Text())

		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		key, value, ok := strings.Cut(text, ":")
		if !ok {
			return nil, fmt.Errorf("line %d: expected 'Key: value', got %q", line, text)
		}

		value = strings.TrimSpace(value)

		switch strings.TrimSpace(key) {
		case "Frames":
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 || n > maxFrames {
				return nil, fmt.Errorf("line %d: Frames must be between 1 and %d, got %q", line, maxFrames, value)
			}

			rc.Frames = n
			seenFrames = true
		case "Bits":
			switch value {
			case "8":
				rc.Sixteen = false
			case "16":
				rc.Sixteen = true
			default:
				return nil, fmt.Errorf("line %d: Bits must be 8 or 16, got %q", line, value)
			}
		case "Layer":
			l, err := parseLayer(value)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}

			rc.Layers = append(rc.Layers, l)
		default:
			return nil, fmt.Errorf("line %d: unknown key %q", line, key)
		}
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	if !seenFrames {
		return nil, errors.New("missing Frames")
	}

	if len(rc.Layers) == 0 {
		return nil, errors.New("no layers")
	}

	return rc, nil
}

// next returns a value from -1 to 1
func (x *xorshift) next() float64 {
	v := uint32(*x)
	v ^= v << 13
	v ^= v >> 17
	v ^= v << 5
	*x = xorshift(v)

	return float64(v)/math.MaxUint32*2 - 1
}

// waveAt returns the value of a wave at phase p, in cycles
func waveAt(w wave, p, duty float64, rng *xorshift) float64 {
	p -= math.Floor(p)

	switch w {
	case sine:
		return math.Sin(2 * math.Pi * p)
	case square:
		if p < duty {
			return 1
		}

		return -1
	case triangle:
		return 4*math.Abs(p-0.5) - 1
	case saw:
		return 2*p - 1
	}

	return rng.next()
}

// Render returns the sample data in XM form: signed 8-bit samples, or signed little-endian 16-bit samples
func (rc *Recipe) Render() []uint8 {
	sum := make([]float64, rc.Frames)

	for li, l := range rc.Layers {
		rng := xorshift(uint32(li)*2654435761 + 1)

		for i := range sum {
			t := float64(i) / Rate
			sum[i] += l.amp * waveAt(l.wave, l.freq*t+l.phase, l.duty, &rng)
		}
	}

	var peak float64

	for _, s := range sum {
		peak = math.Max(peak, math.Abs(s))
	}

	scale := 1.0

	if peak > 1 {
		scale = 1 / peak
	}

	bytesPerFrame := 1

	if rc.Sixteen {
		bytesPerFrame = 2
	}

	pcm := make([]uint8, 0, rc.Frames*bytesPerFrame)

	for _, s := range sum {
		if v := s * scale; rc.Sixteen {
			n := int16(math.Round(v * 32767))
			pcm = append(pcm, uint8(n), uint8(uint16(n)>>8))
		} else {
			pcm = append(pcm, uint8(int8(math.Round(v*127))))
		}
	}

	return pcm
}

// Render reads a recipe from r and renders it, returning the sample data and whether it is 16-bit
func Render(r io.Reader) (pcm []uint8, sixteen bool, err error) {
	rc, err := Parse(r)
	if err != nil {
		return nil, false, err
	}

	return rc.Render(), rc.Sixteen, nil
}
