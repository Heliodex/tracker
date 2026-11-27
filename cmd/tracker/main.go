// Command tracker converts FastTracker II (.xm) modules to and from a text format.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Heliodex/tracker/load"
	"github.com/Heliodex/tracker/loadtext"
	"github.com/Heliodex/tracker/project"
	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/save"
	"github.com/Heliodex/tracker/wav"
)

type options struct {
	out, dir string
	force bool
}

const usage = `usage:
  tracker export [-d dir] [-o out.txt] [-f] module.xm
      convert an XM module to the text format
      modules with instruments become a folder: dir/song.txt and dir/samples/*.wav
      (dir defaults to the module name), other modules print the text to stdout
  tracker build [-o out.xm] [-f] [-no-opus] module.txt
  tracker build [-o out.xm] [-f] [-no-opus] folder
      convert a text module, or a folder holding song.txt, to XM (module.xm by default),
      and render it beside the XM as an Opus file (module.opus), using openmpt123 and ffmpeg
  tracker build [-f] [-no-opus] folder
      a folder holding song.go is built by running it with go run; the program writes the module,
      and the Opus file is rendered beside it
  tracker render -o out.wav [-f] recipe.txt
      render a sample recipe to a WAV file, to listen to it
  tracker new [-name NAME] [-channels N] [-rows N] [-f] folder
      create a folder with an empty song.txt to start from

Flags:
  -d DIR    folder to export a module with instruments into
  -o FILE   output file, "-" for stdout
  -f        overwrite existing files

Use "-" as the input to read from stdin. Reading from stdin needs -o with build,
and -d with export when the module has instruments.
`

var (
	// errUsage is returned for bad command lines, after the usage text has been printed
	errUsage = errors.New("bad usage")

	// opusTools are the programs that render an Opus file
	opusTools = []string{"openmpt123", "ffmpeg"}
)

// checkOpusTools returns an error if a program that renders Opus is not installed
func checkOpusTools() error {
	for _, tool := range opusTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("rendering Opus needs %s (install it, or use -no-opus): %w", tool, err)
		}
	}

	return nil
}

// newFlagSet returns the flags shared by every command
func newFlagSet(name string, stderr io.Writer, o *options) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)

	fs.SetOutput(stderr)
	fs.StringVar(&o.out, "o", "", `output file, "-" for stdout`)
	fs.StringVar(&o.dir, "d", "", "folder for a module with instruments")
	fs.BoolVar(&o.force, "f", false, "overwrite existing files")

	return fs
}

// opusPath gives the Opus file beside an XM file
func opusPath(xm string) string {
	return strings.TrimSuffix(xm, filepath.Ext(xm)) + ".opus"
}

// parseArgs parses flags and returns the positional arguments, allowing flags after them
func parseArgs(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string

	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}

		args = fs.Args()

		if len(args) == 0 {
			return positional, nil
		}

		positional = append(positional, args[0])
		args = args[1:]
	}
}

// newSong creates a folder with an empty song.txt to start from
func newSong(args []string, stderr io.Writer) error {
	var (
		name string
		channels, rows int
		force bool
	)

	fs := flag.NewFlagSet("new", flag.ContinueOnError)

	fs.SetOutput(stderr)
	fs.StringVar(&name, "name", "", "song name, up to 20 characters")
	fs.IntVar(&channels, "channels", 4, "number of channels")
	fs.IntVar(&rows, "rows", 64, "rows in the first pattern")
	fs.BoolVar(&force, "f", false, "overwrite an existing song.txt")

	positional, err := parseArgs(fs, args)
	if err != nil {
		return err
	}

	if len(positional) != 1 {
		fmt.Fprintf(stderr, "tracker new: expected one folder, got %d\n\n%s", len(positional), usage)
		return errUsage
	}

	return project.NewSong(positional[0], name, channels, rows, force)
}

// parseOne parses a command line that takes exactly one input
func parseOne(name string, args []string, stderr io.Writer, o *options) (string, error) {
	fs := newFlagSet(name, stderr, o)

	positional, err := parseArgs(fs, args)
	if err != nil {
		return "", err
	}

	if len(positional) != 1 {
		fmt.Fprintf(stderr, "tracker %s: expected one input, got %d\n\n%s", name, len(positional), usage)
		return "", errUsage
	}

	return positional[0], nil
}

