package loadtext

import (
	"fmt"
	"io"
	"strings"

	"github.com/Heliodex/tracker/types"
)

// checkLine fails if s would break a line of the text format
func checkLine(what, s string) error {
	if strings.ContainsAny(s, "\r\n") {
		return fmt.Errorf("%s contains a line break", what)
	}

	return nil
}

// formatHex formats a hex column, using ".." for zero
func formatHex(v uint8) string {
	if v == 0 {
		return ".."
	}

	return fmt.Sprintf("%02X", v)
}

func formatCell(c ChannelData) (string, error) {
	if c.Note > 97 {
		return "", fmt.Errorf("note outside range: %d", c.Note)
	}

	return strings.Join([]string{
		types.Note(c.Note).String(),
		formatHex(c.Instrument),
		formatHex(c.Volume),
		formatHex(c.Effect),
		formatHex(c.EffectParameter),
	}, " "), nil
}

// formatInts joins values with spaces, with a leading space so it can follow a "Name:" label
func formatInts[T ~uint8 | ~int8](vs []T) string {
	var b strings.Builder

	for _, v := range vs {
		fmt.Fprintf(&b, " %d", v)
	}

	return b.String()
}

// formatPoints formats envelope points as " x,y x,y", with a leading space
func formatPoints(points []types.EnvPoint) string {
	var b strings.Builder

	for _, p := range points {
		fmt.Fprintf(&b, " %d,%d", p.X, p.Y)
	}

	return b.String()
}

// trimName returns a fixed-size name without its trailing NUL padding
func trimName(name []uint8) string {
	return strings.TrimRight(string(name), "\x00")
}

func writePattern(b *strings.Builder, index int, tf *TextFormat, pattern Pattern) error {
	if len(pattern) < 1 || len(pattern) > maxRows {
		return fmt.Errorf("pattern %d: row count out of range: %d", index, len(pattern))
	}

	fmt.Fprintf(b, "\nPattern %d\n", index)
	fmt.Fprintf(b, "Rows: %d\n", len(pattern))

	for r, row := range pattern {
		if len(row) != int(tf.NumChannels) {
			return fmt.Errorf("pattern %d row %d: expected %d channels, got %d", index, r, tf.NumChannels, len(row))
		}

		cells := make([]string, len(row))

		for c, cd := range row {
			cell, err := formatCell(cd)
			if err != nil {
				return fmt.Errorf("pattern %d row %d channel %d: %w", index, r, c+1, err)
			}

			cells[c] = cell
		}

		fmt.Fprintf(b, "Row %03d: %s\n", r, strings.Join(cells, " | "))
	}

	return nil
}

func writeSample(b *strings.Builder, index int, s *Sample) error {
	name := trimName(s.Name[:])

	if err := checkLine("sample name", name); err != nil {
		return err
	}

	if s.File == "" {
		return fmt.Errorf("sample %d has no file", index)
	}

	if err := checkLine("sample file", s.File); err != nil {
		return err
	}

	if s.LoopMode > 3 {
		return fmt.Errorf("sample %d: loop mode out of range: %d", index, s.LoopMode)
	}

	fmt.Fprintf(b, "\nSample %d\n", index)
	fmt.Fprintf(b, "Name: %s\n", name)
	fmt.Fprintf(b, "File: %s\n", s.File)
	fmt.Fprintf(b, "Volume: %d\n", s.Volume)
	fmt.Fprintf(b, "Finetune: %d\n", s.Finetune)
	fmt.Fprintf(b, "Panning: %d\n", s.Panning)
	fmt.Fprintf(b, "RelativeNote: %d\n", s.RelativeNote)
	fmt.Fprintf(b, "LoopMode: %d\n", s.LoopMode)
	fmt.Fprintf(b, "LoopStart: %d\n", s.LoopStart)
	fmt.Fprintf(b, "LoopLength: %d\n", s.LoopLength)

	return nil
}

