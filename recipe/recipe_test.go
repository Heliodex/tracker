package recipe

import (
	"strings"
	"testing"
)

func TestErrors(t *testing.T) {
	tests := map[string]string{
		"amplitude high": "Frames: 10\nLayer: sine 100 2\n",
		"bad bits": "Frames: 10\nBits: 24\nLayer: sine 100 1\n",
		"bad frames": "Frames: 0\nLayer: sine 100 1\n",
		"duty on sine": "Frames: 10\nLayer: sine 100 1 duty=0.5\n",
		"missing amp": "Frames: 10\nLayer: sine 100\n",
		"missing frames": "Layer: sine 100 1\n",
		"no colon": "Frames 10\n",
		"no layers": "Frames: 10\n",
		"too high": "Frames: 10\nLayer: sine 5000 1\n",
		"unknown key": "Frames: 10\nVolume: 1\nLayer: sine 100 1\n",
		"unknown option": "Frames: 10\nLayer: sine 100 1 pan=0.5\n",
		"unknown wave": "Frames: 10\nLayer: pulse 100 1\n",
		"zero frequency": "Frames: 10\nLayer: sine 0 1\n",
	}

	for name, text := range tests {
		t.Run(name, func(t *testing.T) {
			if _, _, err := Render(strings.NewReader(text)); err == nil {
				t.Errorf("expected an error")
			}
		})
	}
}

func render(t *testing.T, text string) ([]uint8, bool) {
	t.Helper()

	pcm, sixteen, err := Render(strings.NewReader(text))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	return pcm, sixteen
}

func TestCommentsAndBlankLines(t *testing.T) {
	if pcm, _ := render(t, "# a comment\n\nFrames: 10\n  Layer: sine 100 1  \n"); len(pcm) != 10 {
		t.Errorf("got %d bytes, want 10", len(pcm))
	}
}

func TestLayersAreSummedAndScaledDown(t *testing.T) {
	// two full-scale layers in phase would clip, so the sum is scaled to fit
	pcm, _ := render(t, "Frames: 64\nLayer: square 100 1\nLayer: square 100 1\n")

	for _, s := range pcm {
		if v := int8(s); v != 127 && v != -127 {
			t.Fatalf("got %d, want full scale 127 or -127", v)
		}
	}
}

func TestLengthAndBits(t *testing.T) {
	pcm, sixteen := render(t, "Frames: 100\nLayer: sine 100 1\n")
	if sixteen || len(pcm) != 100 {
		t.Errorf("8-bit: got %d bytes, sixteen %v", len(pcm), sixteen)
	}

	pcm, sixteen = render(t, "Frames: 100\nBits: 16\nLayer: sine 100 1\n")

	if !sixteen || len(pcm) != 200 {
		t.Errorf("16-bit: got %d bytes, sixteen %v", len(pcm), sixteen)
	}
}

func TestNoiseIsDeterministic(t *testing.T) {
	const text string = "Frames: 128\nLayer: noise 100 0.8\n"

	a, _ := render(t, text)
	if b, _ := render(t, text); string(a) != string(b) {
		t.Error("noise differs between renders")
	}
}

func TestSineStartsAtZeroAndIsSymmetric(t *testing.T) {
	// one cycle over 256 frames, so the sample loops
	pcm, _ := render(t, "Frames: 256\nLayer: sine 32.66796875 1\n")
	if pcm[0] != 0 {
		t.Errorf("first frame is %d, want 0", int8(pcm[0]))
	}

	if peak := int8(pcm[64]); peak < 125 {
		t.Errorf("quarter cycle is %d, want close to 127", peak)
	}
}

func TestSquareDuty(t *testing.T) {
	// a quarter of the cycle is high at duty 0.25
	pcm, _ := render(t, "Frames: 100\nLayer: square 83.63 1 duty=0.25\n")

	var high int

	for _, s := range pcm {
		if int8(s) > 0 {
			high++
		}
	}

	if high < 24 || high > 26 {
		t.Errorf("got %d high frames, want about 25", high)
	}
}
