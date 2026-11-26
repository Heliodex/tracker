package project

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Heliodex/tracker/loadtext"
)

// NewSong creates the folder dir with a song.txt holding one empty pattern of rows rows and channels channels.
// The song has no instruments. An existing song.txt is only replaced when force is set.
func NewSong(dir, name string, channels, rows int, force bool) error {
	if channels < 1 || channels > 32 {
		return fmt.Errorf("channels must be between 1 and 32, got %d", channels)
	}

	if rows < 1 || rows > 256 {
		return fmt.Errorf("rows must be between 1 and 256, got %d", rows)
	}

	if len(name) > 20 {
		return fmt.Errorf("name is longer than 20 characters: %q", name)
	}

	pattern := make(loadtext.Pattern, rows)

	for r := range pattern {
		pattern[r] = make([]loadtext.ChannelData, channels)
	}

	tf := &loadtext.TextFormat{
		NumChannels: uint16(channels),
		NumPatterns: 1,
		Speed: 6,
		Tempo: 125,
		LinearSlides: true,
		OrderTable: []uint8{0},
		Patterns: []loadtext.Pattern{pattern},
	}

	copy(tf.Name[:], name)

	var text bytes.Buffer

	if err := loadtext.WriteText(&text, tf); err != nil {
		return err
	}

	songPath := filepath.Join(dir, SongFile)

	if !force {
		has, err := exists(songPath)
		if err != nil {
			return err
		}

		if has {
			return errors.New(songPath + " already exists (use -f to overwrite)")
		}
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(songPath, text.Bytes(), 0o644)
}
