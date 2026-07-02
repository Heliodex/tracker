package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"testing"

	"github.com/Heliodex/tracker/load"
	"github.com/Heliodex/tracker/save"
)

func convert16(data []uint8) error {
	converted := load.ConvertSample16Bit(data)
	og := save.UnconvertSample16Bit(converted)

	if len(og) != len(data) {
		return errors.New("length mismatch")
	}

	for i := range og {
		if og[i] != data[i] {
			return fmt.Errorf("data mismatch at index %d: got %d, want %d", i, og[i], data[i])
		}
	}

	return nil
}

func convert8(data []uint8) error {
	converted := load.ConvertSample8Bit(data)
	og := save.UnconvertSample8Bit(converted)

	if len(og) != len(data) {
		return errors.New("length mismatch")
	}

	for i := range og {
		if og[i] != data[i] {
			return fmt.Errorf("data mismatch at index %d: got %d, want %d", i, og[i], data[i])
		}
	}

	return nil
}

func TestConvert(t *testing.T) {
	randomData := make([]uint8, 10000)
	rand.Read(randomData)

	if err := convert16(randomData); err != nil {
		t.Errorf("16-bit conversion failed: %v", err)
	}

	if err := convert8(randomData); err != nil {
		t.Errorf("8-bit conversion failed: %v", err)
	}
}
