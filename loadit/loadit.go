// Package loadit reads Impulse Tracker (.it) modules into the text format, so they can be saved as XM.
//
// Only instrument mode modules are read, and compressed samples are refused. Anything XM cannot
// represent is dropped or approximated, and Read reports each case as a warning.
package loadit

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/Heliodex/tracker/loadtext"
	"github.com/Heliodex/tracker/types"
)

const (
	headerLen = 0xC0
	instLen = 0x226
	sampleLen = 0x50
	envLen = 82
	envNodes = 25
	maxXMEnv = 12
	maxXMChans = 32
	maxXMSmps = 16
	defaultC5 = 8363
	emptyRows = 64
	flagInstMode = 0x04
	flagLinear = 0x08
	smpExists = 0x01
	smp16Bit = 0x02
	smpStereo = 0x04
	smpCompress = 0x08
	smpLoop = 0x10
	smpPingPong = 0x40
)

type (
	// cell is one unpacked IT pattern cell. note and vol are -1 when empty, ins and cmd are 0 when empty
	cell struct {
		note, ins, vol, cmd, param int
	}

	// envelope is an XM envelope, with its loop and sustain points
	envelope struct {
		points []types.EnvPoint
		flags, sustain, start, end uint8
	}

	// report counts the things XM cannot keep, so each is reported once
	report struct {
		counts map[string]int
	}

	conv struct {
		d []byte
		rep *report
		smpOff []int
	}
)

var emptyCell = cell{note: -1, vol: -1}

func (r *report) add(msg string) {
	r.counts[msg]++
}

func (r *report) list() []string {
	out := make([]string, 0, len(r.counts))

	for msg, n := range r.counts {
		if n > 1 {
			out = append(out, fmt.Sprintf("%s (%d times)", msg, n))
		} else {
			out = append(out, msg)
		}
	}

	sort.Strings(out)

	return out
}

var errTruncated = errors.New("truncated Impulse Tracker file")

// get returns n bytes at off, or an error if they are past the end of d
func get(d []byte, off, n int) ([]byte, error) {
	if off < 0 || n < 0 || off > len(d) || n > len(d)-off {
		return nil, errTruncated
	}

	return d[off : off+n], nil
}

func u16(b []byte, off int) int {
	return int(binary.LittleEndian.Uint16(b[off:]))
}

func u32(b []byte, off int) int {
	return int(binary.LittleEndian.Uint32(b[off:]))
}

func clamp(v, lo, hi int) int {
	return max(lo, min(hi, v))
}

