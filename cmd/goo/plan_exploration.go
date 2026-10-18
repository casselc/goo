package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/casselc/goo/exploration"
	"io"
	"os"
)

func planExploration(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("plan-exploration", flag.ContinueOnError)
	path := fs.String("query", "", "bounded exploration question JSON")
	if e := fs.Parse(args); e != nil {
		return e
	}
	if *path == "" || fs.NArg() != 0 {
		return fmt.Errorf("require -query JSON file")
	}
	f, e := os.Open(*path)
	if e != nil {
		return e
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil {
		return e
	}
	if info.Size() > 16384 {
		return fmt.Errorf("query exceeds 16 KiB")
	}
	d := json.NewDecoder(io.LimitReader(f, 16385))
	d.DisallowUnknownFields()
	var q exploration.Query
	if e = d.Decode(&q); e != nil {
		return e
	}
	if d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("expected one JSON object")
	}
	p, e := exploration.Compile(q)
	if e != nil {
		return e
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(struct {
		Question exploration.Query `json:"question"`
		Plan     exploration.Plan  `json:"plan"`
	}{q, p})
}
