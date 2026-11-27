package recipe

import (
	"fmt"
	"math"
	"strings"
)

// Loopable returns an error naming each layer that does not hold a whole number of cycles over the sample.
// A looping sample must hold whole cycles of every layer, or it clicks at the loop point. Noise has no cycle,
// so it is not checked. A one-shot sample never loops, so it does not need to pass this check.
func (rc *Recipe) Loopable() error {
	var bad []string

	for _, l := range rc.Layers {
		if l.wave == noise {
			continue
		}

		cycles := l.freq * float64(rc.Frames) / Rate
		if math.Abs(cycles-math.Round(cycles)) > 1e-9 {
			bad = append(bad, fmt.Sprintf("%g Hz is %.3f cycles", l.freq, cycles))
		}
	}

	if len(bad) == 0 {
		return nil
	}

	return fmt.Errorf("not whole cycles over %d frames, so the loop would click: %s", rc.Frames, strings.Join(bad, ", "))
}
