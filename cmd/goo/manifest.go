package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/casselc/goo/annotations"
	"io"
)

func manifest(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("manifest", flag.ContinueOnError)
	dir := fs.String("dir", "", "one Go source package directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dir == "" || fs.NArg() != 0 {
		return fmt.Errorf("require -dir, without positional arguments")
	}
	m, err := annotations.ScanDir(*dir)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}
