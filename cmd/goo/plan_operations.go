package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/casselc/goo/annotations"
	"io"
)

func planOperations(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("plan-operations", flag.ContinueOnError)
	path := fs.String("package", "", "declarative operation package YAML")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" || fs.NArg() != 0 {
		return fmt.Errorf("require -package, without positional arguments")
	}
	p, err := annotations.LoadOperationPackage(*path)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}
