package loadtext

import (
	"strings"
	"testing"
)

const sample = `Name: cool guy
NumChannels: 2
NumPatterns: 1
Order: 0

Pattern 0
Rows: 2
Row 000: C-4 01 .. .. .. | ... .. .. .. ..
Row 001: ### .. .. .. .. | A#5 02 40 0F 1F
`

func TestReadText(t *testing.T) {
	tf, err := ReadText(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tf.NumChannels != 2 || tf.NumPatterns != 1 {
		t.Fatalf("got NumChannels %d, NumPatterns %d", tf.NumChannels, tf.NumPatterns)
	}

	if len(tf.OrderTable) != 1 || tf.OrderTable[0] != 0 {
		t.Fatalf("got order table %v", tf.OrderTable)
	}

	if got := strings.TrimRight(string(tf.Name[:]), "\x00"); got != "cool guy" {
		t.Fatalf("got name %q", got)
	}

	p := tf.Patterns[0]

	if got := p[0][0]; got != (ChannelData{Note: 49, Instrument: 1}) {
		t.Errorf("row 0 channel 0: got %+v", got)
	}

	if got := p[0][1]; got != (ChannelData{}) {
		t.Errorf("row 0 channel 1: got %+v", got)
	}

	if got := p[1][0]; got.Note != 97 {
		t.Errorf("row 1 channel 0: want note off, got %+v", got)
	}

	if got := p[1][1]; got != (ChannelData{Note: 71, Instrument: 2, Volume: 0x40, Effect: 0x0F, EffectParameter: 0x1F}) {
		t.Errorf("row 1 channel 1: got %+v", got)
	}
}

func TestReadTextErrors(t *testing.T) {
	tests := map[string]string{
		"missing channel":    strings.Replace(sample, "| A#5", "A#5", 1),
		"bad note":           strings.Replace(sample, "C-4", "H-4", 1),
		"bad hex":            strings.Replace(sample, "02 40", "0G 40", 1),
		"row out of order":   strings.Replace(sample, "Row 001", "Row 002", 1),
		"order out of range": strings.Replace(sample, "Order: 0", "Order: 3", 1),
		"missing blank line": strings.Replace(sample, "\n\nPattern", "\nPattern", 1),
	}

	for name, text := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ReadText(strings.NewReader(text)); err == nil {
				t.Errorf("expected an error")
			}
		})
	}
}