// Read converts an Impulse Tracker module to the text format. The warnings list what was dropped or approximated.
// Call ToFile on the result, then save.Write, to get an XM file.
func Read(d []byte) (*loadtext.TextFormat, []string, error) {
	hdr, err := get(d, 0, headerLen)
	if err != nil {
		return nil, nil, err
	}

	if string(hdr[:4]) != "IMPM" {
		return nil, nil, errors.New("not an Impulse Tracker module")
	}

	ordNum, insNum, smpNum, patNum := u16(hdr, 0x20), u16(hdr, 0x22), u16(hdr, 0x24), u16(hdr, 0x26)
	flags := u16(hdr, 0x2C)

	if flags&flagInstMode == 0 {
		return nil, nil, errors.New("sample mode modules are not supported, only instrument mode")
	}

	if insNum > 128 || patNum > 256 {
		return nil, nil, errors.New("too many instruments or patterns for XM")
	}

	rep := &report{counts: map[string]int{}}
	c := &conv{d: d, rep: rep}

	tf := &loadtext.TextFormat{
		LinearSlides: flags&flagLinear != 0,
		Speed: uint16(hdr[0x32]),
		Tempo: uint16(hdr[0x33]),
		NumInstruments: uint16(insNum),
		NumPatterns: uint16(patNum),
	}

	copy(tf.Name[:], hdr[4:24])

	if hdr[0x30] != 128 {
		rep.add("initial global volume is not kept")
	}

	orders, err := get(d, headerLen, ordNum)
	if err != nil {
		return nil, nil, err
	}

	for _, o := range orders {
		if o == 255 {
			break
		}

		// 254 is a marker for skipped positions, not a pattern
		if o == 254 {
			continue
		}

		if int(o) >= patNum {
			return nil, nil, fmt.Errorf("order refers to pattern %d, but there are %d", o, patNum)
		}

		tf.OrderTable = append(tf.OrderTable, o)
	}

	tables, err := get(d, headerLen+ordNum, 4*(insNum+smpNum+patNum))
	if err != nil {
		return nil, nil, err
	}

	insOff := make([]int, insNum)
	for i := range insOff {
		insOff[i] = u32(tables, 4*i)
	}

	c.smpOff = make([]int, smpNum)
	for i := range c.smpOff {
		c.smpOff[i] = u32(tables, 4*(insNum+i))
	}

	patOff := make([]int, patNum)
	for i := range patOff {
		patOff[i] = u32(tables, 4*(insNum+smpNum+i))
	}

	// unpack every pattern first, to find how many channels are used
	cellsByPat := make([][][64]cell, patNum)
	rowsByPat := make([]int, patNum)
	maxChan := -1

	for p, off := range patOff {
		if off == 0 {
			rowsByPat[p] = emptyRows
			continue
		}

		rows, cells, err := unpackPattern(d, off)
		if err != nil {
			return nil, nil, fmt.Errorf("pattern %d: %w", p, err)
		}

		rowsByPat[p] = rows
		cellsByPat[p] = cells

		for _, row := range cells {
			for ch, x := range row {
				if x.note != -1 || x.ins != 0 || x.vol != -1 || x.cmd != 0 {
					maxChan = max(maxChan, ch)
				}
			}
		}
	}

	if maxChan >= maxXMChans {
		rep.add("channels past 32 are folded into free channels")

		for p := range cellsByPat {
			if cellsByPat[p] != nil {
				fold(cellsByPat[p], rep)
			}
		}

		maxChan = maxXMChans - 1
	}

	channels := max(2, maxChan+1)
	tf.NumChannels = uint16(channels)

	for p := range patNum {
		pat := make(loadtext.Pattern, rowsByPat[p])

		for r := range pat {
			pat[r] = make([]loadtext.ChannelData, channels)

			if cellsByPat[p] == nil {
				continue
			}

			for ch := range channels {
				pat[r][ch] = c.channelData(cellsByPat[p][r][ch])
			}
		}

		tf.Patterns = append(tf.Patterns, pat)
	}

	c.channelPans(tf, hdr)

	for i, off := range insOff {
		ins, err := c.instrument(off)
		if err != nil {
			return nil, nil, fmt.Errorf("instrument %d: %w", i+1, err)
		}

		tf.Instruments = append(tf.Instruments, ins)
	}

	return tf, rep.list(), nil
}

// unpackPattern decodes the packed pattern at off into rows of cells
func unpackPattern(d []byte, off int) (int, [][64]cell, error) {
	h, err := get(d, off, 8)
	if err != nil {
		return 0, nil, err
	}

	rows := u16(h, 2)

	data, err := get(d, off+8, u16(h, 0))
	if err != nil {
		return 0, nil, err
	}

	cells := make([][64]cell, rows)

	for r := range cells {
		for ch := range cells[r] {
			cells[r][ch] = emptyCell
		}
	}

	var (
		mask [64]uint8
		last [64]cell
		pos int
	)

	for ch := range last {
		last[ch] = emptyCell
	}

	read := func() (int, error) {
		if pos >= len(data) {
			return 0, errors.New("truncated packed pattern")
		}

		pos++

		return int(data[pos-1]), nil
	}

	for row := 0; row < rows; {
		b, err := read()
		if err != nil {
			return 0, nil, err
		}

		if b == 0 {
			row++
			continue
		}

		ch := (b - 1) & 63

		if b&0x80 != 0 {
			m, err := read()
			if err != nil {
				return 0, nil, err
			}

			mask[ch] = uint8(m)
		}

		m := mask[ch]
		x := emptyCell

		if m&0x01 != 0 {
			v, err := read()
			if err != nil {
				return 0, nil, err
			}

			last[ch].note = v
		}

		if m&0x01 != 0 || m&0x10 != 0 {
			x.note = last[ch].note
		}

		if m&0x02 != 0 {
			v, err := read()
			if err != nil {
				return 0, nil, err
			}

			last[ch].ins = v
		}

		if m&0x02 != 0 || m&0x20 != 0 {
			x.ins = last[ch].ins
		}

		if m&0x04 != 0 {
			v, err := read()
			if err != nil {
				return 0, nil, err
			}

			last[ch].vol = v
		}

		if m&0x04 != 0 || m&0x40 != 0 {
			x.vol = last[ch].vol
		}

		if m&0x08 != 0 {
			cmd, err := read()
			if err != nil {
				return 0, nil, err
			}

			param, err := read()
			if err != nil {
				return 0, nil, err
			}

			last[ch].cmd, last[ch].param = cmd, param
		}

		if m&0x08 != 0 || m&0x80 != 0 {
			x.cmd, x.param = last[ch].cmd, last[ch].param
		}

		cells[row][ch] = x
	}

	return rows, cells, nil
}

