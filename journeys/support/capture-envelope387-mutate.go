//go:build ignore

// Journey-only corruption fixture. Decode valid captures, then mutate typed Go
// values and marshal them to use Go's exact field ordering/map ordering/escaping.
// It intentionally bypasses admission while producing invalid isolated files.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/casselc/goo/exploration"
)

var arms = map[string]string{"explore": "Explore", "inventory": "Inventory", "comparison": "Comparison", "measurement": "Measurement", "explanation": "Explanation", "trace-search": "TraceSearch"}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func read(file string) exploration.SavedCapture {
	raw, e := os.ReadFile(file)
	must(e)
	var c exploration.SavedCapture
	must(json.Unmarshal(raw, &c))
	return c
}
func insertNull(raw []byte, key string) []byte {
	if len(raw) < 2 || raw[len(raw)-1] != '}' {
		panic("object required")
	}
	suffix, e := json.Marshal(key)
	must(e)
	out := append([]byte{}, raw[:len(raw)-1]...)
	out = append(out, ',')
	out = append(out, suffix...)
	return append(out, ":null}"...)
}
func questionNull(raw []byte, key string) []byte {
	d := json.NewDecoder(bytes.NewReader(raw))
	_, e := d.Token()
	must(e)
	for d.More() {
		token, e := d.Token()
		must(e)
		var value json.RawMessage
		must(d.Decode(&value))
		end := int(d.InputOffset())
		if token == "question" {
			start := end - len(value)
			return append(append(append([]byte{}, raw[:start]...), insertNull(value, key)...), raw[end:]...)
		}
	}
	panic("question field required")
}
func main() {
	if len(os.Args) != 5 {
		panic("usage: capture-envelope387-mutate SOURCE.json FOREIGN.json MODE OUTPUT_DIR")
	}
	c, donor := read(os.Args[1]), read(os.Args[2])
	mode := os.Args[3]
	originalID := c.ID
	foreign := donor.Question.Kind
	if foreign == c.Question.Kind {
		panic("different donor family required")
	}
	foreignJSON := foreign
	if foreign == "trace-search" {
		foreignJSON = "traceSearch"
	}
	field := arms[foreign]
	if field == "" {
		panic("unsupported donor")
	}
	c.Title = "387 negative " + mode + " " + foreign + " " + originalID[:8]
	switch mode {
	case "payload-mixed":
		reflect.ValueOf(&c).Elem().FieldByName(field).Set(reflect.ValueOf(donor).FieldByName(field))
	case "question-mixed":
		reflect.ValueOf(&c.Question).Elem().FieldByName(field).Set(reflect.ValueOf(donor.Question).FieldByName(field))
	case "payload-null", "question-null": // Marshal a valid typed body first; explicit null presence is injected after hashing.
	case "scope-corrupt":
		switch c.Question.Kind {
		case "explore":
			c.Explore.Question.FromMS++
		case "inventory":
			c.Inventory.Question.FromMS++
		case "comparison":
			c.Comparison.Request.Selection.Question.FromMS++
		case "measurement":
			c.Measurement.Question.FromMS++
		case "explanation":
			c.Explanation.Request.Selection.FromMS++
		case "trace-search":
			c.TraceSearch.Request.FromMS++
		default:
			panic("unsupported source")
		}
	default:
		panic("unsupported mutation")
	}
	c.ID = ""
	canonical, e := json.Marshal(c)
	must(e)
	sum := sha256.Sum256(canonical)
	c.ID = hex.EncodeToString(sum[:])
	raw, e := json.Marshal(c)
	must(e)
	switch mode {
	case "payload-null":
		raw = insertNull(raw, foreignJSON)
	case "question-null":
		raw = questionNull(raw, foreignJSON)
	}
	output, e := filepath.Abs(os.Args[4])
	must(e)
	permitted := regexp.MustCompile("^" + regexp.QuoteMeta(filepath.Join(os.TempDir(), "goo-generic-191-")) + "[^/]+/captures/exploration$")
	if !permitted.MatchString(output) {
		panic("only an isolated generic-runtime capture directory is permitted")
	}
	actual, e := filepath.EvalSymlinks(output)
	must(e)
	if actual != output {
		panic("symlinked capture directory is not permitted")
	}
	file := filepath.Join(output, c.ID+".json")
	must(os.WriteFile(file, raw, 0600))
	must(json.NewEncoder(os.Stdout).Encode(map[string]any{"sourceID": originalID, "id": c.ID, "family": c.Question.Kind, "foreignFamily": foreign, "foreignField": foreignJSON, "mode": mode, "path": file, "canonicalSHA256": c.ID, "rawBytes": len(raw), "nullIgnoredDigestWouldMatch": strings.HasSuffix(mode, "-null")}))
}
