// Command track15 writes "marmelade rain", an atmospheric jungle track, as an XM module, song.xm, in this folder.
//
// The track takes its cue from marmelade_dream.xm: a hazy, vocal-heavy breakbeat at a slow jungle tempo, stitched
// together from chopped breaks and long detuned pads. Every sample here is synthesised in this file rather than taken
// from that module, so the folder holds only this program.
//
// Build it from anywhere inside the repository with go run ./tracks/track15, or with tracker build.
package main

import (
	"encoding/binary"
	"log"
	"math"

	"github.com/Heliodex/tracker/loadtext"
	"github.com/Heliodex/tracker/song"
	"github.com/Heliodex/tracker/types"
	"github.com/Heliodex/tracker/wav"
)

type (
	// Sample synthesis

	rng struct {
		state uint64
	}

	// marmelade rain – DeepSeek V4.1 Flash

	// chord is one bar of harmony: pads, a bass root and fifth, an arpeggio and a lead phrase
	chord struct {
		pad [3]string
		bass, fifth string
		arp, lead [8]string
		minor bool
	}

	// kit holds the instruments, numbered as they are added to the song
	kit struct {
		pad, choir, air, lead, bell, vox, bass, kick, snare, hat, open, chop, breakLoop, crash, riser, boom, vinyl *song.Instrument
	}

	// pattern wraps a song pattern with the instruments to play it with
	pattern struct {
		p *song.Pattern
		k *kit
	}
)

// channels, in the order they appear in the tracker
const (
	chPadLo = iota
	chPadMid
	chPadHi
	chBreath
	chAir
	chLead
	chBell
	chVox
	chBass
	chKick
	chSnare
	chHat
	chOpen
	chChopL
	chChopR
	chBreak
	chFX
	chVinyl
	numChannels
)

const (
	// rows is a full section: eight bars of sixteen rows
	rows = 128
	// the rows are sixteenth notes at 168 BPM, which gives a slow jungle roll
	speed = 6
	tempo = 168
	// base is C-4
	base = 261.6255653005986
)

var (
	// rate is the rate at which C-4 plays
	rate = 8363.0
	cm = chord{
		pad: [3]string{"C-4", "D#4", "G-4"},
		bass: "C-2",
		fifth: "G-2",
		arp: [8]string{"C-5", "D#5", "G-5", "C-6", "G-5", "D#5", "C-5", "G-4"},
		lead: [8]string{"G-4", "C-5", "D#5", "G-5", "D#5", "C-5", "G-4", "A#4"},
		minor: true,
	}
	ab = chord{
		pad: [3]string{"G#3", "C-4", "D#4"},
		bass: "G#2",
		fifth: "D#3",
		arp: [8]string{"G#4", "C-5", "D#5", "G#5", "D#5", "C-5", "G#4", "D#4"},
		lead: [8]string{"C-5", "D#5", "G-5", "C-6", "G-5", "D#5", "C-5", "D#5"},
	}
	eb = chord{
		pad: [3]string{"D#4", "G-4", "A#4"},
		bass: "D#2",
		fifth: "A#2",
		arp: [8]string{"D#5", "G-5", "A#5", "D#6", "A#5", "G-5", "D#5", "A#4"},
		lead: [8]string{"A#4", "D#5", "G-5", "A#5", "G-5", "D#5", "A#4", "G-4"},
	}
	bb = chord{
		pad: [3]string{"A#3", "D-4", "F-4"},
		bass: "A#2",
		fifth: "F-2",
		arp: [8]string{"A#4", "D-5", "F-5", "A#5", "F-5", "D-5", "A#4", "F-4"},
		lead: [8]string{"D-5", "F-5", "A#5", "F-5", "D-5", "C-5", "A#4", "F-4"},
	}
	fm = chord{
		pad: [3]string{"F-3", "G#3", "C-4"},
		bass: "F-2",
		fifth: "C-3",
		arp: [8]string{"F-4", "G#4", "C-5", "F-5", "C-5", "G#4", "F-4", "C-4"},
		lead: [8]string{"C-5", "F-5", "G#5", "C-6", "G#5", "F-5", "C-5", "G#4"},
		minor: true,
	}
	gMaj = chord{
		pad: [3]string{"G-3", "B-3", "D-4"},
		bass: "G-2",
		fifth: "D-3",
		arp: [8]string{"G-4", "B-4", "D-5", "G-5", "D-5", "B-4", "G-4", "D-4"},
		lead: [8]string{"D-5", "G-5", "B-5", "D-6", "B-5", "G-5", "D-5", "B-4"},
	}

	// i - VI - III - VII, the dark four-bar loop the track keeps returning to
	verse = []chord{cm, ab, eb, bb}
	// the same shape pushed towards the dominant, for breaks and builds
	turn = []chord{fm, ab, bb, gMaj}
)

// addWave adds a sine at a whole number of cycles over the buffer, so a loop joins without a click
func addWave(out []float64, cycles, amp, phase float64) {
	for i := range out {
		out[i] += amp * math.Sin(2*math.Pi*(cycles*float64(i)/float64(len(out))+phase))
	}
}