// isEmpty reports whether a cell holds nothing
func isEmpty(x cell) bool {
	return x.note == -1 && x.ins == 0 && x.vol == -1 && x.cmd == 0
}

// fold moves the cells of channels past XM's limit into a free channel of the same row.
// A cell is dropped where every channel in the row is in use
func fold(cells [][64]cell, rep *report) {
	for r := range cells {
		for ch := maxXMChans; ch < len(cells[r]); ch++ {
			x := cells[r][ch]

			if isEmpty(x) {
				continue
			}

			cells[r][ch] = emptyCell

			slot := -1

			for s := range maxXMChans {
				if isEmpty(cells[r][s]) {
					slot = s
					break
				}
			}

			if slot < 0 {
				rep.add("cells past channel 32 are dropped where no channel is free")
				continue
			}

			cells[r][slot] = x
		}
	}
}

// channelData converts an IT cell to an XM cell
func (c *conv) channelData(x cell) loadtext.ChannelData {
	var cd loadtext.ChannelData

	cut := x.note == 254

	switch {
	case x.note == -1:
	case x.note <= 95:
		cd.Note = uint8(x.note + 1)
	case x.note == 255:
		cd.Note = 97
	case cut:
	default:
		c.rep.add("notes above B-7 are dropped")
	}

	if x.ins > 0 {
		cd.Instrument = uint8(x.ins)
	}

	if x.vol != -1 {
		if v, ok := volumeColumn(x.vol); ok {
			cd.Volume = v
		} else {
			c.rep.add("volume column pitch slides are dropped")
		}
	}

	if x.cmd != 0 {
		if eff, param, ok := effect(x.cmd, x.param); ok {
			cd.Effect, cd.EffectParameter = eff, param
		} else {
			c.rep.add(fmt.Sprintf("effect %c is dropped", 'A'+x.cmd-1))
		}
	}

	// a note cut is ECx with x at 0, which cuts the note at once, as IT does
	if cut {
		if cd.Effect == 0 {
			cd.Effect, cd.EffectParameter = 0x0E, 0xC0
		} else {
			c.rep.add("note cut becomes note off where the row has an effect")
			cd.Note = 97
		}
	}

	return cd
}

// volumeColumn maps an IT volume column value to an XM volume column byte
func volumeColumn(v int) (uint8, bool) {
	switch {
	case v <= 64:
		return uint8(0x10 + v), true
	case v <= 74:
		// fine volume up, approximated as a volume slide up
		return uint8(0x70 + v - 65), true
	case v <= 84:
		return uint8(0x60 + v - 75), true
	case v <= 94:
		return uint8(0x70 + v - 85), true
	case v <= 104:
		return uint8(0x60 + v - 95), true
	case v >= 128 && v <= 192:
		return uint8(0xC0 + min(15, (v-128)*15/64)), true
	case v >= 193 && v <= 202:
		return uint8(0xF0 + v - 193), true
	case v >= 203 && v <= 212:
		return uint8(0xB0 + v - 203), true
	}

	return 0, false
}