func writeInstrument(b *strings.Builder, index int, ins *Instrument) error {
	name := trimName(ins.Name[:])

	if err := checkLine("instrument name", name); err != nil {
		return err
	}

	if len(ins.VolPoints) > maxEnvPoints || len(ins.PanPoints) > maxEnvPoints {
		return fmt.Errorf("instrument %d: more than %d envelope points", index, maxEnvPoints)
	}

	if len(ins.Samples) > maxSamples {
		return fmt.Errorf("instrument %d: too many samples: %d", index, len(ins.Samples))
	}

	fmt.Fprintf(b, "\nInstrument %d\n", index)
	fmt.Fprintf(b, "Name: %s\n", name)
	fmt.Fprintf(b, "Type: %d\n", ins.Type)
	fmt.Fprintf(b, "Keymap:%s\n", formatInts(ins.Keymap[:]))
	fmt.Fprintf(b, "VolumeEnvelope:%s\n", formatPoints(ins.VolPoints))
	fmt.Fprintf(b, "VolumeSustain: %d\n", ins.VolSustain)
	fmt.Fprintf(b, "VolumeLoop: %d %d\n", ins.VolLoopStart, ins.VolLoopEnd)
	fmt.Fprintf(b, "VolumeFlags: %d\n", ins.VolFlags)
	fmt.Fprintf(b, "PanEnvelope:%s\n", formatPoints(ins.PanPoints))
	fmt.Fprintf(b, "PanSustain: %d\n", ins.PanSustain)
	fmt.Fprintf(b, "PanLoop: %d %d\n", ins.PanLoopStart, ins.PanLoopEnd)
	fmt.Fprintf(b, "PanFlags: %d\n", ins.PanFlags)
	fmt.Fprintf(b, "Vibrato: %d %d %d %d\n", ins.VibratoType, ins.VibratoSweep, ins.VibratoDepth, ins.VibratoRate)
	fmt.Fprintf(b, "VolumeFadeout: %d\n", ins.VolumeFadeout)
	fmt.Fprintf(b, "Samples: %d\n", len(ins.Samples))

	for i := range ins.Samples {
		if err := writeSample(b, i, &ins.Samples[i]); err != nil {
			return fmt.Errorf("instrument %d: %w", index, err)
		}
	}

	return nil
}

// WriteText writes tf to w in the text format
func WriteText(w io.Writer, tf *TextFormat) error {
	if tf.NumChannels < 1 || tf.NumChannels > maxChannels {
		return fmt.Errorf("invalid NumChannels value: %d", tf.NumChannels)
	}

	if int(tf.NumPatterns) != len(tf.Patterns) {
		return fmt.Errorf("NumPatterns is %d but %d patterns are present", tf.NumPatterns, len(tf.Patterns))
	}

	if int(tf.NumInstruments) != len(tf.Instruments) {
		return fmt.Errorf("NumInstruments is %d but %d instruments are present", tf.NumInstruments, len(tf.Instruments))
	}

	name := trimName(tf.Name[:])

	if err := checkLine("name", name); err != nil {
		return err
	}

	var b strings.Builder

	fmt.Fprintf(&b, "Name: %s\n", name)
	fmt.Fprintf(&b, "NumChannels: %d\n", tf.NumChannels)
	fmt.Fprintf(&b, "NumPatterns: %d\n", tf.NumPatterns)
	fmt.Fprintf(&b, "NumInstruments: %d\n", tf.NumInstruments)
	fmt.Fprintf(&b, "Speed: %d\n", tf.Speed)
	fmt.Fprintf(&b, "Tempo: %d\n", tf.Tempo)

	var linear int

	if tf.LinearSlides {
		linear = 1
	}

	fmt.Fprintf(&b, "LinearSlides: %d\n", linear)
	fmt.Fprintf(&b, "Order:%s\n", formatInts(tf.OrderTable))

	for i := range tf.Patterns {
		if err := writePattern(&b, i, tf, tf.Patterns[i]); err != nil {
			return err
		}
	}

	for i := range tf.Instruments {
		if err := writeInstrument(&b, i, &tf.Instruments[i]); err != nil {
			return err
		}
	}

	_, err := io.WriteString(w, b.String())

	return err
}
