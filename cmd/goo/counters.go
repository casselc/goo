package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/casselc/goo/metrics"
	"io"
	"os"
)

func counters(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("evaluate-counters", flag.ContinueOnError)
	input := fs.String("input", "", "counter request JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" || fs.NArg() != 0 {
		return fmt.Errorf("require -input")
	}
	f, e := os.Open(*input)
	if e != nil {
		return e
	}
	defer f.Close()
	var request metrics.CounterRequest
	data, e := io.ReadAll(io.LimitReader(f, (8<<20)+1))
	if e != nil {
		return e
	}
	if len(data) > 8<<20 {
		return fmt.Errorf("counter request exceeds 8 MiB")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if e = d.Decode(&request); e != nil {
		return e
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return fmt.Errorf("require one counter request")
	}
	result, e := metrics.EvaluateCounters(request)
	if e != nil {
		return e
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