// effect maps an IT command, such as 1 for A, to an XM effect. ok is false when XM has no equivalent
func effect(cmd, p int) (uint8, uint8, bool) {
	x, y := p>>4, p&0xF

	switch 'A' + cmd - 1 {
	case 'A':
		if p > 0 && p < 0x20 {
			return 0x0F, uint8(p), true
		}
	case 'B':
		return 0x0B, uint8(p), true
	case 'C':
		return 0x0D, uint8(p), true
	case 'D':
		switch {
		case y == 0xF && x == 0xF:
		case y == 0xF:
			return 0x0E, 0xA0 | uint8(x), true
		case x == 0xF:
			return 0x0E, 0xB0 | uint8(y), true
		default:
			return 0x0A, uint8(p), true
		}
	case 'E':
		if p >= 0xF0 {
			return 0x0E, 0x20 | uint8(y), true
		}

		if p < 0xE0 {
			return 0x02, uint8(p), true
		}
	case 'F':
		if p >= 0xF0 {
			return 0x0E, 0x10 | uint8(y), true
		}

		if p < 0xE0 {
			return 0x01, uint8(p), true
		}
	case 'G':
		return 0x03, uint8(p), true
	case 'H', 'U':
		return 0x04, uint8(p), true
	case 'J':
		return 0x00, uint8(p), true
	case 'K':
		return 0x06, uint8(p), true
	case 'L':
		return 0x05, uint8(p), true
	case 'O':
		return 0x09, uint8(p), true
	case 'Q':
		return 0x0E, 0x90 | uint8(y), true
	case 'R':
		return 0x07, uint8(p), true
	case 'S':
		switch x {
		case 0x3:
			return 0x0E, 0x40 | uint8(y), true
		case 0x4:
			return 0x0E, 0x70 | uint8(y), true
		case 0x8:
			// IT pan is 0 to 15, XM pan is 0 to 255
			return 0x08, uint8(y * 17), true
		case 0xB:
			return 0x0E, 0x60 | uint8(y), true
		case 0xC:
			return 0x0E, 0xC0 | uint8(y), true
		case 0xD:
			return 0x0E, 0xD0 | uint8(y), true
		case 0xE:
			return 0x0E, 0xE0 | uint8(y), true
		}
	case 'T':
		// tempo below 0x20 is a tempo slide, which XM cannot do
		if p >= 0x20 {
			return 0x0F, uint8(p), true
		}
	case 'V':
		// IT global volume is 0 to 128, XM is 0 to 64
		return 0x10, uint8(min(64, p/2)), true
	case 'W':
		return 0x11, uint8(p), true
	case 'X':
		return 0x08, uint8(p), true
	}

	return 0, 0, false
}

// channelPans sets the starting pan of each channel with an 8xx effect on the first row of the first order
func (c *conv) channelPans(tf *loadtext.TextFormat, hdr []byte) {
	if len(tf.OrderTable) == 0 || len(tf.Patterns[tf.OrderTable[0]]) == 0 {
		return
	}

	first := tf.Patterns[tf.OrderTable[0]][0]

	for ch := range int(tf.NumChannels) {
		pv := int(hdr[0x40+ch])

		if pv&0x80 != 0 || pv == 32 {
			continue
		}

		cd := &first[ch]

		if cd.Effect != 0 {
			c.rep.add("channel pan is dropped where the first row has an effect")
			continue
		}

		cd.Effect, cd.EffectParameter = 0x08, uint8(min(255, pv*4))
	}
}