// arpType gives the arpeggio that turns a mono vocal stab into the chord: minor or major
func arpType(c chord) int {
	if c.minor {
		return 0x037
	}

	return 0x047
}

// at returns the row of row r in bar b
func at(bar, row int) int {
	return bar*song.RowsPerBar + row
}

// barChord returns the chord of bar b from a four-bar progression
func barChord(cs []chord, b int) chord {
	return cs[b%len(cs)]
}

// effect sets an effect column directly, for the effects the song package does not wrap
func effect(eff, param int) song.Opt {
	return func(c *loadtext.ChannelData) error {
		c.Effect, c.EffectParameter = uint8(eff), uint8(param)

		return nil
	}
}

// fadeOut fades the last n samples to zero, so a one-shot ends without a click
func fadeOut(samples []float64, n int) []float64 {
	n = min(n, len(samples))

	for i := range n {
		samples[len(samples)-1-i] *= float64(i) / float64(n)
	}

	return samples
}

// formantGain shapes a harmonic series like a voice
func formantGain(hz float64) float64 {
	g := 0.10

	for _, f := range []struct {
		centre, width, gain float64
	}{
		{700, 380, 0.80},
		{1220, 520, 0.55},
		{2600, 950, 0.45},
	} {
		g += f.gain * math.Exp(-math.Pow((hz-f.centre)/f.width, 2))
	}

	return g
}

// highpass is a first difference, which tilts noise towards the top end
func highpass(in []float64) []float64 {
	out := make([]float64, len(in))

	var prev float64

	for i, v := range in {
		out[i] = v - prev
		prev = v
	}

	return out
}

// lowpass is a one-pole filter
func lowpass(in []float64, a float64) []float64 {
	out := make([]float64, len(in))

	var y float64

	for i, v := range in {
		y += a * (v - y)
		out[i] = y
	}

	return out
}

// normalise scales a buffer to full scale, so every sample starts at a predictable level
func normalise(samples []float64) []float64 {
	var peak float64

	for _, s := range samples {
		peak = math.Max(peak, math.Abs(s))
	}

	if peak == 0 {
		return samples
	}

	scale := 0.96 / peak

	for i := range samples {
		samples[i] *= scale
	}

	return samples
}

// bassSample is two detuned saws with a sub, tuned to C-4
func bassSample(frames int) []float64 {
	out := make([]float64, frames)
	c := math.Round(base * float64(frames) / rate)

	for h := 1; h <= 10; h++ {
		amp := 1 / math.Pow(float64(h), 2.2)

		addWave(out, c*float64(h), amp, 0)

		if h <= 4 {
			addWave(out, c*float64(h)+1, amp*0.55, 0.25)
		}
	}

	addWave(out, math.Round(c/2), 0.70, 0)
	addWave(out, math.Round(c/4), 0.18, 0)

	return normalise(out)
}

// bellSample is a struck bell with inharmonic partials, tuned to C-4
func bellSample(sec float64) []float64 {
	n := int(sec * rate)
	out := make([]float64, n)

	for _, p := range []struct {
		ratio, amp, decay float64
	}{
		{1, 0.55, 2.2},
		{2.01, 0.20, 3.0},
		{2.76, 0.16, 3.6},
		{5.40, 0.08, 5.0},
		{8.93, 0.04, 7.0},
	} {
		for i := range out {
			t := float64(i) / rate
			env := math.Min(t/0.004, 1) * math.Exp(-p.decay*t)
			out[i] += p.amp * env * math.Sin(2*math.Pi*base*p.ratio*t)
		}
	}

	return normalise(out)
}

// boomSample is a low sine drop for the last transition
func boomSample() []float64 {
	n := int(1.2 * rate)
	out := make([]float64, n)

	var phase float64

	for i := range out {
		t := float64(i) / rate
		f := 34 + 70*math.Exp(-t*2.2)
		phase += f / rate
		out[i] = math.Sin(2*math.Pi*phase) * math.Exp(-t*3)
	}

	return normalise(fadeOut(out, 300))
}

// breakSample stitches the drum one-shots into one bar at the song tempo
func breakSample(kick, snare, hat, open []float64) []float64 {
	n := int(math.Round(60.0 / float64(tempo) * 4 * rate))
	out := make([]float64, n)
	step := float64(n) / 16
	pos := func(s int) int {
		return int(math.Round(float64(s) * step))
	}
	mix := func(at int, src []float64, gain float64) {
		for i, v := range src {
			j := at + i
			if j >= len(out) {
				break
			}

			out[j] += v * gain
		}
	}

	for s := range 16 {
		gain := 0.26

		if s%4 == 0 {
			gain = 0.46
		}

		mix(pos(s), hat, gain)
	}

	for _, h := range []struct {
		step int
		gain float64
	}{
		{0, 0.95},
		{3, 0.52},
		{7, 0.70},
		{10, 0.86},
		{14, 0.48},
	} {
		mix(pos(h.step), kick, h.gain)
	}

	for _, h := range []struct {
		step int
		gain float64
	}{
		{4, 0.88},
		{12, 1.0},
		{15, 0.32},
	} {
		mix(pos(h.step), snare, h.gain)
	}

	mix(pos(6), open, 0.38)
	mix(pos(14), open, 0.34)

	return normalise(fadeOut(out, 160))
}

