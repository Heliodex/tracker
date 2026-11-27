package recipe

import (
	"strings"
	"testing"
)

func TestLoopableDetunedLayerFails(t *testing.T) {
	rc := parseText(t, "Frames: 256\nLayer: saw 262.4 0.3\nLayer: sine 261.34375 0.5\n")

	err := rc.Loopable()
	if err == nil {
		t.Fatalf("got nil, want an error")
	}

	if !strings.Contains(err.Error(), "262.4 Hz") || strings.Contains(err.Error(), "261.34375") {
		t.Errorf("error %q should name only the detuned layer", err)
	}
}

func TestLoopableIgnoresNoise(t *testing.T) {
	rc := parseText(t, "Frames: 300\nLayer: noise 1000 1\n")

	if err := rc.Loopable(); err != nil {
		t.Errorf("got %v, want nil", err)
	}
}

func TestLoopableOneShotCheckedOnlyWhenAsked(t *testing.T) {
	// Loopable reports every off-cycle layer, whether or not the sample will loop: the caller decides when it matters
	rc := parseText(t, "Frames: 2000\nLayer: sine 261.34375 0.5\n")

	if err := rc.Loopable(); err == nil {
		t.Errorf("got nil, want an error for 62.5 cycles")
	}
}

func TestLoopableWholeCycles(t *testing.T) {
	rc := parseText(t, "Frames: 256\nLayer: sine 261.34375 0.5\nLayer: square 522.6875 0.2 duty=0.25\nLayer: noise 1000 0.1\n")

	if err := rc.Loopable(); err != nil {
		t.Errorf("got %v, want nil", err)
	}
}
