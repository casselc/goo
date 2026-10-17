package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/casselc/goo/annotations"
	"io"
	"os"
)

func observe(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("observe-operations", flag.ContinueOnError)
	dir := fs.String("dir", "", "source package directory")
	input := fs.String("spans", "", "normalized completed operation spans JSON")
	resource := fs.String("resource", "", "explicit selected resource identity")
	basis := fs.String("basis", "", "evidence basis label (not a completeness assertion)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 || *dir == "" || *input == "" {
		return fmt.Errorf("require -dir, -spans, -resource and -basis")
	}
	m, err := annotations.ScanDir(*dir)
	if err != nil {
		return err
	}
	f, err := os.Open(*input)
	if err != nil {
		return err
	}
	defer f.Close()
	var spans []annotations.Observation
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err = dec.Decode(&spans); err != nil {
		return err
	}
	var extra any
	if err = dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("expected one JSON array")
	}
	result, err := m.Observe(*resource, *basis, spans)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