// brighten mixes a looping air sample into a pad, adding top end without adding mud
func brighten(base, air []float64, amount float64) []float64 {
	out := append([]float64(nil), base...)

	for i := range out {
		out[i] += air[i%len(air)] * amount
	}

	return normalise(out)
}

// chopSample lifts a short snare slice out of the break, for the pitched chop line
func chopSample(breakBuf []float64) []float64 {
	step := float64(len(breakBuf)) / 16
	length := int(0.24 * rate)
	start := int(4 * step)

	if start+length > len(breakBuf) {
		length = len(breakBuf) - start
	}

	out := append([]float64(nil), breakBuf[start:start+length]...)
	fadeIn := min(48, len(out))
	fadeOutLen := min(300, len(out))

	for i := range fadeIn {
		out[i] *= float64(i) / float64(fadeIn)
	}

	for i := range fadeOutLen {
		out[len(out)-1-i] *= float64(i) / float64(fadeOutLen)
	}

	return normalise(out)
}

// leadSample is a hollow reed with a little detune, tuned to C-4
func leadSample(frames int) []float64 {
	out := make([]float64, frames)
	c := math.Round(base * float64(frames) / rate)

	addWave(out, c, 0.50, 0)
	addWave(out, c+1, 0.18, 0.2)
	addWave(out, c*2, 0.10, 0)
	addWave(out, c*3, 0.18, 0.1)
	addWave(out, c*4, 0.05, 0)
	addWave(out, c*5, 0.08, 0.3)
	addWave(out, c*8, 0.04, 0)
	addWave(out, c*9, 0.03, 0.2)

	return normalise(out)
}

// air adds the breathy high texture, alternating sides each bar
func (w pattern) air(b int, note string, vol int) {
	pan := 0x50

	if b%2 == 1 {
		pan = 0xB0
	}

	w.p.Hit(at(b, 0), chAir, note, w.k.air, song.VolOpts(vol, song.Pan(pan))...)
}

// bass plays a reese bass: a long root, or a root that slides to the fifth in the second half of the bar
func (w pattern) bass(b int, c chord, style, vol int) {
	w.p.Hit(at(b, 0), chBass, c.bass, w.k.bass, song.VolOpts(vol)...)

	if style == 0 {
		return
	}

	if style == 1 {
		w.p.Hit(at(b, 10), chBass, c.fifth, w.k.bass, song.VolOpts(vol-0x06, effect(0x03, 0x06))...)
		return
	}

	for _, h := range []struct {
		row int
		note string
		vol int
	}{
		{6, c.bass, vol - 0x04},
		{10, c.fifth, vol - 0x02},
		{14, c.bass, vol - 0x08},
	} {
		w.p.Hit(at(b, h.row), chBass, h.note, w.k.bass, song.VolOpts(h.vol, effect(0x03, 0x08))...)
	}
}

// bellLine plays a bell arpeggio, panning alternate notes
func (w pattern) bellLine(b int, notes [8]string, step, vol int) {
	for i, note := range notes {
		pan := 0x58

		if i%2 == 1 {
			pan = 0xA8
		}

		w.p.Hit(at(b, i*step), chBell, note, w.k.bell, song.VolOpts(vol, song.Pan(pan))...)
	}
}

// breakBeat starts the sampled break at the top of a bar
func (w pattern) breakBeat(b, vol int) {
	w.p.Hit(at(b, 0), chBreak, "C-4", w.k.breakLoop, song.VolOpts(vol)...)
}

// choir adds a long formant voice above the pads
func (w pattern) choir(b int, c chord, vol int) {
	w.p.Hit(at(b, 0), chBreath, c.pad[1], w.k.choir, song.VolOpts(vol, song.Pan(0x70))...)
	w.p.Hit(at(b, 8), chBreath, "###", nil)
}

// chops plays a pitched slice of the break. The first note sets the pan, the rest slide into place.
func (w pattern) chops(b int, notes [8]string, vol, pan, slide, ch int) {
	for i, note := range notes {
		opts := []song.Opt{song.Vol(vol)}

		if i == 0 {
			opts = append(opts, song.Pan(pan))
		} else {
			opts = append(opts, effect(0x03, slide))
		}

		if i%2 == 0 {
			opts[0] = song.Vol(vol + 0x06)
		}

		w.p.Hit(at(b, i*2), ch, note, w.k.chop, opts...)
	}
}

// fill rolls the snare over the last beat of a bar
func (w pattern) fill(b int) {
	for i, r := range []int{12, 13, 14, 15} {
		w.p.Hit(at(b, r), chSnare, "C-4", w.k.snare, song.VolOpts(0x24+i*0x08, song.Pan(0x78+i*0x08))...)
	}
}

// fx plays a one-shot on the effects channel
func (w pattern) fx(b, row int, note string, ins *song.Instrument, vol int) {
	w.p.Hit(at(b, row), chFX, note, ins, song.VolOpts(vol)...)
}

