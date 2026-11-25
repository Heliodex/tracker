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
	"github.com/Heliodex/tracker/save"
)

const usage = `usage:
  tracker export [-o out.txt] [-f] module.xm
      convert an XM module to the text format (stdout by default)
  tracker build [-o out.xm] [-f] module.txt
      convert a text module to XM (module.xm by default)

Flags:
  -o FILE   output file, "-" for stdout
  -f        overwrite the output file if it exists

Use "-" as the input to read from stdin. Reading from stdin with build needs -o.
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
	force bool
}

// newFlagSet returns the flags shared by every command
func newFlagSet(name string, stderr io.Writer, o *options) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&o.out, "o", "", `output file, "-" for stdout`)
	fs.BoolVar(&o.force, "f", false, "overwrite the output file if it exists")

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

// parseOne parses a command line that takes exactly one input file
func parseOne(name string, args []string, stderr io.Writer, o *options) (string, error) {
	fs := newFlagSet(name, stderr, o)

	positional, err := parseArgs(fs, args)
	if err != nil {
		return "", err
	}

	if len(positional) != 1 {
		fmt.Fprintf(stderr, "tracker %s: expected one input file, got %d\n\n%s", name, len(positional), usage)
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

	if n := f.Head.NumInstruments; n > 0 {
		fmt.Fprintf(stderr, "warning: %d instruments dropped, the text format does not store them\n", n)
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

	tf, err := loadtext.ReadText(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	f, err := tf.ToFile()
	if err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	var xm bytes.Buffer

	if err := save.Write(&xm, f); err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	return writeOutput(out, o.force, xm.Bytes(), stdout)
}
