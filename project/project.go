// Package project reads and writes a module as a folder: a song.txt in the text format, and a samples
// directory holding each sample as a WAV file.
//
//	song/
//	  song.txt
//	  samples/
//	    00-00-kick.wav
package project

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Heliodex/tracker/loadtext"
	"github.com/Heliodex/tracker/types"
	"github.com/Heliodex/tracker/wav"
)

// SongFile is the name of the text file inside a project folder
const SongFile = "song.txt"

// trimName returns a fixed-size name without its trailing NUL padding
func trimName(name []uint8) string {
	return strings.TrimRight(string(name), "\x00")
}

// slug turns a sample name into part of a file name
func slug(name string) string {
	var b strings.Builder

	dash := true

	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)

			dash = false
		} else if !dash {
			b.WriteByte('-')

			dash = true
		}
	}

	s := strings.Trim(b.String(), "-")
	if s == "" {
		return "sample"
	}

	if len(s) > 24 {
		s = strings.TrimRight(s[:24], "-")
	}

	return s
}

// Export writes the module f into the folder dir, as dir/song.txt and dir/samples/*.wav.
// Existing files are only replaced when force is set.
func Export(f *types.File, dir string, force bool) error {
	tf, err := loadtext.FromFile(f)
	if err != nil {
		return err
	}

	type wavFile struct {
		path string
		data []byte
	}

	var wavs []wavFile

	for i := range tf.Instruments {
		for s := range tf.Instruments[i].Samples {
			sample := &tf.Instruments[i].Samples[s]

			sample.File = fmt.Sprintf("samples/%02d-%02d-%s.wav", i, s, slug(trimName(sample.Name[:])))

			data, err := wav.Encode(sample.Data, sample.Sixteen)
			if err != nil {
				return fmt.Errorf("instrument %d sample %d: %w", i, s, err)
			}

			wavs = append(wavs, wavFile{filepath.Join(dir, filepath.FromSlash(sample.File)), data})
		}
	}

	var text bytes.Buffer

	if err := loadtext.WriteText(&text, tf); err != nil {
		return err
	}

	songPath := filepath.Join(dir, SongFile)

	if !force {
		paths := []string{songPath}

		for _, w := range wavs {
			paths = append(paths, w.path)
		}

		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("%s already exists (use -f to overwrite)", path)
			}
		}
	}

	if err := os.MkdirAll(filepath.Join(dir, "samples"), 0o755); err != nil {
		return err
	}

	for _, w := range wavs {
		if err := os.WriteFile(w.path, w.data, 0o644); err != nil {
			return err
		}
	}

	return os.WriteFile(songPath, text.Bytes(), 0o644)
}

// Build reads the text-format module from r, and loads the WAV file for each sample from inside dir.
// The result is ready for save.Write.
func Build(r io.Reader, dir string) (*types.File, error) {
	tf, err := loadtext.ReadText(r)
	if err != nil {
		return nil, err
	}

	for i := range tf.Instruments {
		for s := range tf.Instruments[i].Samples {
			sample := &tf.Instruments[i].Samples[s]

			rel := filepath.FromSlash(sample.File)
			if !filepath.IsLocal(rel) {
				return nil, fmt.Errorf("instrument %d sample %d: %q is not inside the project folder", i, s, sample.File)
			}

			file, err := os.ReadFile(filepath.Join(dir, rel))
			if err != nil {
				return nil, fmt.Errorf("instrument %d sample %d: %w", i, s, err)
			}

			pcm, sixteen, err := wav.Decode(file)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", sample.File, err)
			}

			sample.Data = pcm
			sample.Sixteen = sixteen
		}
	}

	return tf.ToFile()
}