// hats plays closed hats every step rows, with accents on the beats
func (w pattern) hats(b, step, baseVol, accentVol int) {
	for i, r := range song.Steps(0, song.RowsPerBar, step) {
		vol, pan := baseVol, 0x70

		if r%4 == 0 {
			vol = accentVol
		}

		if i%2 == 1 {
			pan = 0xA0
		}

		w.p.Hit(at(b, r), chHat, "C-4", w.k.hat, song.VolOpts(vol, song.Pan(pan))...)
	}
}

// drums plays a programmed jungle break. Style 1 is sparse, 2 is a full break and 3 is busy.
func (w pattern) drums(b, style int) {
	if style == 0 {
		return
	}

	step := 4

	if style >= 2 {
		step = 2
	}

	w.hats(b, step, 0x30, 0x44)

	if style >= 2 {
		for i, r := range []int{1, 3, 5, 7, 9, 11, 13, 15} {
			pan := 0x58

			if i%2 == 0 {
				pan = 0xA8
			}

			w.p.Hit(at(b, r), chHat, "C-4", w.k.hat, song.VolOpts(0x24, song.Pan(pan))...)
		}

		w.p.Hit(at(b, 6), chOpen, "C-4", w.k.open, song.VolOpts(0x38, song.Pan(0x60))...)
		w.p.Hit(at(b, 14), chOpen, "C-4", w.k.open, song.VolOpts(0x40, song.Pan(0xA0))...)
	}

	kicks := []int{0}

	if style >= 2 {
		kicks = []int{0, 3, 10}
	}

	if style >= 3 {
		kicks = []int{0, 3, 7, 10, 14}
	}

	for _, r := range kicks {
		vol := 0x34

		if r == 0 {
			vol = 0x50
		}

		w.p.Hit(at(b, r), chKick, "C-4", w.k.kick, song.VolOpts(vol)...)
	}

	w.p.Hit(at(b, 4), chSnare, "C-4", w.k.snare, song.VolOpts(0x48, song.Pan(0x88))...)
	w.p.Hit(at(b, 12), chSnare, "C-4", w.k.snare, song.VolOpts(0x50, song.Pan(0x78))...)

	if style >= 2 {
		w.p.Hit(at(b, 15), chSnare, "C-4", w.k.snare, song.VolOpts(0x22, song.Pan(0x90))...)
	}

	if style >= 3 {
		w.p.Hit(at(b, 7), chSnare, "C-4", w.k.snare, song.VolOpts(0x14, song.Pan(0x98))...)
		w.p.Hit(at(b, 14), chSnare, "C-4", w.k.snare, song.VolOpts(0x1A, song.Pan(0x68))...)
	}
}

// leadLine plays the lead phrase, with a light vibrato on every note
func (w pattern) leadLine(b int, notes [8]string, step, vol int) {
	for i, note := range notes {
		w.p.Hit(at(b, i*step), chLead, note, w.k.lead, song.VolOpts(vol, song.Vibrato(4, 2))...)
	}
}

// pads holds a chord on the three pad voices, spread across the stereo field
func (w pattern) pads(b int, c chord, vol int) {
	pans := [3]int{0x48, 0x80, 0xB8}

	for i, note := range c.pad {
		w.p.Hit(at(b, 0), chPadLo+i, note, w.k.pad, song.VolOpts(vol, song.Pan(pans[i]))...)
		w.p.Hit(at(b, 8), chPadLo+i, "###", nil)
	}
}

// outro lets the last chord ring out under the vinyl
func outro(s *song.Song, k *kit) *song.Pattern {
	// the final C minor, held while the vinyl runs out
	const last int = 3

	w := pattern{s.Pattern(rows / 2), k}

	for b := range 3 {
		c := barChord(verse, b)

		w.pads(b, c, 0x1e)
		w.choir(b, c, 0x1a)

		if b == 0 {
			w.fx(b, 0, "C-4", k.crash, 0x40)
			w.bass(b, c, 0, 0x2a)
		}
	}

	w.pads(last, cm, 0x1e)
	w.choir(last, cm, 0x1c)
	w.bass(last, cm, 0, 0x28)
	w.bellLine(last, cm.arp, 2, 0x26)

	// ease the last chord and the vinyl down over the final rows
	for r := 10; r < 15; r++ {
		for _, ch := range []int{chPadLo, chPadMid, chPadHi, chBreath, chBass, chAir, chVinyl} {
			w.p.Hit(at(last, r), ch, "...", nil, effect(0x0A, 0x01))
		}
	}

	for _, ch := range []int{
		chLead,
		chBell,
		chVox,
		chBass,
		chKick,
		chSnare,
		chHat,
		chOpen,
		chChopL,
		chChopR,
		chBreak,
		chAir,
		chBreath,
		chVinyl,
		chFX,
	} {
		w.p.Hit(rows/2-1, ch, "###", nil)
	}

	return w.p
}

// roll is a snare crescendo from row start up to, but not including, row end
func (w pattern) roll(b, start, end, fromVol, toVol int) {
	for r := start; r < end; r++ {
		t := float64(r-start) / float64(max(1, end-start-1))
		vol := fromVol + int(t*float64(toVol-fromVol))

		w.p.Hit(at(b, r), chSnare, "C-4", w.k.snare, song.VolOpts(vol, song.Pan(0x70+int(t*0x20)))...)
	}
}

