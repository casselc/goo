package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/casselc/goo/prebuilt"
	"io"
	"os"
)

func expandPackage(command string, args []string, out io.Writer) error {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	binding := fs.String("binding", "", "explicit YAML package binding")
	inventory := fs.String("inventory", "", "versioned telemetry inventory JSON")
	requests := fs.String("requests", "", "normalized selected HTTP requests JSON array")
	evaluation := fs.String("evaluation", "", "evaluation identity for normalized input")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 || *binding == "" || *inventory == "" {
		return fmt.Errorf("require -binding and -inventory, without positional arguments")
	}
	bf, err := os.Open(*binding)
	if err != nil {
		return err
	}
	defer bf.Close()
	b, err := prebuilt.ReadBinding(bf)
	if err != nil {
		return err
	}
	inf, err := os.Open(*inventory)
	if err != nil {
		return err
	}
	defer inf.Close()
	in, err := prebuilt.ReadInventory(inf)
	if err != nil {
		return err
	}
	expanded, err := prebuilt.Expand(in, b)
	if err != nil {
		return err
	}
	var result any = expanded
	if command == "prepare-http" {
		if *requests == "" || *evaluation == "" {
			return fmt.Errorf("prepare-http requires -requests and -evaluation")
		}
		data, err := os.ReadFile(*requests)
		if err != nil {
			return err
		}
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		var input []prebuilt.HTTPRequest
		if err = dec.Decode(&input); err != nil {
			return err
		}
		if err = dec.Decode(new(any)); err != io.EOF {
			return fmt.Errorf("requests require one JSON document")
		}
		result, err = expanded.PrepareHTTP(input, *evaluation)
		if err != nil {
			return err
		}
	} else if *requests != "" || *evaluation != "" {
		return fmt.Errorf("request/evaluation flags belong to prepare-http")
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