// readInput reads the named file, or stdin for "-"
func readInput(path string, stdin io.Reader) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(stdin)
	}

	return os.ReadFile(path)
}

// refuseExisting returns an error if path already exists
func refuseExisting(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists (use -f to overwrite)", path)
	}

	return nil
}

// renderOpus renders an XM module to a 48 kHz stereo Opus file. openmpt123 writes raw audio into a pipe that ffmpeg
// encodes, so no WAV file is written. The Opus file is removed if either program fails.
func renderOpus(xm, opus string, force bool, stderr io.Writer) error {
	if !force {
		if err := refuseExisting(opus); err != nil {
			return err
		}
	}

	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	defer r.Close()

	player := exec.Command("openmpt123", "--batch", "--stdout", "--quiet", "--no-float", "--samplerate", "48000", "--channels", "2", "--", xm)
	player.Stdout = w
	player.Stderr = stderr
	encoder := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-f", "s16le", "-ar", "48000", "-ac", "2", "-i", "pipe:0", "-c:a", "libopus", opus)
	encoder.Stdin = r
	encoder.Stderr = stderr

	if err := encoder.Start(); err != nil {
		w.Close()
		return fmt.Errorf("ffmpeg: %w", err)
	}

	if err := player.Start(); err != nil {
		// closing the pipe ends ffmpeg's input, so it stops
		w.Close()
		encoder.Wait()

		return fmt.Errorf("openmpt123: %w", err)
	}

	// the parent's copy of the write end must close, or ffmpeg never sees the end of the audio
	w.Close()

	playErr := player.Wait()
	if encErr := encoder.Wait(); playErr != nil || encErr != nil {
		os.Remove(opus)

		if playErr != nil {
			return fmt.Errorf("openmpt123 %s: %w", xm, playErr)
		}

		return fmt.Errorf("ffmpeg %s: %w", opus, encErr)
	}

	return nil
}

// runSongGo builds a song folder by running its song.go, which writes its own module.
// The folder must be inside a Go module, such as this repository.
func runSongGo(dir string, stdout, stderr io.Writer) error {
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: go run failed: %w", dir, err)
	}

	return nil
}

// songGoDir returns the folder holding a song.go, when in is that folder or the song.go itself
func songGoDir(in string) (string, bool) {
	if in == "-" {
		return "", false
	}

	info, err := os.Stat(in)
	if err != nil {
		return "", false
	}

	dir := in

	if !info.IsDir() {
		if filepath.Base(in) != "song.go" {
			return "", false
		}

		dir = filepath.Dir(in)
	}

	if _, err := os.Stat(filepath.Join(dir, "song.go")); err != nil {
		return "", false
	}

	return dir, true
}

// writeOutput writes data to path, or stdout for "-", refusing to overwrite unless force is set
func writeOutput(path string, force bool, data []byte, stdout io.Writer) error {
	if path == "-" {
		_, err := stdout.Write(data)

		return err
	}

	if !force {
		if err := refuseExisting(path); err != nil {
			return err
		}
	}

	return os.WriteFile(path, data, 0o644)
}

