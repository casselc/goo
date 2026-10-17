package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/casselc/goo/annotations"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// weave is deliberately test-only in this first slice. It requires an opted-in
// module, validates one source package, and never edits dependency declarations.
func weave(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("weave", flag.ContinueOnError)
	module := fs.String("module", "", "opted-in Go module directory")
	dir := fs.String("dir", ".", "source package relative to module")
	pattern := fs.String("test", "", "Go test package pattern")
	output := fs.String("manifest", "", "manifest output path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *module == "" || *pattern == "" || *output == "" || fs.NArg() != 0 {
		return fmt.Errorf("weave requires -module, -test and -manifest")
	}
	absolute, err := filepath.Abs(*module)
	if err != nil {
		return err
	}
	m, err := annotations.ScanDir(filepath.Join(absolute, *dir))
	if err != nil {
		return err
	}
	if _, err = os.Stat(filepath.Join(absolute, "orchestrion.tool.go")); err != nil {
		return fmt.Errorf("module must explicitly opt in with orchestrion.tool.go")
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(*output), 0755); err != nil {
		return err
	}
	if err = os.WriteFile(*output, append(data, '\n'), 0644); err != nil {
		return err
	}
	fmt.Fprintf(out, "Validated manifest %s (%s)\n", *output, m.Revision)
	command := exec.Command("go", "run", "github.com/DataDog/orchestrion", "go", "test", *pattern, "-v")
	command.Dir = absolute
	command.Stdout = out
	command.Stderr = out
	return command.Run()
}
