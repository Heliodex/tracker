// Command it2xm converts an Impulse Tracker (.it) module to a FastTracker II (.xm) module.
//
// Usage:
//
//	it2xm [-o out.xm] module.it
//
// Without -o, the output is the input path with an .xm extension. What XM cannot represent is listed on stderr.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Heliodex/tracker/loadit"
	"github.com/Heliodex/tracker/save"
)

func run(args []string, stdout, stderr *os.File) error {
	fs := flag.NewFlagSet("it2xm", flag.ContinueOnError)
	fs.SetOutput(stderr)

	out := fs.String("o", "", "output .xm path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return fmt.Errorf("usage: it2xm [-o out.xm] module.it")
	}

	in := fs.Arg(0)

	if *out == "" {
		*out = strings.TrimSuffix(in, filepath.Ext(in)) + ".xm"
	}

	data, err := os.ReadFile(in)
	if err != nil {
		return err
	}

	tf, warnings, err := loadit.Read(data)
	if err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	f, err := tf.ToFile()
	if err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	var b bytes.Buffer

	if err := save.Write(&b, f); err != nil {
		return fmt.Errorf("%s: %w", in, err)
	}

	if err := os.WriteFile(*out, b.Bytes(), 0o644); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "wrote %s\n", *out)

	for _, w := range warnings {
		fmt.Fprintf(stderr, "warning: %s\n", w)
	}

	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
