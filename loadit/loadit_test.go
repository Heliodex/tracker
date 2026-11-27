package loadit

import (
	"bytes"
	"os"
	"testing"

	"github.com/Heliodex/tracker/load"
	"github.com/Heliodex/tracker/save"
)

func TestEffect(t *testing.T) {
	tests := []struct {
		name string
		cmd, p int
		effect, param uint8
		ok bool
	}{
		{"set speed", 1, 0x06, 0x0F, 0x06, true},
		{"speed past 31 is dropped", 1, 0x20, 0, 0, false},
		{"volume slide up", 4, 0x30, 0x0A, 0x30, true},
		{"fine volume up", 4, 0x3F, 0x0E, 0xA3, true},
		{"fine volume down", 4, 0xF2, 0x0E, 0xB2, true},
		{"fine portamento down", 5, 0xF4, 0x0E, 0x24, true},
		{"extra fine portamento is dropped", 5, 0xE4, 0, 0, false},
		{"pan 8 is scaled", 19, 0x8F, 0x08, 0xFF, true},
		{"note cut in S", 19, 0xC3, 0x0E, 0xC3, true},
		{"global volume is halved", 22, 128, 0x10, 64, true},
		{"tempo at 0x20 is kept", 20, 0x7D, 0x0F, 0x7D, true},
		{"tempo slide is dropped", 20, 0x05, 0, 0, false},
		{"tremor is dropped", 9, 0x33, 0, 0, false},
	}

	for _, tt := range tests {
		if eff, param, ok := effect(tt.cmd, tt.p); ok != tt.ok || eff != tt.effect || param != tt.param {
			t.Errorf("%s: effect(%d, %#x) = %#x %#x %v, want %#x %#x %v", tt.name, tt.cmd, tt.p, eff, param, ok, tt.effect, tt.param, tt.ok)
		}
	}
}

func TestMarmelade(t *testing.T) {
	const path = "../marmelade_dream.it"

	data, err := os.ReadFile(path)
	if err != nil {
		t.Skip("no sample module:", err)
	}

	tf, warnings, err := Read(data)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	for _, w := range warnings {
		t.Log("warning:", w)
	}

	f, err := tf.ToFile()
	if err != nil {
		t.Fatalf("to file: %v", err)
	}

	var b bytes.Buffer

	if err := save.Write(&b, f); err != nil {
		t.Fatalf("save: %v", err)
	}

	// the XM must read back through the XM loader
	if _, err := load.Read(bytes.NewBuffer(b.Bytes())); err != nil {
		t.Fatalf("reload: %v", err)
	}

	if int(tf.NumChannels) > maxXMChans {
		t.Errorf("%d channels, XM has at most %d", tf.NumChannels, maxXMChans)
	}
}

func TestNoteCut(t *testing.T) {
	c := &conv{
		rep: &report{
			counts: map[string]int{},
		},
	}
	cd := c.channelData(cell{
		note: 254,
		ins: 3,
		vol: -1,
	})

	if cd.Note != 0 || cd.Effect != 0x0E || cd.EffectParameter != 0xC0 || cd.Instrument != 3 {
		t.Errorf("note cut = %+v, want EC0 with no note", cd)
	}

	cd = c.channelData(cell{
		note: 60,
		vol: -1,
		cmd: 1,
		param: 0x06,
	})

	if cd.Note != 61 {
		t.Errorf("C-5 = note %d, want 61", cd.Note)
	}
}

func TestRelativeNote(t *testing.T) {
	// C5Speed of 8363 plays one octave below C-5, so the relative note is -12
	for _, tt := range []struct {
		c5 int
		rel, fine int8
	}{
		{8363, -12, 0},
		{16726, 0, 0},
		{44100, 17, -28}, // about 16.78 semitones above the XM base: 17 with finetune -27.8, rounded
	} {
		h := make([]byte, sampleLen)

		copy(h, "IMPS")

		h[0x12] = smpExists
		h[0x30] = 4
		h[0x3C] = byte(tt.c5)
		h[0x3D] = byte(tt.c5 >> 8)
		h[0x3E] = byte(tt.c5 >> 16)
		d := append(make([]byte, 0x200), h...)
		d = append(d, 0, 0, 0, 0)
		c := &conv{
			d: d,
			rep: &report{
				counts: map[string]int{},
			},
			smpOff: []int{0x200},
		}
		s, _, ok, err := c.sample(1, 0)

		if err != nil || !ok {
			t.Fatalf("sample for %d Hz: ok %v err %v", tt.c5, ok, err)
		}

		if s.RelativeNote != tt.rel || s.Finetune != tt.fine {
			t.Errorf("C5Speed %d: relative %d finetune %d, want %d %d", tt.c5, s.RelativeNote, s.Finetune, tt.rel, tt.fine)
		}
	}
}

func TestVolumeColumn(t *testing.T) {
	tests := []struct {
		v int
		want uint8
		ok bool
	}{
		{0, 0x10, true},
		{64, 0x50, true},
		{65, 0x70, true},
		{128, 0xC0, true},
		{192, 0xCF, true},
		{193, 0xF0, true},
		{203, 0xB0, true},
		{110, 0, false},
	}

	for _, tt := range tests {
		if got, ok := volumeColumn(tt.v); got != tt.want || ok != tt.ok {
			t.Errorf("volumeColumn(%d) = %#x %v, want %#x %v", tt.v, got, ok, tt.want, tt.ok)
		}
	}
}