// intro opens with pads and breath, then lets the break surface
func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	w.p.Hit(0, chVinyl, "C-4", k.vinyl, song.VolOpts(0x28)...)

	for b := range rows / song.RowsPerBar {
		c := barChord(verse, b)

		w.pads(b, c, 0x1a)

		if b%2 == 0 {
			w.choir(b, c, 0x16)
		}

		if b >= 2 {
			w.air(b, c.pad[2], 0x24)
		}

		if b >= 4 {
			w.bellLine(b, c.arp, 2, 0x1e)
		}
	}

	w.breakBeat(6, 0x18)
	w.fx(7, 0, "C-4", k.riser, 0x30)
	w.roll(7, 8, 16, 0x14, 0x30)

	return w.p
}

// stutter speeds the rows up over the last beat, then restores the song speed
func (w pattern) stutter(b int) {
	w.p.Hit(at(b, 8), chFX, "...", nil, effect(0x0F, 4))
	w.p.Hit(at(b, 12), chFX, "...", nil, effect(0x0F, 3))
	w.p.Hit(at(b, 14), chFX, "...", nil, effect(0x0F, 6))
}

// build brings the drums and a snare roll back
func build(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b := range rows / song.RowsPerBar {
		c := barChord(turn, b)

		w.pads(b, c, 0x24)
		w.choir(b, c, 0x1e)
		w.bass(b, c, 1, 0x32)

		style := 1

		if b >= 2 {
			style = 2
		}

		if b >= 4 {
			style = 3
		}

		w.drums(b, style)

		if b >= 4 {
			w.chops(b, c.arp, 0x40, 0x60, 0x18, chChopL)
		}

		if b >= 6 {
			w.roll(b, 0, 16, 0x16, 0x38)
		}
	}

	w.fx(7, 0, "C-5", k.riser, 0x44)
	w.stutter(7)

	return w.p
}

// voxStab drops a short vocal chord on an offbeat
func (w pattern) voxStab(b, row int, note string, vol, arp int) {
	w.p.Hit(at(b, row), chVox, note, w.k.vox, song.VolOpts(vol, effect(0x00, arp))...)
}

// reverseSamples returns the buffer reversed
func reverseSamples(in []float64) []float64 {
	out := make([]float64, len(in))

	for i := range in {
		out[i] = in[len(in)-1-i]
	}

	return out
}

// next returns a deterministic value from -1 to 1
func (r *rng) next() float64 {
	r.state += 0x9E3779B97F4A7C15
	z := r.state
	z = (z ^ z>>30) * 0xBF58476D1CE4E5B9
	z = (z ^ z>>27) * 0x94D049BB133111EB
	z ^= z >> 31

	return float64(z)/float64(math.MaxUint64)*2 - 1
}

// crashSample is a bright metallic wash
func crashSample(sec float64, seed uint64) []float64 {
	n := int(sec * rate)
	r := &rng{
		state: seed,
	}
	raw := make([]float64, n)

	for i := range raw {
		raw[i] = r.next()
	}

	hp := highpass(raw)
	out := make([]float64, n)

	for i := range out {
		t := float64(i) / rate
		out[i] = hp[i] * math.Exp(-t*2.6) * 0.85

		for _, hz := range []float64{3150, 3480, 3760, 4050} {
			out[i] += 0.05 * math.Sin(2*math.Pi*hz*t) * math.Exp(-t*3.2)
		}
	}

	return normalise(fadeOut(out, 300))
}

// hatSample is a short closed hat
func hatSample(seed uint64) []float64 {
	n := int(0.11 * rate)
	r := &rng{
		state: seed,
	}
	raw := make([]float64, n)

	for i := range raw {
		raw[i] = r.next()
	}

	hp := highpass(highpass(raw))
	out := make([]float64, n)

	for i := range out {
		t := float64(i) / rate
		out[i] = hp[i] * math.Exp(-t*58)
	}

	return normalise(fadeOut(out, 60))
}

// kickSample is a deep pitch-swept kick with a click
func kickSample(seed uint64) []float64 {
	n := int(0.55 * rate)
	out := make([]float64, n)
	r := &rng{
		state: seed,
	}

	var phase float64

	for i := range out {
		t := float64(i) / rate
		f := 46 + 170*math.Exp(-t*34)
		phase += f / rate
		body := math.Sin(2*math.Pi*phase) * math.Exp(-t*8)
		click := r.next() * math.Exp(-t*70) * 0.45
		out[i] = body + click
	}

	return normalise(fadeOut(out, 120))
}

// openHatSample is a longer, softer hat
func openHatSample(seed uint64) []float64 {
	n := int(0.48 * rate)
	r := &rng{
		state: seed,
	}
	raw := make([]float64, n)

	for i := range raw {
		raw[i] = r.next()
	}

	hp := highpass(highpass(raw))
	out := make([]float64, n)

	for i := range out {
		t := float64(i) / rate
		out[i] = hp[i] * math.Exp(-t*12)
	}

	return normalise(fadeOut(out, 250))
}

