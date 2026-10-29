package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/casselc/goo/exploration"
	"io"
	"os"
	"strings"

	"github.com/casselc/goo/lifecycle"
	"github.com/casselc/goo/prebuilt"
	"github.com/casselc/goo/subjectfacts"
)

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: goo validate|compile|evaluate|rebuild-capture-catalog|prepare|evaluate-prepared|discover|catalog|expand|prepare-http|manifest|weave|observe-operations|plan-operations|evaluate-counters|evaluate-histograms|merge-histograms [options]")
	}
	command := args[0]
	if command == "rebuild-capture-catalog" {
		fs := flag.NewFlagSet(command, flag.ContinueOnError)
		dir := fs.String("dir", "", "immutable exploration capture directory")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *dir == "" || fs.NArg() != 0 {
			return fmt.Errorf("rebuild-capture-catalog requires -dir and no positional arguments")
		}
		result, err := (exploration.CaptureStore{Dir: *dir}).RebuildCatalog(context.Background())
		if writeErr := json.NewEncoder(out).Encode(result); writeErr != nil {
			return writeErr
		}
		return err
	}
	if command == "plan-exploration" {
		return planExploration(args[1:], out)
	}
	if command == "evaluate-histograms" || command == "merge-histograms" {
		return histograms(args[1:], out, command == "merge-histograms")
	}
	if command == "evaluate-counters" {
		return counters(args[1:], out)
	}
	if command == "plan-operations" {
		return planOperations(args[1:], out)
	}
	if command == "observe-operations" {
		return observe(args[1:], out)
	}
	if command == "weave" {
		return weave(args[1:], out)
	}
	if command == "manifest" {
		return manifest(args[1:], out)
	}
	if command == "serve-http" {
		return serveHTTP(args[1:])
	}
	if command == "expand" || command == "prepare-http" {
		return expandPackage(command, args[1:], out)
	}
	if command == "catalog" {
		if len(args) != 1 {
			return fmt.Errorf("catalog takes no arguments")
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(prebuilt.Catalog())
	}
	if command == "discover" {
		fs := flag.NewFlagSet("discover", flag.ContinueOnError)
		inventoryPath := fs.String("inventory", "", "versioned telemetry shape inventory JSON")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 || *inventoryPath == "" {
			return fmt.Errorf("discover requires -inventory, without positional arguments")
		}
		f, err := os.Open(*inventoryPath)
		if err != nil {
			return err
		}
		defer f.Close()
		in, err := prebuilt.ReadInventory(f)
		if err != nil {
			return err
		}
		result, err := prebuilt.JSON(in)
		if err != nil {
			return err
		}
		_, err = out.Write(result)
		return err
	}
	if command == "prepare" {
		return prepare(args[1:], out)
	}
	if command != "validate" && command != "compile" && command != "evaluate" && command != "evaluate-prepared" {
		return fmt.Errorf("unknown command %q", command)
	}
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	definition := fs.String("definition", "", "YAML measurement definition")
	target := fs.String("target", "", "compile target: clickhouse or databricks")
	factsPath := fs.String("facts", "", "JSON array of prepared subject facts")
	preparedPath := fs.String("prepared", "", "complete lifecycle preparation artifact")
	evaluation := fs.String("evaluation", "", "one evaluation ID for reference execution")
	groupBy := fs.String("group-by", "", "comma-separated dimension names; empty means total")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 0 || *definition == "" {
		return fmt.Errorf("require -definition and no positional arguments")
	}
	f, err := os.Open(*definition)
	if err != nil {
		return err
	}
	defer f.Close()
	d, err := subjectfacts.Read(f)
	if err != nil {
		return err
	}
	groups := []string{}
	if *groupBy != "" {
		for _, g := range strings.Split(*groupBy, ",") {
			groups = append(groups, strings.TrimSpace(g))
		}
	}
	var result any
	switch command {
	case "validate":
		result = d
	case "compile":
		result, err = d.Compile(*target, groups)
	case "evaluate":
		var b []byte
		b, err = os.ReadFile(*factsPath)
		if err == nil {
			var facts []subjectfacts.Fact
			if err = json.Unmarshal(b, &facts); err == nil {
				result, err = d.Evaluate(facts, *evaluation, groups)
			}
		}
	case "evaluate-prepared":
		if *evaluation != "" {
			return fmt.Errorf("evaluate-prepared takes its evaluation ID from the receipt; it cannot be overridden")
		}
		var b []byte
		b, err = os.ReadFile(*preparedPath)
		if err == nil {
			var p lifecycle.Prepared
			if err = json.Unmarshal(b, &p); err == nil {
				result, err = p.Evaluate(d, groups)
			}
		}
	}
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func prepare(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("prepare", flag.ContinueOnError)
	policy := fs.String("policy", "", "lifecycle policy and frozen evaluation context")
	events := fs.String("events", "", "JSON array of normalized lifecycle events")
	previous := fs.String("previous", "", "prior preparation artifact when superseding an evaluation")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 || *policy == "" || *events == "" {
		return fmt.Errorf("prepare requires -policy and -events, without positional arguments")
	}
	f, err := os.Open(*policy)
	if err != nil {
		return err
	}
	defer f.Close()
	p, err := lifecycle.ReadPolicy(f)
	if err != nil {
		return err
	}
	if p.SupersedesEvaluationID != "" || *previous != "" {
		priorBytes, err := os.ReadFile(*previous)
		if err != nil {
			return fmt.Errorf("correction requires -previous: %w", err)
		}
		var prior lifecycle.Prepared
		if err := json.Unmarshal(priorBytes, &prior); err != nil {
			return err
		}
		if err := p.ValidateCorrection(prior.Receipt); err != nil {
			return err
		}
	}
	b, err := os.ReadFile(*events)
	if err != nil {
		return err
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	result, err := p.PrepareRaw(raw)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "goo:", err)
		os.Exit(1)
	}
}
