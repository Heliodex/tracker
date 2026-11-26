package project

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNewSongBuilds(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "mysong")

	if err := NewSong(dir, "my song", 4, 64, false); err != nil {
		t.Fatalf("NewSong: %v", err)
	}

	f, err := Build(strings.NewReader(readFile(t, filepath.Join(dir, SongFile))), dir)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if f.Head.NumChannels != 4 || f.Head.NumPatterns != 1 || len(f.Patterns[0].Data) != 64 {
		t.Errorf("got %d channels, %d patterns", f.Head.NumChannels, f.Head.NumPatterns)
	}
}

func TestNewSongRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()

	if err := NewSong(dir, "", 4, 64, false); err != nil {
		t.Fatal(err)
	}

	if err := NewSong(dir, "", 4, 64, false); err == nil {
		t.Error("expected an error without force")
	}

	if err := NewSong(dir, "", 4, 64, true); err != nil {
		t.Errorf("forced: %v", err)
	}
}

func TestNewSongValidates(t *testing.T) {
	dir := t.TempDir()

	if err := NewSong(dir, "", 0, 64, false); err == nil {
		t.Error("expected an error for 0 channels")
	}

	if err := NewSong(dir, "", 4, 257, false); err == nil {
		t.Error("expected an error for 257 rows")
	}

	if err := NewSong(dir, strings.Repeat("x", 21), 4, 64, false); err == nil {
		t.Error("expected an error for a long name")
	}
}
