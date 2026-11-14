package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
)

// Read content, rather than trusting Stat.Size, so pipes and virtual files obey
// the same authoring budget as ordinary files.
func readPlanInput(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("planning input exceeds %d bytes", limit)
	}
	return data, nil
}

func planCEL(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("plan-cel", flag.ContinueOnError)
	path := fs.String("query", "", "existing fixed exploration question JSON")
	expression := fs.String("expression", "", "supported CEL predicate")
	expressionFile := fs.String("expression-file", "", "UTF-8 CEL predicate file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" || fs.NArg() != 0 || (*expression == "") == (*expressionFile == "") {
		return fmt.Errorf("plan-cel requires -query and exactly one of -expression or -expression-file")
	}
	data, err := readPlanInput(*path, 16384)
	if err != nil {
		return err
	}
	var parent exploration.Query
	if err := json.Unmarshal(data, &parent); err != nil {
		return err
	}
	if *expressionFile != "" {
		data, err := readPlanInput(*expressionFile, querycel.MaxExpressionBytes)
		if err != nil {
			return err
		}
		*expression = string(data)
	}
	program, err := querycel.Compile(parent.Signal, *expression)
	if err != nil {
		return err
	}
	question, err := program.Apply(parent)
	if err != nil {
		return err
	}
	plan, err := exploration.Compile(question)
	if err != nil {
		return err
	}
	// This is a source-free planning receipt, not captured telemetry or evidence.
	result := struct {
		Version    string                 `json:"version"`
		CELLibrary string                 `json:"celLibrary"`
		Expression string                 `json:"expression"`
		Predicate  *exploration.Predicate `json:"predicate"`
		Question   exploration.Query      `json:"question"`
		Plan       exploration.Plan       `json:"plan"`
	}{"goo.cel-plan/1", querycel.LibraryVersion, *expression, program.Predicate(), question, plan}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
