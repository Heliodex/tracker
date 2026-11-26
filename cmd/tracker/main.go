// Command tracker converts FastTracker II (.xm) modules to and from a text format.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Heliodex/tracker/load"
	"github.com/Heliodex/tracker/loadtext"
	"github.com/Heliodex/tracker/project"
	"github.com/Heliodex/tracker/recipe"
	"github.com/Heliodex/tracker/save"
	"github.com/Heliodex/tracker/wav"
)

const usage = `usage:
  tracker export [-d dir] [-o out.txt] [-f] module.xm
      convert an XM module to the text format
      modules with instruments become a folder: dir/song.txt and dir/samples/*.wav
      (dir defaults to the module name), other modules print the text to stdout
  tracker build [-o out.xm] [-f] module.txt
  tracker build [-o out.xm] [-f] folder
      convert a text module, or a folder holding song.txt, to XM (module.xm by default)
  tracker render -o out.wav [-f] recipe.txt
      render a sample recipe to a WAV file, to listen to it

Flags:
  -d DIR    folder to export a module with instruments into
  -o FILE   output file, "-" for stdout
  -f        overwrite existing files

Use "-" as the input to read from stdin. Reading from stdin needs -o with build,
and -d with export when the module has instruments.
`

// errUsage is returned for bad command lines, after the usage text has been printed
var errUsage = errors.New("bad usage")

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
	case "help", "-h", "-help", "--help":
		fmt.Fprint(stdout, usage)
		return nil
	default:
		fmt.Fprintf(stderr, "tracker: unknown command %q\n\n%s", args[0], usage)
		return errUsage
	}
}

type options struct {
	out string
	dir string
	force bool
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

// writeOutput writes data to path, or stdout for "-", refusing to overwrite unless force is set
func writeOutput(path string, force bool, data []byte, stdout io.Writer) error {
	if path == "-" {
		_, err := stdout.Write(data)

		return err
	}

	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists (use -f to overwrite)", path)
		}
	}

	return os.WriteFile(path, data, 0o644)
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

func build(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var o options

	in, err := parseOne("build", args, stderr, &o)
	if err != nil {
		return err
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

	return writeOutput(out, o.force, xm.Bytes(), stdout)
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