// envelope reads an IT envelope at off. It returns a disabled envelope when the envelope is off
func (c *conv) envelope(off int, what string) (envelope, error) {
	b, err := get(c.d, off, envLen)
	if err != nil {
		return envelope{}, err
	}

	if b[0]&1 == 0 {
		return envelope{}, nil
	}

	env := envelope{flags: 1}

	for i := range min(int(b[1]), envNodes) {
		if len(env.points) == maxXMEnv {
			c.rep.add(what + " envelope has more than 12 points, the rest are dropped")
			break
		}

		env.points = append(env.points, types.EnvPoint{
			X: binary.LittleEndian.Uint16(b[7+3*i:]),
			Y: uint16(clamp(int(int8(b[6+3*i])), 0, 64)),
		})
	}

	n := len(env.points)

	if b[0]&0x02 != 0 {
		start, end := int(b[2]), int(b[3])

		if start < n && end < n && start <= end {
			env.flags |= 4
			env.start, env.end = uint8(start), uint8(end)
		} else {
			c.rep.add(what + " envelope loop is dropped")
		}
	}

	if b[0]&0x04 != 0 {
		// XM has one sustain point, where IT holds a sustain loop
		if s := int(b[4]); s < n {
			env.flags |= 2
			env.sustain = uint8(s)
		} else {
			c.rep.add(what + " envelope sustain is dropped")
		}
	}

	return env, nil
}

// instrument converts the instrument at off, which is 0 for an empty instrument
func (c *conv) instrument(off int) (loadtext.Instrument, error) {
	var li loadtext.Instrument

	if off == 0 {
		return li, nil
	}

	h, err := get(c.d, off, instLen)
	if err != nil {
		return li, err
	}

	if string(h[:4]) != "IMPI" {
		return li, errors.New("missing IMPI header")
	}

	copy(li.Name[:], h[0x20:0x20+22])

	if h[0x11] != 0 || h[0x12] != 0 {
		c.rep.add("new note actions and duplicate checks are not kept")
	}

	li.VolumeFadeout = uint16(min(65535, u16(h, 0x14)*64))

	if h[0x1D4]&1 != 0 {
		c.rep.add("pitch envelopes are dropped")
	}

	vol, err := c.envelope(off+0x130, "volume")
	if err != nil {
		return li, err
	}

	if vol.flags == 0 {
		// with no volume envelope, a flat one is kept on, which lets a note off fade out
		vol = envelope{points: []types.EnvPoint{{X: 0, Y: 64}}, flags: 1}
	}

	li.VolPoints = vol.points
	li.VolFlags = vol.flags
	li.VolSustain = vol.sustain
	li.VolLoopStart = vol.start
	li.VolLoopEnd = vol.end

	pan, err := c.envelope(off+0x182, "panning")
	if err != nil {
		return li, err
	}

	if pan.flags != 0 {
		li.PanPoints = pan.points
		li.PanFlags = pan.flags
		li.PanSustain = pan.sustain
		li.PanLoopStart = pan.start
		li.PanLoopEnd = pan.end
	}

	instPan := h[0x19]
	idx := map[int]int{}
	vibrato := false

	for n := range 96 {
		t, s := int(h[0x40+2*n]), int(h[0x41+2*n])

		if s == 0 {
			continue
		}

		if t != n {
			c.rep.add("keyboard note remapping is not kept")
		}

		x, ok := idx[s]
		if !ok {
			smp, vib, found, err := c.sample(s, instPan)
			if err != nil {
				return li, err
			}

			x = -1

			if found {
				if len(li.Samples) == maxXMSmps {
					return li, errors.New("uses more than 16 samples")
				}

				li.Samples = append(li.Samples, smp)
				x = len(li.Samples) - 1

				if !vibrato && vib[1] != 0 {
					vibrato = true
					// IT depth runs to 64 and XM depth to 15
					li.VibratoRate, li.VibratoDepth, li.VibratoSweep = vib[0], uint8(min(15, int(vib[1])/4)), vib[2]
					li.VibratoType = itVibratoType(vib[3])
				}
			}

			idx[s] = x
		}

		if x >= 0 {
			li.Keymap[n] = uint8(x)
		}
	}

	if vibrato {
		c.rep.add("sample vibrato is approximated")
	}

	return li, nil
}

// itVibratoType maps an IT vibrato waveform to the XM one
func itVibratoType(t byte) uint8 {
	switch t {
	case 1:
		return 2
	case 2:
		return 1
	case 3:
		return 3
	}

	return 0
}

