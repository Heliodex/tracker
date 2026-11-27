// Command track16 writes "golden fever dream", a slow ambient breakbeat in F# minor, as an XM module, song.xm,
// in this folder. It is inspired by Danderian's "The Gold Fever": the same 125 BPM pulse, long drifting chords,
// panning arpeggios and a soft hat-on-every-beat break. Every sample is synthesised here, none are lifted from it.
//
// Build it from anywhere inside the repository with go run ./tracks/track16, or with tracker build.
package main

// golden fever dream – Hy4 flash Free

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
	// rng is a deterministic noise source, so the song sounds the same on every build
	rng struct {
		state uint64
	}

	// event is one note in a lead phrase, on a row of its bar
	event struct {
		row int
		note string
		slide bool
		vol int
	}

	// chord is one bar: three pad voices, a bass root, an eight-step arpeggio and a lead phrase
	chord struct {
		pad [3]string
		bass, fifth string
		arp [8]string
		phrase []event
		bell string
		minor bool
	}

	// kit holds the instruments, numbered as they are added to the song
	kit struct {
		pad, warm, air, lead, bell, bass, kick, snare, hat, open, cym, rim, arp, stabMaj, stabMin, riser, boom *song.Instrument
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
	chAir
	chLead
	chBell
	chBass
	chKick
	chSnare
	chHat
	chOpen
	chCym
	chRim
	chArpL
	chArpR
	chStab
	chFX
	numChannels
)

const (
	// rows is an eight-bar section: sixteen rows to the bar
	rows = 128
	// sixteenth notes at 125 BPM, the tempo of the original
	speed = 6
	tempo = 125
	// base is C-4, the note a sample tuned to the tracker rate plays at
	base = 261.6255653005986
)

var (
	// rate is the rate at which C-4 plays
	rate = 8363.0

	fshm = chord{
		pad: [3]string{"F#4", "A-4", "C#5"},
		bass: "F#2", fifth: "C#3",
		arp: [8]string{"F#4", "A-4", "C#5", "F#5", "C#5", "A-4", "F#4", "C#4"},
		phrase: []event{
			{0, "C#5", false, 0},
			{3, "A-4", true, -2},
			{6, "###", false, 0},
			{8, "F#5", false, 2},
			{12, "E-5", true, -3},
			{15, "###", false, 0},
		},
		bell: "C#6", minor: true,
	}
	dMaj = chord{
		pad: [3]string{"D-4", "F#4", "A-4"},
		bass: "D-2", fifth: "A-2",
		arp: [8]string{"D-4", "F#4", "A-4", "D-5", "A-4", "F#4", "D-4", "A-3"},
		phrase: []event{
			{0, "F#5", false, 0},
			{4, "A-4", true, -2},
			{7, "###", false, 0},
			{9, "D-5", false, 0},
			{13, "C#5", true, -2},
			{15, "###", false, 0},
		},
		bell: "A-5",
	}
	aMaj = chord{
		pad: [3]string{"A-3", "C#4", "E-4"},
		bass: "A-2", fifth: "E-3",
		arp: [8]string{"A-4", "C#5", "E-5", "A-5", "E-5", "C#5", "A-4", "E-4"},
		phrase: []event{
			{0, "E-5", false, 0},
			{3, "C#5", true, -2},
			{6, "###", false, 0},
			{8, "A-4", false, 0},
			{12, "B-4", true, -2},
			{15, "###", false, 0},
		},
		bell: "E-6",
	}
	eMaj = chord{
		pad: [3]string{"E-4", "G#4", "B-4"},
		bass: "E-2", fifth: "B-2",
		arp: [8]string{"E-4", "G#4", "B-4", "E-5", "B-4", "G#4", "E-4", "B-3"},
		phrase: []event{
			{0, "B-4", false, 0},
			{4, "G#4", true, -2},
			{7, "###", false, 0},
			{9, "E-5", false, 0},
			{13, "D-5", true, -2},
			{15, "###", false, 0},
		},
		bell: "B-5",
	}
	bm = chord{
		pad: [3]string{"B-3", "D-4", "F#4"},
		bass: "B-2", fifth: "F#3",
		arp: [8]string{"B-4", "D-5", "F#5", "B-5", "F#5", "D-5", "B-4", "F#4"},
		phrase: []event{
			{0, "F#5", false, 0},
			{4, "D-5", true, -2},
			{7, "###", false, 0},
			{9, "B-4", false, 0},
			{13, "A-4", true, -2},
			{15, "###", false, 0},
		},
		bell: "F#6", minor: true,
	}
	csm = chord{
		pad: [3]string{"C#4", "E-4", "G#4"},
		bass: "C#2", fifth: "G#2",
		arp: [8]string{"C#5", "E-5", "G#5", "C#6", "G#5", "E-5", "C#5", "G#4"},
		phrase: []event{
			{0, "G#5", false, 0},
			{3, "E-5", true, -2},
			{6, "###", false, 0},
			{8, "C#5", false, 0},
			{12, "B-4", true, -2},
			{15, "###", false, 0},
		},
		bell: "G#6", minor: true,
	}

	// the i - VI - III - VII loop the track keeps returning to
	verse = []chord{fshm, dMaj, aMaj, eMaj}
	// a darker four-bar turn, for the break and the second drop
	turn = []chord{bm, csm, dMaj, aMaj}
)