func build(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var (
		o options
		noOpus bool
	)

	fs := newFlagSet("build", stderr, &o)

	fs.BoolVar(&noOpus, "no-opus", false, "do not render the module to an Opus file beside it")

	positional, err := parseArgs(fs, args)
	if err != nil {
		return err
	}

	if len(positional) != 1 {
		fmt.Fprintf(stderr, "tracker build: expected one input, got %d\n\n%s", len(positional), usage)
		return errUsage
	}

	in := positional[0]

	if dir, ok := songGoDir(in); ok {
		if o.out != "" {
			return errors.New("a song.go writes its own output, so -o does not apply")
		}

		if !noOpus {
			if err := checkOpusTools(); err != nil {
				return err
			}
		}

		if err := runSongGo(dir, stdout, stderr); err != nil {
			return err
		}

		if noOpus {
			return nil
		}

		// the program has just written song.xm, so its Opus file replaces any older one
		return renderOpus(filepath.Join(dir, "song.xm"), filepath.Join(dir, "song.opus"), true, stderr)
	}

	// a folder means its song.txt
	if in != "-" {
		if info, err := os.Stat(in); err == nil && info.IsDir() {
			in = filepath.Join(in, project.SongFile)
		}
	}

	out := o.out
	if out == "" {
		if in == "-" {
			return errors.New("reading from stdin needs -o")
		}

		out = strings.TrimSuffix(in, filepath.Ext(in)) + ".xm"
	}

	if in != "-" && out == in {
		return fmt.Errorf("output would overwrite the input %s", in)
	}

	// Opus is rendered beside a file, so it is skipped when the module goes to stdout. Everything is checked
	// before anything is written, so a missing tool or an existing file doesn't leave a half-finished build.
	wantOpus := !noOpus && out != "-"
	opus := opusPath(out)

	if wantOpus {
		if err := checkOpusTools(); err != nil {
			return err
		}

		if !o.force {
			if err := refuseExisting(opus); err != nil {
				return err
			}
		}
	}

	data, err := readInput(in, stdin)
	if err != nil {
		return err
	}

	// samples are read relative to the text file, or the current folder for stdin
	dir := "."

	if in != "-" {
		dir = filepath.Dir(in)
	}

	f, err := project.Build(bytes.NewReader(data), dir)
	if err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	var xm bytes.Buffer

	if err := save.Write(&xm, f); err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	if err := writeOutput(out, o.force, xm.Bytes(), stdout); err != nil {
		return err
	}

	if !wantOpus {
		return nil
	}

	return renderOpus(out, opus, o.force, stderr)
}

func export(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var o options

	in, err := parseOne("export", args, stderr, &o)
	if err != nil {
		return err
	}

	data, err := readInput(in, stdin)
	if err != nil {
		return err
	}

	f, err := load.Read(bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	if f.Head.NumInstruments > 0 {
		if o.out != "" {
			return errors.New("a module with instruments is exported to a folder, use -d instead of -o")
		}

		dir := o.dir
		if dir == "" {
			if in == "-" {
				return errors.New("exporting from stdin needs -d")
			}

			dir = strings.TrimSuffix(in, filepath.Ext(in))
		}

		return project.Export(f, dir, o.force)
	}

	if o.dir != "" {
		return errors.New("-d is only used for a module with instruments, use -o for text")
	}

	tf, err := loadtext.FromFile(f)
	if err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	var text bytes.Buffer

	if err := loadtext.WriteText(&text, tf); err != nil {
		return err
	}

	if o.out == "" {
		o.out = "-"
	}

	return writeOutput(o.out, o.force, text.Bytes(), stdout)
}

// render writes a sample recipe as a WAV file. It needs -o, so the WAV is never written beside the recipe,
// where it would make the sample ambiguous
func render(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var o options

	in, err := parseOne("render", args, stderr, &o)
	if err != nil {
		return err
	}

	if o.out == "" {
		return errors.New("render needs -o to choose where to write the WAV")
	}

	if o.out != "-" && !strings.HasSuffix(o.out, ".wav") {
		return fmt.Errorf("render writes WAV data, so -o must end in .wav, got %s", o.out)
	}

	if in != "-" && filepath.Clean(strings.TrimSuffix(o.out, ".wav")+".txt") == filepath.Clean(in) {
		return fmt.Errorf("writing %s beside %s would make the sample ambiguous", o.out, in)
	}

	data, err := readInput(in, stdin)
	if err != nil {
		return err
	}

	pcm, sixteen, err := recipe.Render(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	file, err := wav.Encode(pcm, sixteen)
	if err != nil {
		return err
	}

	return writeOutput(o.out, o.force, file, stdout)
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return errUsage
	}

	switch args[0] {
	case "export":
		return export(args[1:], stdin, stdout, stderr)
	case "build":
		return build(args[1:], stdin, stdout, stderr)
	case "render":
		return render(args[1:], stdin, stdout, stderr)
	case "new":
		return newSong(args[1:], stderr)
	case "--help", "-h", "-help", "help":
		fmt.Fprint(stdout, usage)
		return nil
	}

	fmt.Fprintf(stderr, "tracker: unknown command %q\n\n%s", args[0], usage)

	return errUsage
}

func main() {
	err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	if err == nil || errors.Is(err, flag.ErrHelp) {
		return
	}

	if !errors.Is(err, errUsage) {
		fmt.Fprintln(os.Stderr, "tracker:", err)
	}

	os.Exit(1)
}
