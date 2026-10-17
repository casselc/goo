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

func histograms(args []string, out io.Writer, merge bool) error {
	fs := flag.NewFlagSet("histograms", flag.ContinueOnError)
	input := fs.String("input", "", "histogram request or explicit merge members JSON")
	if e := fs.Parse(args); e != nil {
		return e
	}
	if *input == "" || fs.NArg() != 0 {
		return fmt.Errorf("require -input")
	}
	f, e := os.Open(*input)
	if e != nil {
		return e
	}
	defer f.Close()
	data, e := io.ReadAll(io.LimitReader(f, (8<<20)+1))
	if e != nil {
		return e
	}
	if len(data) > 8<<20 {
		return fmt.Errorf("histogram input exceeds 8 MiB")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	var result any
	if merge {
		var members []metrics.HistogramMember
		if e = d.Decode(&members); e != nil {
			return e
		}
		result, e = metrics.MergeAlignedHistograms(members)
	} else {
		var request metrics.HistogramRequest
		if e = d.Decode(&request); e != nil {
			return e
		}
		result, e = metrics.EvaluateHistograms(request)
	}
	if e != nil {
		return e
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return fmt.Errorf("require one JSON document")
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
