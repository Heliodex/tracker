package recipe

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	builderLayer struct {
		wave string
		freq, amp float64
		opts []string
	}

	// Builder writes a recipe in code, instead of as recipe text. Its String method gives the same text a recipe file holds,
	// so the result can go anywhere a recipe string does:
	//
	//	recipe.New(256).Square(261.34375, 0.45).Duty(0.25).Sine(130.671875, 0.35).String()
	//
	// Each wave method adds a layer. Duty and Phase set an option on the layer added last, and panic if no layer has been
	// added yet, since that is a mistake in the code. Values are checked when the text is parsed, not here, so String never fails.
	Builder struct {
		frames int
		sixteen bool
		layers []builderLayer
	}
)

// Sixteen makes the recipe 16-bit, rather than the default 8-bit
func (b *Builder) Sixteen() *Builder {
	b.sixteen = true

	return b
}

func (b *Builder) layer(wave string, freq, amp float64) *Builder {
	b.layers = append(b.layers, builderLayer{
		wave: wave,
		freq: freq,
		amp: amp,
	})

	return b
}

// Noise adds a noise layer. The format requires a frequency, so pass one even though it has no pitch
func (b *Builder) Noise(freq, amp float64) *Builder {
	return b.layer("noise", freq, amp)
}

// Saw adds a saw layer at freq Hz and amp from 0 to 1
func (b *Builder) Saw(freq, amp float64) *Builder {
	return b.layer("saw", freq, amp)
}

// Sine adds a sine layer at freq Hz and amp from 0 to 1
func (b *Builder) Sine(freq, amp float64) *Builder {
	return b.layer("sine", freq, amp)
}

// Square adds a square layer at freq Hz and amp from 0 to 1
func (b *Builder) Square(freq, amp float64) *Builder {
	return b.layer("square", freq, amp)
}

// Triangle adds a triangle layer at freq Hz and amp from 0 to 1
func (b *Builder) Triangle(freq, amp float64) *Builder {
	return b.layer("triangle", freq, amp)
}

// New starts a recipe with the given length in frames, at 8 bits
func New(frames int) *Builder {
	return &Builder{
		frames: frames,
	}
}

// formatFloat gives the shortest form that parses back to the same value
func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// String returns the recipe as text, in the same form Parse reads
func (b *Builder) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Frames: %d\n", b.frames)

	if b.sixteen {
		sb.WriteString("Bits: 16\n")
	}

	for _, l := range b.layers {
		fmt.Fprintf(&sb, "Layer: %s %s %s", l.wave, formatFloat(l.freq), formatFloat(l.amp))

		for _, o := range l.opts {
			sb.WriteString(" " + o)
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func (b *Builder) option(key string, v float64) *Builder {
	if len(b.layers) == 0 {
		panic(fmt.Sprintf("recipe: %s set before any layer was added", key))
	}

	l := &b.layers[len(b.layers)-1]
	l.opts = append(l.opts, key+"="+formatFloat(v))

	return b
}

// Duty sets the duty of the last layer, from 0 to 1. Only square layers accept it
func (b *Builder) Duty(duty float64) *Builder {
	return b.option("duty", duty)
}

// Phase sets the phase of the last layer, in cycles
func (b *Builder) Phase(phase float64) *Builder {
	return b.option("phase", phase)
}
