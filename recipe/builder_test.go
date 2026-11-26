package recipe

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuilderInvalidValuesFailAtParse(t *testing.T) {
	tests := map[string]*Builder{
		"amplitude high": New(10).Sine(100, 2),
		"duty on sine": New(10).Sine(100, 1).Duty(0.5),
		"zero frames": New(0).Sine(100, 1),
		"zero frequency": New(10).Sine(0, 1),
		"no layers": New(10),
	}

	for name, b := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(strings.NewReader(b.String())); err == nil {
				t.Errorf("expected an error from %q", b.String())
			}
		})
	}
}

func TestBuilderOptionWithoutLayerPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("expected a panic")
		}
	}()

	New(64).Duty(0.5)
}

func TestBuilderRendersLikeText(t *testing.T) {
	built := New(64).Square(100, 1).Duty(0.25).String()

	pcmBuilt, _, err := Render(strings.NewReader(built))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	pcmText, _, err := Render(strings.NewReader("Frames: 64\nLayer: square 100 1 duty=0.25\n"))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !reflect.DeepEqual(pcmBuilt, pcmText) {
		t.Errorf("built and text recipes render differently")
	}
}

func parseText(t *testing.T, text string) *Recipe {
	t.Helper()

	rc, err := Parse(strings.NewReader(text))
	if err != nil {
		t.Fatalf("Parse(%q): %v", text, err)
	}

	return rc
}

func TestBuilderDefaultsToEightBit(t *testing.T) {
	if rc := parseText(t, New(64).Sine(100, 1).String()); rc.Sixteen {
		t.Errorf("got 16-bit, want 8-bit")
	}
}

func TestBuilderMatchesTextRecipe(t *testing.T) {
	const text string = "Frames: 256\n" + "Layer: square 261.34375 0.45 duty=0.25\n" + "Layer: sine 130.671875 0.35\n" + "Layer: triangle 523.4375 0.2 phase=0.5\n" + "Layer: saw 65.4 0.1\n" + "Layer: noise 1000 0.05\n"

	built := New(256).Square(261.34375, 0.45).Duty(0.25).Sine(130.671875, 0.35).Triangle(523.4375, 0.2).Phase(0.5).Saw(65.4, 0.1).Noise(1000, 0.05).String()
	if built != text {
		t.Fatalf("got\n%s\nwant\n%s", built, text)
	}

	if got, want := parseText(t, built), parseText(t, text); !reflect.DeepEqual(got, want) {
		t.Errorf("parsed recipes differ: got %+v, want %+v", got, want)
	}
}

func TestBuilderSixteenBit(t *testing.T) {
	if rc := parseText(t, New(64).Sixteen().Sine(100, 1).String()); !rc.Sixteen || rc.Frames != 64 {
		t.Errorf("got Sixteen %v, Frames %d", rc.Sixteen, rc.Frames)
	}
}
