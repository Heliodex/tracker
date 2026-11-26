package wav

import (
	"bytes"
	"testing"
)

func TestDecodeRejectsNonWAV(t *testing.T) {
	if _, _, err := Decode([]byte("not a wav file at all")); err == nil {
		t.Error("expected an error")
	}
}

func TestDecodeRejectsStereo(t *testing.T) {
	file, err := Encode([]uint8{1, 2}, false)
	if err != nil {
		t.Fatal(err)
	}

	// channel count is at byte 22 of the file
	file[22] = 2

	if _, _, err := Decode(file); err == nil {
		t.Error("expected an error")
	}
}

func TestEightBitIsUnsignedInFile(t *testing.T) {
	// XM value -128 (0x80) is the lowest sample, which is 0 in unsigned WAV
	file, err := Encode([]uint8{0x80}, false)
	if err != nil {
		t.Fatal(err)
	}

	if got := file[len(file)-1]; got != 0x00 {
		t.Errorf("sample byte in file is %#x, want 0x00", got)
	}
}

func TestOdd16BitRejected(t *testing.T) {
	if _, err := Encode([]uint8{1, 2, 3}, true); err == nil {
		t.Error("expected an error")
	}
}

func TestRoundTrip16Bit(t *testing.T) {
	pcm := []uint8{0x00, 0x80, 0xFF, 0x7F, 0x34, 0x12}

	file, err := Encode(pcm, true)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, sixteen, err := Decode(file)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if !sixteen {
		t.Error("got 8-bit, want 16-bit")
	}

	if !bytes.Equal(got, pcm) {
		t.Errorf("got %v, want %v", got, pcm)
	}
}

func TestRoundTrip8Bit(t *testing.T) {
	// odd length, so the pad byte is exercised
	pcm := []uint8{0x00, 0x7F, 0x80, 0xFF, 0x01}

	file, err := Encode(pcm, false)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	got, sixteen, err := Decode(file)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if sixteen {
		t.Error("got 16-bit, want 8-bit")
	}

	if !bytes.Equal(got, pcm) {
		t.Errorf("got %v, want %v", got, pcm)
	}
}
