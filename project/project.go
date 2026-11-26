// Package project reads and writes a module as a folder: a song.txt in the text format, and a samples
// directory holding each sample as a WAV file, or as a recipe that is rendered on build.
//
//	song/
//	  song.txt
//	  samples/
//	    00-00-kick.wav     or     00-00-kick.txt
//
// A sample's File is its path without an extension. Build uses base.wav if it exists, or base.txt
// rendered as a recipe, and fails if both exist. A File ending in .wav or .txt is accepted too.
package project

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Heliodex/tracker/loadtext"
	"github.com/Heliodex/tracker/recipe"
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

// exists reports whether path exists, returning any error other than it not existing
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	return false, err
}

// Export writes the module f into the folder dir, as dir/song.txt and dir/samples/*.wav.
// Existing files are only replaced when force is set. A sample that already has a recipe is never replaced.
func Export(f *types.File, dir string, force bool) error {
	tf, err := loadtext.FromFile(f)
	if err != nil {
		return err
	}

	type outFile struct {
		path string
		data []byte
	}

	var wavs []outFile

	for i := range tf.Instruments {
		for s := range tf.Instruments[i].Samples {
			sample := &tf.Instruments[i].Samples[s]

			sample.File = fmt.Sprintf("samples/%02d-%02d-%s", i, s, slug(trimName(sample.Name[:])))

			base := filepath.Join(dir, filepath.FromSlash(sample.File))

			if has, err := exists(base + ".txt"); err != nil {
				return err
			} else if has {
				return fmt.Errorf("%s.txt is a recipe, remove it before exporting over it", base)
			}

			data, err := wav.Encode(sample.Data, sample.Sixteen)
			if err != nil {
				return fmt.Errorf("instrument %d sample %d: %w", i, s, err)
			}

			wavs = append(wavs, outFile{base + ".wav", data})
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
			if has, err := exists(path); err != nil {
				return err
			} else if has {
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

// sampleFile finds the file for a sample reference, which is base.wav or base.txt, but not both.
// It returns true when the file is a recipe.
func sampleFile(dir, ref string) (path string, isRecipe bool, err error) {
	base := strings.TrimSuffix(strings.TrimSuffix(ref, ".wav"), ".txt")
	rel := filepath.FromSlash(base)

	if !filepath.IsLocal(rel) {
		return "", false, fmt.Errorf("%q is not inside the project folder", ref)
	}

	wavPath := filepath.Join(dir, rel+".wav")
	txtPath := filepath.Join(dir, rel+".txt")

	hasWav, err := exists(wavPath)
	if err != nil {
		return "", false, err
	}

	hasTxt, err := exists(txtPath)
	if err != nil {
		return "", false, err
	}

	switch {
	case hasWav && hasTxt:
		return "", false, fmt.Errorf("%s has both a .wav and a .txt recipe, remove one", base)
	case hasWav:
		return wavPath, false, nil
	case hasTxt:
		return txtPath, true, nil
	default:
		return "", false, fmt.Errorf("no sample file for %s, expected a .wav or .txt", base)
	}
}

// loadSample reads the sample data and bit depth from a WAV file, or renders a recipe
func loadSample(path string, isRecipe bool) (pcm []uint8, sixteen bool, err error) {
	if isRecipe {
		fh, err := os.Open(path)
		if err != nil {
			return nil, false, err
		}

		defer fh.Close()

		return recipe.Render(fh)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}

	return wav.Decode(file)
}

// Build reads the text-format module from r, and loads the sample for each sample from inside dir.
// The result is ready for save.Write.
func Build(r io.Reader, dir string) (*types.File, error) {
	tf, err := loadtext.ReadText(r)
	if err != nil {
		return nil, err
	}

	for i := range tf.Instruments {
		for s := range tf.Instruments[i].Samples {
			sample := &tf.Instruments[i].Samples[s]

			path, isRecipe, err := sampleFile(dir, sample.File)
			if err != nil {
				return nil, fmt.Errorf("instrument %d sample %d: %w", i, s, err)
			}

			pcm, sixteen, err := loadSample(path, isRecipe)
			if err != nil {
				return nil, fmt.Errorf("instrument %d sample %d: %w", i, s, err)
			}

			sample.Data = pcm
			sample.Sixteen = sixteen
		}
	}

	return tf.ToFile()
}