// sample converts IT sample number n, from 1. found is false when the sample is empty
func (c *conv) sample(n int, instPan byte) (loadtext.Sample, [4]uint8, bool, error) {
	var (
		s loadtext.Sample
		vib [4]uint8
	)

	if n < 1 || n > len(c.smpOff) || c.smpOff[n-1] == 0 {
		return s, vib, false, nil
	}

	h, err := get(c.d, c.smpOff[n-1], sampleLen)
	if err != nil {
		return s, vib, false, err
	}

	if string(h[:4]) != "IMPS" {
		return s, vib, false, fmt.Errorf("sample %d: missing IMPS header", n)
	}

	flg := h[0x12]
	frames := u32(h, 0x30)

	if flg&smpExists == 0 || frames == 0 {
		return s, vib, false, nil
	}

	if flg&smpCompress != 0 {
		return s, vib, false, fmt.Errorf("sample %d is compressed, which is not supported", n)
	}

	sixteen := flg&smp16Bit != 0
	signed := h[0x2E]&1 != 0
	chans, bps := 1, 1

	if flg&smpStereo != 0 {
		chans = 2
		c.rep.add("stereo samples are mixed to mono")
	}

	if sixteen {
		bps = 2
	}

	raw, err := get(c.d, u32(h, 0x48), frames*chans*bps)
	if err != nil {
		return s, vib, false, fmt.Errorf("sample %d: %w", n, err)
	}

	data := make([]uint8, frames*bps)

	for i := range frames {
		sum := 0

		// stereo is stored as all the left frames, then all the right frames
		for ch := range chans {
			pos := (ch*frames + i) * bps

			var v int

			switch {
			case sixteen && signed:
				v = int(int16(binary.LittleEndian.Uint16(raw[pos:])))
			case sixteen:
				v = int(binary.LittleEndian.Uint16(raw[pos:])) - 32768
			case signed:
				v = int(int8(raw[pos]))
			default:
				v = int(raw[pos]) - 128
			}

			sum += v
		}

		v := sum / chans

		if sixteen {
			binary.LittleEndian.PutUint16(data[2*i:], uint16(int16(v)))
		} else {
			data[i] = uint8(int8(v))
		}
	}

	s.Data = data
	s.Sixteen = sixteen

	loopStart, loopEnd := u32(h, 0x34), min(u32(h, 0x38), frames)

	if flg&smpLoop != 0 && loopEnd > loopStart {
		s.LoopMode = 1

		if flg&smpPingPong != 0 {
			s.LoopMode = 2
		}

		s.LoopStart = uint32(loopStart)
		s.LoopLength = uint32(loopEnd - loopStart)
	} else {
		s.LoopLength = uint32(frames)
	}

	// the sample volume is scaled by its global volume, which XM does not have
	s.Volume = uint8(clamp(int(h[0x13])*int(min(h[0x11], 64))/64, 0, 64))

	s.Panning = 128

	switch {
	case instPan&0x80 != 0:
		s.Panning = uint8(min(255, int(instPan&0x7F)*4))
	case h[0x2F]&0x80 != 0:
		s.Panning = uint8(min(255, int(h[0x2F]&0x7F)*4))
	}

	c5 := float64(u32(h, 0x3C))
	if c5 <= 0 {
		c5 = defaultC5
	}

	// XM plays note C-4 at 8363 Hz, and IT plays C-5 at C5Speed, so the relative note is one octave lower
	total := 12*math.Log2(c5/defaultC5) - 12
	rel := math.Round(total)
	ft := math.Round((total - rel) * 128)

	if rel < -96 || rel > 95 {
		c.rep.add("sample rate is out of XM's range and is clamped")
	}

	s.RelativeNote = int8(clamp(int(rel), -96, 95))
	s.Finetune = int8(clamp(int(ft), -128, 127))

	copy(s.Name[:], h[0x14:0x14+22])

	vib = [4]uint8{h[0x4C], h[0x4D], h[0x4E], h[0x4F]}

	return s, vib, true, nil
}