// next01 returns a deterministic value from 0 to 1
func (r *rng) next01() float64 {
	return r.next()*0.5 + 0.5
}

// airSample is a loop of many high partials: filtered noise that still loops without a click
func airSample(frames int, seed uint64) []float64 {
	out := make([]float64, frames)
	r := &rng{
		state: seed,
	}

	for range 320 {
		hz := 500 + r.next01()*3450
		cycles := math.Max(1, math.Round(hz*float64(frames)/rate))
		amp := (0.03 + 0.09*r.next01()) * (0.25 + math.Exp(-math.Pow((hz-2800)/1600, 2)))

		addWave(out, cycles, amp, r.next01())
	}

	return normalise(out)
}

// padSample is a lush detuned pad, tuned to C-4
func padSample(frames int, formant bool, seed uint64) []float64 {
	out := make([]float64, frames)
	r := &rng{
		state: seed,
	}
	c := math.Round(base * float64(frames) / rate)

	for h := 1; h <= 15; h++ {
		hz := base * float64(h)
		amp := 1 / math.Pow(float64(h), 1.05)

		if formant {
			amp *= formantGain(hz)
		} else if hz > 300 && hz < 1600 {
			// scoop the muddy lower mids, leaving the fundamental and the air above
			amp *= 0.30
		} else if hz > 2200 {
			amp *= 2.2
		}

		addWave(out, c*float64(h), amp, r.next01())

		if h <= 4 {
			addWave(out, c*float64(h)+1, amp*0.34, r.next01())
		}
	}

	addWave(out, math.Round(c/2), 0.18, 0)

	return normalise(out)
}

// choirSample is a formant pad with a slow swell. Every layer is whole cycles, so it loops cleanly.
func choirSample(frames int) []float64 {
	out := padSample(frames, true, 0xC401)

	for i := range out {
		out[i] *= 1 + 0.08*math.Sin(2*math.Pi*float64(i)/float64(frames))
	}

	return normalise(out)
}

// snareSample is a noisy snare with a tuned body
func snareSample(seed uint64) []float64 {
	n := int(0.42 * rate)
	r := &rng{
		state: seed,
	}
	raw := make([]float64, n)

	for i := range raw {
		raw[i] = r.next()
	}

	hp := highpass(raw)
	lp := lowpass(raw, 0.18)
	out := make([]float64, n)

	for i := range out {
		t := float64(i) / rate
		noise := hp[i]*math.Exp(-t*15)*0.70 + lp[i]*math.Exp(-t*9)*0.35
		body := math.Sin(2*math.Pi*185*t) * math.Exp(-t*20) * 0.38
		top := math.Sin(2*math.Pi*330*t) * math.Exp(-t*28) * 0.14
		out[i] = noise + body + top
	}

	return normalise(fadeOut(out, 150))
}

// toWAV turns floating-point samples into a 16-bit mono WAV file
func toWAV(samples []float64) []byte {
	pcm := make([]byte, len(samples)*2)

	for i, s := range samples {
		v := int16(math.Round(math.Max(-1, math.Min(1, s)) * 32767))

		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(v))
	}

	file, err := wav.Encode(pcm, true)
	if err != nil {
		log.Fatal(err)
	}

	return file
}

// upOctave returns note raised an octave, for the right-hand chop line
func upOctave(note string) string {
	n, err := types.ParseNote(note)
	if err != nil {
		panic(err)
	}

	return types.Note(int(n) + 12).String()
}

// breakdown strips the drums away and lets the voices carry the turn
func breakdown(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b := range rows / song.RowsPerBar {
		c := barChord(turn, b)

		w.pads(b, c, 0x1a)
		w.choir(b, c, 0x22)
		w.air(b, c.pad[2], 0x34)

		if b < 2 {
			w.bass(b, c, 0, 0x26)
		}

		if b >= 2 {
			w.bellLine(b, c.arp, 2, 0x2A)
		}

		if b >= 4 {
			w.voxStab(b, 8, upOctave(c.pad[0]), 0x40, arpType(c))
		}
	}

	w.fx(7, 0, "C-5", k.riser, 0x40)
	w.roll(7, 12, 16, 0x16, 0x30)

	return w.p
}

// reverseUp reverses a phrase and lifts it an octave, so the two chop channels answer each other
func reverseUp(notes [8]string) (out [8]string) {
	for i := range notes {
		out[i] = upOctave(notes[len(notes)-1-i])
	}

	return
}

// dropA is the sampled-break section
func (w pattern) dropA(lead, fillBar bool) {
	for b := range rows / song.RowsPerBar {
		c := barChord(verse, b)

		w.pads(b, c, 0x24)
		w.choir(b, c, 0x1e)

		if b%2 == 1 {
			w.air(b, c.pad[1], 0x2C)
		}

		w.bass(b, c, 1, 0x34)
		w.breakBeat(b, 0x40)

		if b >= 4 {
			w.chops(b, c.arp, 0x40, 0x60, 0x10, chChopL)
			w.chops(b, reverseUp(c.arp), 0x3C, 0xA0, 0x14, chChopR)
			w.bellLine(b, c.arp, 2, 0x1e)
		}

		if lead && b >= 4 {
			w.leadLine(b, c.lead, 2, 0x26)
		}

		if b%2 == 0 {
			w.voxStab(b, 14, upOctave(c.pad[0]), 0x38, arpType(c))
		}

		if b%4 == 0 {
			w.fx(b, 0, "C-4", w.k.crash, 0x38)
		}
	}

	if fillBar {
		w.fill(7)
	}
}