// addWave adds a sine at a whole number of cycles over the buffer, so a loop joins without a click
func addWave(out []float64, cycles, amp, phase float64) {
	for i := range out {
		out[i] += amp * math.Sin(2*math.Pi*(cycles*float64(i)/float64(len(out))+phase))
	}
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

	scale := 0.94 / peak

	for i := range samples {
		samples[i] *= scale
	}

	return samples
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

// next01 returns a deterministic value from 0 to 1
func (r *rng) next01() float64 {
	return r.next()*0.5 + 0.5
}

// airSample is a loop of soft high partials, all whole cycles so it joins cleanly
func airSample(frames int, seed uint64) []float64 {
	out := make([]float64, frames)
	r := &rng{state: seed}

	for range 260 {
		hz := 600 + r.next01()*2900
		cycles := math.Max(1, math.Round(hz*float64(frames)/rate))
		amp := (0.02 + 0.07*r.next01()) * (0.3 + math.Exp(-math.Pow((hz-2200)/1400, 2)))

		addWave(out, cycles, amp, r.next01())
	}

	return normalise(out)
}

// padSample is a slow, detuned pad tuned to C-4. Every layer is whole cycles, so it loops without a click.
func padSample(frames int, seed uint64, formant bool) []float64 {
	out := make([]float64, frames)
	r := &rng{state: seed}
	c := math.Round(base * float64(frames) / rate)

	for h := 1; h <= 13; h++ {
		hz := base * float64(h)
		amp := 1 / math.Pow(float64(h), 1.2)

		if formant {
			// a soft vowel: a formant pair in the mids, with a little air above
			for _, f := range [][3]float64{{420, 300, 1.2}, {1150, 480, 0.8}, {2500, 900, 0.3}} {
				amp *= 0.55 + f[2]*math.Exp(-math.Pow((hz-f[0])/f[1], 2))
			}
		} else if hz > 300 && hz < 1400 {
			amp *= 0.45
		}

		addWave(out, c*float64(h), amp, r.next01())

		if h <= 4 {
			// a close, unrelated detune beats gently rather than cancelling the harmonics
			addWave(out, c*float64(h)+1, amp*0.32, r.next01())
		}
	}

	addWave(out, math.Round(c/2), 0.16, 0)

	return normalise(out)
}

// bassSample is a round sub with a little saw edge, tuned to C-4
func bassSample(frames int) []float64 {
	out := make([]float64, frames)
	c := math.Round(base * float64(frames) / rate)

	for h := 1; h <= 8; h++ {
		addWave(out, c*float64(h), 1/math.Pow(float64(h), 2.3), 0)
	}

	addWave(out, math.Round(c/2), 0.65, 0)
	addWave(out, math.Round(c/4), 0.16, 0)

	return normalise(out)
}

// bellSample is a struck glass bell, tuned to C-4
func bellSample(sec float64) []float64 {
	n := int(sec * rate)
	out := make([]float64, n)

	for _, p := range []struct {
		ratio, amp, decay float64
	}{
		{1, 0.48, 2.0},
		{2.02, 0.22, 2.8},
		{2.76, 0.15, 3.4},
		{5.41, 0.08, 4.8},
		{8.93, 0.04, 6.5},
	} {
		for i := range out {
			t := float64(i) / rate
			env := math.Min(t/0.005, 1) * math.Exp(-p.decay*t)
			out[i] += p.amp * env * math.Sin(2*math.Pi*base*p.ratio*t)
		}
	}

	return normalise(fadeOut(out, 400))
}

// leadSample is a breathy voice that opens from "oh" to "ah", tuned to C-4
func leadSample(sec float64, seed uint64) []float64 {
	const harmonics = 13

	formantsA := [][3]float64{{420, 280, 1.1}, {800, 380, 0.7}, {2600, 900, 0.25}}
	formantsB := [][3]float64{{700, 380, 1.1}, {1220, 520, 0.75}, {2600, 950, 0.4}}
	n := int(sec * rate)
	out := make([]float64, n)
	r := &rng{state: seed}
	phases := make([]float64, harmonics+1)
	weightsA := make([]float64, harmonics+1)
	weightsB := make([]float64, harmonics+1)

	for h := 1; h <= harmonics; h++ {
		ga, gb := 0.05, 0.05
		hz := base * float64(h)

		for _, f := range formantsA {
			ga += f[2] * math.Exp(-math.Pow((hz-f[0])/f[1], 2))
		}

		for _, f := range formantsB {
			gb += f[2] * math.Exp(-math.Pow((hz-f[0])/f[1], 2))
		}

		weightsA[h], weightsB[h] = ga/float64(h), gb/float64(h)
		phases[h] = r.next01()
	}

	var phase float64

	for i := range out {
		t := float64(i) / rate
		mix := math.Min(t/(sec*0.6), 1)
		vib := 1 + 0.006*math.Sin(2*math.Pi*4.6*t) + 0.003*math.Sin(2*math.Pi*1.1*t)
		phase += base * vib / rate

		var s float64

		for h := 1; h <= harmonics; h++ {
			w := weightsA[h]*(1-mix) + weightsB[h]*mix
			s += w * math.Sin(2*math.Pi*float64(h)*phase+phases[h])
		}

		env := math.Min(t/0.05, 1) * math.Exp(-t*0.6)
		out[i] = (s + r.next()*0.02) * env
	}

	return normalise(fadeOut(out, 500))
}

// arpSample is a plucked string: a bright attack that settles into a soft tone, tuned to C-4
func arpSample(sec float64, seed uint64) []float64 {
	n := int(sec * rate)
	out := make([]float64, n)
	r := &rng{state: seed}

	for h := 1; h <= 9; h++ {
		amp := 1 / math.Pow(float64(h), 1.5)
		decay := 3.2 + float64(h)*0.55
		phase := r.next01()

		for i := range out {
			t := float64(i) / rate
			env := math.Min(t/0.003, 1) * math.Exp(-decay*t)
			out[i] += amp * env * math.Sin(2*math.Pi*base*float64(h)*t+phase)
		}
	}

	return normalise(fadeOut(out, 200))
}

// stabSample is a short three-note chord with a soft piano-like attack
func stabSample(sec float64, ratios []float64) []float64 {
	n := int(sec * rate)
	out := make([]float64, n)

	for i := range out {
		t := float64(i) / rate
		env := math.Min(t/0.008, 1) * math.Exp(-t*5.2)

		for _, r := range ratios {
			for h := 1; h <= 4; h++ {
				out[i] += env * 0.16 / math.Pow(float64(h), 1.6) *
					math.Sin(2*math.Pi*base*r*float64(h)*t)
			}
		}
	}

	return normalise(fadeOut(out, 250))
}

// kickSample is a deep pitch-swept kick with a soft click
func kickSample(seed uint64) []float64 {
	n := int(0.55 * rate)
	out := make([]float64, n)
	r := &rng{state: seed}

	var phase float64

	for i := range out {
		t := float64(i) / rate
		f := 44 + 165*math.Exp(-t*30)
		phase += f / rate
		body := math.Sin(2*math.Pi*phase) * math.Exp(-t*8)
		click := r.next() * math.Exp(-t*75) * 0.35
		out[i] = body + click
	}

	return normalise(fadeOut(out, 120))
}

// snareSample is a soft snare with a tuned body
func snareSample(seed uint64) []float64 {
	n := int(0.42 * rate)
	r := &rng{state: seed}
	raw := make([]float64, n)

	for i := range raw {
		raw[i] = r.next()
	}

	hp, lp := highpass(raw), lowpass(raw, 0.16)
	out := make([]float64, n)

	for i := range out {
		t := float64(i) / rate
		out[i] = hp[i]*math.Exp(-t*16)*0.62 + lp[i]*math.Exp(-t*10)*0.32 +
			math.Sin(2*math.Pi*180*t)*math.Exp(-t*22)*0.3
	}

	return normalise(fadeOut(out, 150))
}

// hatSample is a short closed hat
func hatSample(seed uint64) []float64 {
	n := int(0.11 * rate)
	r := &rng{state: seed}
	raw := make([]float64, n)

	for i := range raw {
		raw[i] = r.next()
	}

	hp := highpass(highpass(raw))
	out := make([]float64, n)

	for i := range out {
		t := float64(i) / rate
		out[i] = hp[i] * math.Exp(-t*60)
	}

	return normalise(fadeOut(out, 60))
}

// openHatSample is a longer, softer hat
func openHatSample(seed uint64) []float64 {
	n := int(0.42 * rate)
	r := &rng{state: seed}
	raw := make([]float64, n)

	for i := range raw {
		raw[i] = r.next()
	}

	hp := highpass(highpass(raw))
	out := make([]float64, n)

	for i := range out {
		t := float64(i) / rate
		out[i] = hp[i] * math.Exp(-t*13)
	}

	return normalise(fadeOut(out, 220))
}

// cymbalSample is a wash of metallic partials for the top of a section
func cymbalSample(sec float64, seed uint64) []float64 {
	n := int(sec * rate)
	r := &rng{state: seed}
	raw := make([]float64, n)

	for i := range raw {
		raw[i] = r.next()
	}

	hp := highpass(raw)
	out := make([]float64, n)

	for i := range out {
		t := float64(i) / rate
		out[i] = hp[i] * math.Exp(-t*2.4) * 0.8

		for _, hz := range []float64{3100, 3470, 3760, 4070} {
			out[i] += 0.05 * math.Sin(2*math.Pi*hz*t) * math.Exp(-t*3)
		}
	}

	return normalise(fadeOut(out, 350))
}

// rimSample is a dry wooden click, used for the offbeat
func rimSample() []float64 {
	n := int(0.14 * rate)
	out := make([]float64, n)

	for i := range out {
		t := float64(i) / rate
		out[i] = (math.Sin(2*math.Pi*340*t) + 0.35*math.Sin(2*math.Pi*810*t)) * math.Exp(-t*32)
	}

	return normalise(fadeOut(out, 80))
}

// reverseSamples returns the buffer reversed, for a riser
func reverseSamples(in []float64) []float64 {
	out := make([]float64, len(in))

	for i := range in {
		out[i] = in[len(in)-1-i]
	}

	return out
}

// boomSample is a low sine drop for the last transition
func boomSample() []float64 {
	n := int(1.2 * rate)
	out := make([]float64, n)

	var phase float64

	for i := range out {
		t := float64(i) / rate
		f := 32 + 72*math.Exp(-t*2.2)
		phase += f / rate
		out[i] = math.Sin(2*math.Pi*phase) * math.Exp(-t*3)
	}

	return normalise(fadeOut(out, 300))
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

// brighten mixes a looping air sample into a pad, adding top end without adding mud
func brighten(base, air []float64, amount float64) []float64 {
	out := append([]float64(nil), base...)

	for i := range out {
		out[i] += air[i%len(air)] * amount
	}

	return normalise(out)
}

// upOctave returns note raised an octave, for the answering arpeggio
func upOctave(note string) string {
	n, err := types.ParseNote(note)
	if err != nil {
		panic(err)
	}

	return types.Note(int(n) + 12).String()
}

// reverseUp reverses an arpeggio and lifts it an octave
func reverseUp(notes [8]string) (out [8]string) {
	for i := range notes {
		out[i] = upOctave(notes[len(notes)-1-i])
	}

	return
}

// pads holds a chord on the three pad voices, spread across the stereo field
func (w pattern) pads(b int, c chord, vol int) {
	w.padVoices(b, c, vol, w.k.pad)
}

// warmPads uses the vowel pad in the break, where the drums leave room for it
func (w pattern) warmPads(b int, c chord, vol int) {
	w.padVoices(b, c, vol, w.k.warm)
}

// padVoices plays a chord on the three pad channels with the chosen instrument.
// Each voice holds for the whole bar while its pan drifts slowly across the stereo field.
func (w pattern) padVoices(b int, c chord, vol int, ins *song.Instrument) {
	pans := [3]int{0x40, 0x80, 0xC0}

	for i, note := range c.pad {
		ch := chPadLo + i
		w.p.Hit(at(b, 0), ch, note, ins, song.VolOpts(vol, song.Pan(pans[i]))...)

		for _, r := range []int{3, 6, 9, 12} {
			t := float64(r) / song.RowsPerBar
			pan := 0x34 + int(t*0x98)
			if b%2 == 1 {
				pan = 0xCC - int(t*0x98)
			}

			// keep the three voices apart as they sweep, so the chord keeps its width
			pan = max(0, min(255, pan+(i-1)*0x14))
			w.p.Hit(at(b, r), ch, "...", nil, song.Pan(pan))
		}

		w.p.Hit(at(b, song.RowsPerBar-1), ch, "###", nil)
	}
}

// air adds a breathy high texture, alternating sides each bar
func (w pattern) air(b int, vol int) {
	pan := 0x50
	if b%2 == 1 {
		pan = 0xB0
	}

	w.p.Hit(at(b, 0), chAir, "C-5", w.k.air, song.VolOpts(vol, song.Pan(pan))...)
}

// bass plays the root on beat one and the fifth on beat three; style 2 adds an octave pulse
func (w pattern) bass(b int, c chord, style, vol int) {
	w.p.Hit(at(b, 0), chBass, c.bass, w.k.bass, song.VolOpts(vol)...)

	if style == 0 {
		return
	}

	w.p.Hit(at(b, 8), chBass, c.fifth, w.k.bass, song.VolOpts(vol-0x06, effect(0x03, 0x06))...)

	if style == 2 {
		w.p.Hit(at(b, 6), chBass, c.bass, w.k.bass, song.VolOpts(vol-0x0A, effect(0x03, 0x08))...)
		w.p.Hit(at(b, 14), chBass, c.fifth, w.k.bass, song.VolOpts(vol-0x0C, effect(0x03, 0x08))...)
	}
}

// lead plays a phrase of long voice notes; slide notes glide into place
func (w pattern) lead(b int, c chord, vol int) {
	for _, e := range c.phrase {
		if e.note == "###" {
			w.p.Hit(at(b, e.row), chLead, "###", nil)
			continue
		}

		opts := []song.Opt{song.Vol(vol + e.vol), song.Vibrato(4, 2)}
		if e.slide {
			opts = append(opts, effect(0x03, 0x08))
		}

		w.p.Hit(at(b, e.row), chLead, e.note, w.k.lead, opts...)
	}
}

// bellAccent drops one high bell note, panned away from the voice
func (w pattern) bellAccent(b int, c chord, vol, row int) {
	pan := 0x60
	if b%2 == 0 {
		pan = 0xA0
	}

	w.p.Hit(at(b, row), chBell, c.bell, w.k.bell, song.VolOpts(vol, song.Pan(pan))...)
}

// arps plays the arpeggio on two panned channels; the right answers an octave up and backwards
func (w pattern) arps(b int, c chord, vol int, dense bool) {
	step := 2
	if dense {
		step = 1
	}

	answer := reverseUp(c.arp)

	for i, note := range c.arp {
		row := i * step
		if row >= song.RowsPerBar {
			break
		}

		pan := 0x4C
		if i%2 == 1 {
			pan = 0x64
		}

		v := vol
		if i%4 == 0 {
			v += 0x06
		}

		w.p.Hit(at(b, row), chArpL, note, w.k.arp, song.VolOpts(v, song.Pan(pan))...)

		if dense || i%2 == 0 {
			w.p.Hit(at(b, row), chArpR, answer[i], w.k.arp,
				song.VolOpts(v-0x0A, song.Pan(0xB0+(i%2)*0x10))...)
		}
	}
}

// stab drops a short chord on an offbeat, major or minor to match the bar
func (w pattern) stab(b, row int, c chord, vol int) {
	ins := w.k.stabMaj
	if c.minor {
		ins = w.k.stabMin
	}

	w.p.Hit(at(b, row), chStab, "C-5", ins, song.VolOpts(vol)...)
}

// hats plays a closed hat on every beat, with ghost eighths when the beat is busy
func (w pattern) hats(b, style int) {
	for i, r := range song.Steps(0, song.RowsPerBar, 4) {
		vol, pan := 0x28, 0x70
		if r%8 == 0 {
			vol = 0x36
		}
		if i%2 == 1 {
			pan = 0xA0
		}

		w.p.Hit(at(b, r), chHat, "C-5", w.k.hat, song.VolOpts(vol, song.Pan(pan))...)
	}

	if style < 2 {
		return
	}

	for i, r := range song.Steps(2, song.RowsPerBar, 4) {
		pan := 0x58
		if i%2 == 1 {
			pan = 0x98
		}

		w.p.Hit(at(b, r), chHat, "C-5", w.k.hat, song.VolOpts(0x18, song.Pan(pan))...)
	}
}

// drums plays the soft break. Style 1 is sparse, 2 is the full pulse and 3 is busy.
func (w pattern) drums(b, style int) {
	if style == 0 {
		return
	}

	w.hats(b, style)

	kicks := []int{0, 8}
	if style >= 3 {
		kicks = []int{0, 6, 8, 14}
	}

	for _, r := range kicks {
		vol := 0x3C
		if r != 0 && r != 8 {
			vol = 0x26
		}

		w.p.Hit(at(b, r), chKick, "C-4", w.k.kick, song.VolOpts(vol)...)
	}

	w.p.Hit(at(b, 4), chSnare, "C-4", w.k.snare, song.VolOpts(0x38, song.Pan(0x88))...)
	w.p.Hit(at(b, 12), chSnare, "C-4", w.k.snare, song.VolOpts(0x40, song.Pan(0x78))...)

	if style >= 2 {
		w.p.Hit(at(b, 14), chSnare, "C-4", w.k.snare, song.VolOpts(0x1C, song.Pan(0x90))...)
		w.p.Hit(at(b, 6), chOpen, "C-5", w.k.open, song.VolOpts(0x2A, song.Pan(0x60))...)
		w.p.Hit(at(b, 15), chOpen, "C-5", w.k.open, song.VolOpts(0x24, song.Pan(0xA0))...)
		w.p.Hit(at(b, 7), chRim, "C-5", w.k.rim, song.VolOpts(0x20, song.Pan(0x9C))...)
	}

	if style >= 3 {
		w.p.Hit(at(b, 3), chRim, "C-5", w.k.rim, song.VolOpts(0x18, song.Pan(0x64))...)
		w.p.Hit(at(b, 11), chRim, "C-5", w.k.rim, song.VolOpts(0x1A, song.Pan(0x94))...)
	}

	if style >= 2 && b%4 == 0 {
		w.p.Hit(at(b, 0), chCym, "C-5", w.k.cym, song.VolOpts(0x2C, song.Pan(0x80))...)
	}
}

// fill rolls the rim into the next section over the last beat of bar b
func (w pattern) fill(b int) {
	for i, r := range []int{12, 13, 14, 15} {
		w.p.Hit(at(b, r), chRim, "C-5", w.k.rim,
			song.VolOpts(0x1E+i*0x08, song.Pan(0x70+i*0x0A))...)
	}
}

// roll is a snare crescendo from row start up to, but not including, row end
func (w pattern) roll(b, start, end, fromVol, toVol int) {
	for r := start; r < end; r++ {
		t := float64(r-start) / float64(max(1, end-start-1))
		vol := fromVol + int(t*float64(toVol-fromVol))

		w.p.Hit(at(b, r), chSnare, "C-4", w.k.snare,
			song.VolOpts(vol, song.Pan(0x70+int(t*0x20)))...)
	}
}

// stutter speeds the rows up over the last beat, then restores the song speed
func (w pattern) stutter(b int) {
	w.p.Hit(at(b, 8), chFX, "...", nil, effect(0x0F, 4))
	w.p.Hit(at(b, 12), chFX, "...", nil, effect(0x0F, 3))
	w.p.Hit(at(b, 14), chFX, "...", nil, effect(0x0F, 6))
}

// intro opens with pads and air, then lets the hats and arpeggios surface
func intro(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b := range rows / song.RowsPerBar {
		c := barChord(verse, b)

		w.pads(b, c, 0x1C)
		w.air(b, 0x22)

		if b%2 == 0 {
			w.bellAccent(b, c, 0x20, 14)
		}

		if b >= 2 {
			w.arps(b, c, 0x24, false)
		}

		if b >= 4 {
			w.hats(b, 1)
			w.bass(b, c, 0, 0x28)
		}

		if b == 6 {
			w.lead(b, c, 0x22)
		}
	}

	w.fx(7, 0, "C-5", k.riser, 0x28)
	w.roll(7, 8, 16, 0x14, 0x2C)

	return w.p
}

// fx plays a one-shot on the effects channel
func (w pattern) fx(b, row int, note string, ins *song.Instrument, vol int) {
	w.p.Hit(at(b, row), chFX, note, ins, song.VolOpts(vol)...)
}

// verseA lays down the pulse and the first lead
func verseA(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b := range rows / song.RowsPerBar {
		c := barChord(verse, b)

		w.pads(b, c, 0x24)
		w.air(b, 0x26)
		w.bass(b, c, 1, 0x34)
		w.drums(b, 1+(b/4))

		if b >= 2 {
			w.arps(b, c, 0x2E, false)
		}

		if b >= 4 {
			w.lead(b, c, 0x28)
			w.stab(b, 10, c, 0x20)
		}

		if b%2 == 1 {
			w.bellAccent(b, c, 0x24, 6)
		}
	}

	w.fill(7)

	return w.p
}

// verseB turns to the darker chords and opens the beat up
func verseB(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b := range rows / song.RowsPerBar {
		c := barChord(turn, b)

		w.pads(b, c, 0x26)
		w.air(b, 0x28)
		w.bass(b, c, 1+b%2, 0x36)
		w.drums(b, 2)
		w.arps(b, c, 0x30, b >= 4)
		w.lead(b, c, 0x2A)

		if b%2 == 0 {
			w.stab(b, 0, c, 0x24)
			w.bellAccent(b, c, 0x26, 14)
		}
	}

	w.fill(7)

	return w.p
}

// break drops the drums and lets the pads and voice carry the turn
func breakdown(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b := range rows / song.RowsPerBar {
		c := barChord(turn, b)

		w.warmPads(b, c, 0x1E)
		w.air(b, 0x30)

		if b < 2 {
			w.bass(b, c, 0, 0x26)
		}

		w.lead(b, c, 0x2C)

		if b >= 2 {
			w.arps(b, c, 0x26, false)
		}

		if b%2 == 0 {
			w.bellAccent(b, c, 0x28, 14)
		}
	}

	w.fx(7, 0, "C-5", k.riser, 0x30)
	w.roll(7, 12, 16, 0x14, 0x2C)

	return w.p
}

// build brings the drums back one layer at a time
func build(s *song.Song, k *kit) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b := range rows / song.RowsPerBar {
		c := barChord(verse, b)

		w.pads(b, c, 0x24)
		w.air(b, 0x2A)
		w.bass(b, c, 1+b%2, 0x32)

		style := 1
		if b >= 2 {
			style = 2
		}
		if b >= 6 {
			style = 3
		}

		w.drums(b, style)

		if b >= 4 {
			w.arps(b, c, 0x2C, true)
		}

		if b >= 6 {
			w.lead(b, c, 0x2A)
			w.roll(b, 0, 16, 0x16, 0x30)
		}
	}

	w.fx(7, 0, "C-5", k.riser, 0x38)
	w.stutter(7)

	return w.p
}

// drop is the full arrangement; style 2 is the first drop and 3 the last
func drop(s *song.Song, k *kit, cs []chord, style int, final bool) *song.Pattern {
	w := pattern{s.Pattern(rows), k}

	for b := range rows / song.RowsPerBar {
		c := barChord(cs, b)

		w.pads(b, c, 0x28)
		w.air(b, 0x2C)
		w.bass(b, c, 2, 0x3A)
		w.drums(b, style)
		w.arps(b, c, 0x34, true)
		w.lead(b, c, 0x30)

		if b%2 == 0 {
			w.stab(b, 0, c, 0x28)
			w.bellAccent(b, c, 0x2A, 14)
		} else {
			w.stab(b, 8, c, 0x22)
		}

		if b%4 == 0 {
			w.fx(b, 0, "C-5", k.cym, 0x30)
		}
	}

	w.fill(7)

	if final {
		w.fx(7, 8, "C-4", k.boom, 0x3C)
	}

	return w.p
}

// outro lets the last chord ring out and eases the sustained voices away
func outro(s *song.Song, k *kit) *song.Pattern {
	const last int = 3

	w := pattern{s.Pattern(rows / 2), k}

	for b := range 3 {
		c := barChord(verse, b)

		w.pads(b, c, 0x1E)
		w.air(b, 0x22)

		if b == 0 {
			w.fx(b, 0, "C-5", k.cym, 0x2C)
			w.bass(b, c, 0, 0x26)
		}

		if b == 2 {
			w.lead(b, c, 0x24)
			w.bellAccent(b, c, 0x26, 14)
		}
	}

	w.pads(last, fshm, 0x18)
	w.air(last, 0x1C)
	w.bass(last, fshm, 0, 0x20)
	w.bellAccent(last, fshm, 0x24, 14)

	// ease the sustained voices away while the last chord releases
	for r := 8; r < 15; r++ {
		for _, ch := range []int{chPadLo, chPadMid, chPadHi, chBass, chAir} {
			w.p.Hit(at(last, r), ch, "...", nil, effect(0x0A, 0x02))
		}
	}

	for _, ch := range []int{
		chAir, chLead, chBell, chBass, chKick, chSnare, chHat, chOpen,
		chCym, chRim, chArpL, chArpR, chStab, chFX,
	} {
		w.p.Hit(rows/2-1, ch, "###", nil)
	}

	return w.p
}

func main() {
	s := song.New("golden fever dream", numChannels)
	s.Timing(speed, tempo)

	air := airSample(4096, 0xBEEF)
	pad := brighten(padSample(4096, 0xA11CE, false), air, 0.30)
	voicePad := brighten(padSample(8192, 0xC401, true), air, 0.28)
	bass := bassSample(2048)
	kick := kickSample(0x1001)
	snare := snareSample(0x2002)
	hat := hatSample(0x3003)
	open := openHatSample(0x4004)
	cym := cymbalSample(1.8, 0x5005)
	rim := rimSample()
	arp := arpSample(0.8, 0x6006)
	stabMaj := stabSample(0.7, []float64{1, 1.1892, 1.4983})
	stabMin := stabSample(0.7, []float64{1, 1.1225, 1.4983})
	lead := leadSample(1.5, 0xCAFE)
	bell := bellSample(1.4)
	riser := reverseSamples(cymbalSample(60.0/float64(tempo)*8, 0x7007))
	boom := boomSample()

	k := &kit{
		pad: s.Instrument(song.Instrument{
			Name: "drifting pad", WAV: toWAV(pad), Loop: true,
			Volume: 12, Pan: 128, Envelope: song.Points(0, 0, 6, 64), Fadeout: 2048,
		}),
		warm: s.Instrument(song.Instrument{
			Name: "warm vowel pad", WAV: toWAV(voicePad), Loop: true,
			Volume: 10, Pan: 128, Envelope: song.Points(0, 0, 8, 64), Fadeout: 2048,
		}),
		air: s.Instrument(song.Instrument{
			Name: "night air", WAV: toWAV(air), Loop: true,
			Volume: 24, Pan: 128, Envelope: song.Points(0, 0, 16, 56), Fadeout: 1024,
		}),
		lead: s.Instrument(song.Instrument{
			Name: "golden voice", WAV: toWAV(lead),
			Volume: 18, Pan: 112, Fadeout: 512,
		}),
		bell: s.Instrument(song.Instrument{
			Name: "glass bell", WAV: toWAV(bell),
			Volume: 26, Pan: 156,
		}),
		bass: s.Instrument(song.Instrument{
			Name: "round bass", WAV: toWAV(bass), Loop: true,
			Volume: 34, Pan: 128, Envelope: song.Points(0, 0, 2, 64), Fadeout: 512,
		}),
		kick: s.Instrument(song.Instrument{
			Name: "deep kick", WAV: toWAV(kick),
			Volume: 52, Pan: 128,
		}),
		snare: s.Instrument(song.Instrument{
			Name: "soft snare", WAV: toWAV(snare),
			Volume: 42, Pan: 136,
		}),
		hat: s.Instrument(song.Instrument{
			Name: "closed hat", WAV: toWAV(hat),
			Volume: 44, Pan: 168,
		}),
		open: s.Instrument(song.Instrument{
			Name: "open hat", WAV: toWAV(open),
			Volume: 40, Pan: 96,
		}),
		cym: s.Instrument(song.Instrument{
			Name: "soft cymbal", WAV: toWAV(cym),
			Volume: 38, Pan: 128,
		}),
		rim: s.Instrument(song.Instrument{
			Name: "rim click", WAV: toWAV(rim),
			Volume: 36, Pan: 152,
		}),
		arp: s.Instrument(song.Instrument{
			Name: "plucked arp", WAV: toWAV(arp),
			Volume: 38, Pan: 128,
		}),
		stabMaj: s.Instrument(song.Instrument{
			Name: "major stab", WAV: toWAV(stabMaj),
			Volume: 36, Pan: 128,
		}),
		stabMin: s.Instrument(song.Instrument{
			Name: "minor stab", WAV: toWAV(stabMin),
			Volume: 36, Pan: 128,
		}),
		riser: s.Instrument(song.Instrument{
			Name: "reverse wash", WAV: toWAV(riser),
			Volume: 40, Pan: 128,
		}),
		boom: s.Instrument(song.Instrument{
			Name: "sub boom", WAV: toWAV(boom),
			Volume: 34, Pan: 128,
		}),
	}

	patterns := []*song.Pattern{
		intro(s, k),
		verseA(s, k),
		verseB(s, k),
		breakdown(s, k),
		build(s, k),
		drop(s, k, verse, 2, false),
		drop(s, k, turn, 2, false),
		breakdown(s, k),
		build(s, k),
		drop(s, k, verse, 3, false),
		drop(s, k, verse, 3, true),
		outro(s, k),
	}

	s.Play(patterns...)

	// write next to this file, wherever the program is run from
	if err := s.WriteBeside(); err != nil {
		log.Fatal(err)
	}
}
