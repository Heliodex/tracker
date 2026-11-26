package loadtext

import (
	"bytes"
	"strings"
	"testing"
)

const (
	// instrumentSample has an instrument with envelopes, vibrato and one sample. KEYMAP is filled in by withInstrument
	instrumentSample = `Name: t
NumChannels: 1
NumPatterns: 0
NumInstruments: 1
Speed: 6
Tempo: 125
LinearSlides: 1
Order:

Instrument 0
Name: kick
Type: 0
Keymap: KEYMAP
VolumeEnvelope: 0,64 10,32
VolumeSustain: 1
VolumeLoop: 0 1
VolumeFlags: 3
PanEnvelope:
PanSustain: 0
PanLoop: 0 0
PanFlags: 0
Vibrato: 0 0 0 0
VolumeFadeout: 256
Samples: 1

Sample 0
Name: kick
File: samples/00-00-kick.wav
Volume: 64
Finetune: -8
Panning: 128
RelativeNote: 0
LoopMode: 1
LoopStart: 0
LoopLength: 100
`
	sample = `Name: cool guy
NumChannels: 2
NumPatterns: 1
NumInstruments: 0
Speed: 6
Tempo: 125
LinearSlides: 1
Order: 0

Pattern 0
Rows: 2
Row 000: C-4 01 .. .. .. | ... .. .. .. ..
Row 001: ### .. .. .. .. | A#5 02 40 0F 1F
`
)

func TestReadText(t *testing.T) {
	tf, err := ReadText(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tf.NumChannels != 2 || tf.NumPatterns != 1 {
		t.Fatalf("got NumChannels %d, NumPatterns %d", tf.NumChannels, tf.NumPatterns)
	}

	if tf.Speed != 6 || tf.Tempo != 125 || !tf.LinearSlides {
		t.Fatalf("got speed %d, tempo %d, linear %v", tf.Speed, tf.Tempo, tf.LinearSlides)
	}

	if len(tf.OrderTable) != 1 || tf.OrderTable[0] != 0 {
		t.Fatalf("got order table %v", tf.OrderTable)
	}

	if got := strings.TrimRight(string(tf.Name[:]), "\x00"); got != "cool guy" {
		t.Fatalf("got name %q", got)
	}

	p := tf.Patterns[0]

	if got := p[0][0]; got != (ChannelData{
		Note: 49,
		Instrument: 1,
	}) {
		t.Errorf("row 0 channel 0: got %+v", got)
	}

	if got := p[0][1]; got != (ChannelData{}) {
		t.Errorf("row 0 channel 1: got %+v", got)
	}

	if got := p[1][0]; got.Note != 97 {
		t.Errorf("row 1 channel 0: want note off, got %+v", got)
	}

	if got := p[1][1]; got != (ChannelData{
		Note: 71,
		Instrument: 2,
		Volume: 0x40,
		Effect: 0x0F,
		EffectParameter: 0x1F,
	}) {
		t.Errorf("row 1 channel 1: got %+v", got)
	}
}

func withInstrument() string {
	keymap := strings.TrimSpace(strings.Repeat("0 ", 96))

	return strings.Replace(instrumentSample, "KEYMAP", keymap, 1)
}

func TestReadTextErrors(t *testing.T) {
	tests := map[string]string{
		"missing channel": strings.Replace(sample, "| A#5", "A#5", 1),
		"bad note": strings.Replace(sample, "C-4", "H-4", 1),
		"bad hex": strings.Replace(sample, "02 40", "0G 40", 1),
		"row out of order": strings.Replace(sample, "Row 001", "Row 002", 1),
		"order out of range": strings.Replace(sample, "Order: 0", "Order: 3", 1),
		"missing blank line": strings.Replace(sample, "\n\nPattern", "\nPattern", 1),
		"missing speed": strings.Replace(sample, "Speed: 6\n", "", 1),
		"finetune out of range": strings.Replace(withInstrument(), "Finetune: -8", "Finetune: 200", 1),
		"sample without file": strings.Replace(withInstrument(), "File: samples/00-00-kick.wav", "File: ", 1),
		"too many envelope points": strings.Replace(withInstrument(), "0,64 10,32", strings.Repeat("1,1 ", 13), 1),
	}

	for name, text := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ReadText(strings.NewReader(text)); err == nil {
				t.Errorf("expected an error")
			}
		})
	}
}

func TestReadTextInstrument(t *testing.T) {
	tf, err := ReadText(strings.NewReader(withInstrument()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tf.Instruments) != 1 {
		t.Fatalf("got %d instruments", len(tf.Instruments))
	}

	ins := tf.Instruments[0]
	if len(ins.VolPoints) != 2 || ins.VolPoints[1].X != 10 || ins.VolPoints[1].Y != 32 {
		t.Errorf("got volume points %+v", ins.VolPoints)
	}

	if ins.VolumeFadeout != 256 || ins.VolFlags != 3 || ins.VolSustain != 1 {
		t.Errorf("got fadeout %d, flags %d, sustain %d", ins.VolumeFadeout, ins.VolFlags, ins.VolSustain)
	}

	if len(ins.Samples) != 1 {
		t.Fatalf("got %d samples", len(ins.Samples))
	}

	if s := ins.Samples[0]; s.File != "samples/00-00-kick.wav" || s.Finetune != -8 || s.LoopMode != 1 || s.LoopLength != 100 {
		t.Errorf("got sample %+v", s)
	}
}

func TestWriteTextRoundTrip(t *testing.T) {
	for name, text := range map[string]string{"patterns": sample, "instruments": withInstrument()} {
		t.Run(name, func(t *testing.T) {
			tf, err := ReadText(strings.NewReader(text))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var b bytes.Buffer

			if err := WriteText(&b, tf); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := b.String(); got != text {
				t.Errorf("output mismatch:\n%s\nwant:\n%s", got, text)
			}
		})
	}
}