// dropB is the programmed-break section, drier and more forward
func (w pattern) dropB(lead, fillBar bool) {
	for b := range rows / song.RowsPerBar {
		c := barChord(verse, b)

		w.pads(b, c, 0x26)
		w.choir(b, c, 0x20)
		w.bass(b, c, 2, 0x38)
		w.drums(b, 2)
		w.chops(b, c.arp, 0x44, 0x60, 0x18, chChopL)
		w.chops(b, reverseUp(c.arp), 0x3E, 0xA0, 0x1C, chChopR)
		w.bellLine(b, c.arp, 2, 0x26)

		if lead {
			w.leadLine(b, c.lead, 2, 0x2a)
		} else if b >= 4 {
			w.leadLine(b, c.lead, 2, 0x22)
		}

		if b%2 == 1 {
			w.voxStab(b, 14, upOctave(c.pad[0]), 0x3C, arpType(c))
		}

		if b%4 == 0 {
			w.fx(b, 0, "C-4", w.k.crash, 0x38)
		}
	}

	if fillBar {
		w.fill(7)
	}
}

// dropC is the final, full arrangement
func (w pattern) dropC(final bool) {
	for b := range rows / song.RowsPerBar {
		c := barChord(verse, b)

		w.pads(b, c, 0x28)
		w.choir(b, c, 0x22)
		w.air(b, c.pad[1], 0x30)
		w.bass(b, c, 2, 0x3c)
		w.breakBeat(b, 0x2A)
		w.drums(b, 3)
		w.chops(b, c.arp, 0x48, 0x60, 0x20, chChopL)
		w.chops(b, reverseUp(c.arp), 0x42, 0xA0, 0x24, chChopR)
		w.bellLine(b, c.arp, 2, 0x2C)
		w.leadLine(b, c.lead, 2, 0x2e)

		if b%2 == 1 {
			w.voxStab(b, 6, upOctave(c.pad[0]), 0x40, arpType(c))
			w.voxStab(b, 14, upOctave(c.pad[0]), 0x3A, arpType(c))
		}

		w.fx(b, 0, "C-4", w.k.crash, 0x40)
	}

	w.fill(7)

	if final {
		w.fx(7, 8, "C-4", w.k.boom, 0x44)
	}
}

// vinylSample is a quiet crackle loop
func vinylSample(frames int, seed uint64) []float64 {
	out := make([]float64, frames)
	r := &rng{
		state: seed,
	}

	// a periodic rumble, so the loop joins without a click
	for range 220 {
		hz := 40 + r.next01()*900
		cycles := math.Max(1, math.Round(hz*float64(frames)/rate))
		amp := 0.02 + 0.06*r.next01()

		addWave(out, cycles, amp, r.next01())
	}

	// sparse crackles, kept away from the loop boundary
	for range 48 {
		pos := 200 + int(r.next01()*float64(frames-400))
		amp := 0.25 + 0.55*r.next01()
		width := 2 + int(r.next01()*8)

		for j := range width {
			out[pos+j] += amp * math.Exp(-float64(j)/2)
		}
	}

	return normalise(out)
}

// voxSample is a breathy formant voice, tuned to C-4
func voxSample(sec float64, formants [][3]float64, seed uint64) []float64 {
	const harmonics = 15

	n := int(sec * rate)
	out := make([]float64, n)
	phases := make([]float64, harmonics+1)
	r := &rng{
		state: seed,
	}
	weights := make([]float64, harmonics+1)

	for h := 1; h <= harmonics; h++ {
		g := 0.06
		hz := base * float64(h)

		for _, f := range formants {
			g += f[2] * math.Exp(-math.Pow((hz-f[0])/f[1], 2))
		}

		weights[h] = g / float64(h)
		phases[h] = r.next01()
	}

	var phase float64

	for i := range out {
		t := float64(i) / rate
		vib := 1 + 0.006*math.Sin(2*math.Pi*5.4*t) + 0.003*math.Sin(2*math.Pi*1.1*t)
		phase += base * vib / rate

		var s float64

		for h := 1; h <= harmonics; h++ {
			s += weights[h] * math.Sin(2*math.Pi*float64(h)*phase+phases[h])
		}

		env := math.Min(t/0.03, 1) * math.Exp(-t*1.7)
		out[i] = (s + r.next()*0.035) * env
	}

	return normalise(out)
}

func main() {
	s := song.New("marmelade rain", numChannels)

	s.Timing(speed, tempo)

	air := airSample(4096, 0xBEEF)
	pad := brighten(padSample(4096, false, 0xA11CE), air, 0.45)
	choir := brighten(choirSample(8192), air, 0.40)
	lead := leadSample(2048)
	bell := bellSample(1.4)
	vox := voxSample(1.0, [][3]float64{{700, 380, 1.25}, {1220, 520, 0.85}, {2600, 950, 0.45}}, 0xF00D)
	bass := bassSample(4096)
	kick := kickSample(0x1001)
	snare := snareSample(0x2002)
	hat := hatSample(0x3003)
	open := openHatSample(0x4004)
	crash := crashSample(2.2, 0x5005)
	riser := reverseSamples(crashSample(60.0/float64(tempo)*4, 0x6006))
	boom := boomSample()
	vinyl := vinylSample(8192, 0x7007)
	breakBuf := breakSample(kick, snare, hat, open)
	chop := chopSample(breakBuf)
	k := &kit{
		pad: s.Instrument(song.Instrument{
			Name: "detuned pad",
			WAV: toWAV(pad),
			Loop: true,
			Volume: 11,
			Pan: 128,
			Envelope: song.Points(0, 0, 6, 64),
			Fadeout: 2048,
		}),
		choir: s.Instrument(song.Instrument{
			Name: "formant choir",
			WAV: toWAV(choir),
			Loop: true,
			Volume: 8,
			Pan: 128,
			Envelope: song.Points(0, 0, 10, 64),
			Fadeout: 2048,
		}),
		air: s.Instrument(song.Instrument{
			Name: "breath air",
			WAV: toWAV(air),
			Loop: true,
			Volume: 34,
			Pan: 128,
			Envelope: song.Points(0, 0, 16, 56),
			Fadeout: 1024,
		}),
		lead: s.Instrument(song.Instrument{
			Name: "reed lead",
			WAV: toWAV(lead),
			Loop: true,
			Volume: 18,
			Pan: 112,
			Envelope: song.Points(0, 20, 4, 64, 20, 52),
			Fadeout: 768,
		}),
		bell: s.Instrument(song.Instrument{
			Name: "glass bell",
			WAV: toWAV(bell),
			Volume: 30,
			Pan: 160,
		}),
		vox: s.Instrument(song.Instrument{
			Name: "marmelade voice",
			WAV: toWAV(vox),
			Volume: 40,
			Pan: 148,
		}),
		bass: s.Instrument(song.Instrument{
			Name: "reese bass",
			WAV: toWAV(bass),
			Loop: true,
			Volume: 36,
			Pan: 128,
			Envelope: song.Points(0, 0, 2, 64),
			Fadeout: 512,
		}),
		kick: s.Instrument(song.Instrument{
			Name: "deep kick",
			WAV: toWAV(kick),
			Volume: 54,
			Pan: 128,
		}),
		snare: s.Instrument(song.Instrument{
			Name: "break snare",
			WAV: toWAV(snare),
			Volume: 48,
			Pan: 136,
		}),
		hat: s.Instrument(song.Instrument{
			Name: "tight hat",
			WAV: toWAV(hat),
			Volume: 50,
			Pan: 168,
		}),
		open: s.Instrument(song.Instrument{
			Name: "open hat",
			WAV: toWAV(open),
			Volume: 48,
			Pan: 96,
		}),
		chop: s.Instrument(song.Instrument{
			Name: "break chop",
			WAV: toWAV(chop),
			Volume: 46,
			Pan: 128,
		}),
		breakLoop: s.Instrument(song.Instrument{
			Name: "sampled break",
			WAV: toWAV(breakBuf),
			Volume: 36,
			Pan: 128,
		}),
		crash: s.Instrument(song.Instrument{
			Name: "metal crash",
			WAV: toWAV(crash),
			Volume: 46,
			Pan: 128,
		}),
		riser: s.Instrument(song.Instrument{
			Name: "reverse crash",
			WAV: toWAV(riser),
			Volume: 46,
			Pan: 128,
		}),
		boom: s.Instrument(song.Instrument{
			Name: "sub boom",
			WAV: toWAV(boom),
			Volume: 34,
			Pan: 128,
		}),
		vinyl: s.Instrument(song.Instrument{
			Name: "vinyl",
			WAV: toWAV(vinyl),
			Loop: true,
			Volume: 16,
			Pan: 128,
			Fadeout: 4096,
		}),
	}
	introP := intro(s, k)
	dropA1 := pattern{s.Pattern(rows), k}

	dropA1.dropA(false, false)

	dropA2 := pattern{s.Pattern(rows), k}

	dropA2.dropA(true, true)

	dropB1 := pattern{s.Pattern(rows), k}

	dropB1.dropB(false, false)

	dropB2 := pattern{s.Pattern(rows), k}

	dropB2.dropB(true, true)

	dropC1 := pattern{s.Pattern(rows), k}

	dropC1.dropC(false)

	dropC2 := pattern{s.Pattern(rows), k}

	dropC2.dropC(true)

	patterns := []*song.Pattern{
		introP,
		dropA1.p,
		dropA2.p,
		dropB1.p,
		dropA2.p,
		breakdown(s, k),
		build(s, k),
		dropC1.p,
		dropB2.p,
		dropC2.p,
		outro(s, k),
	}

	s.Play(patterns...)

	// write next to this file, wherever the program is run from
	if err := s.WriteBeside(); err != nil {
		log.Fatal(err)
	}
}
