package nativecatalog

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
)

func definition(i int) measurements.Definition {
	return measurements.Definition{APIVersion: measurements.Version, Name: fmt.Sprintf("native_%d", i), Title: "Native", Kind: "gauge", Environment: "test", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"app"}, Scope: "sdk", Metric: "depth", Unit: "1"}}
}
func seed(t *testing.T, s Store, n int) {
	t.Helper()
	if e := os.MkdirAll(s.SavedDir, 0700); e != nil {
		t.Fatal(e)
	}
	for i := range n {
		d := definition(i)
		b, _ := yaml.Marshal(d)
		if e := os.WriteFile(filepath.Join(s.SavedDir, measurements.Hash(d)+".yaml"), b, 0600); e != nil {
			t.Fatal(e)
		}
	}
}
func TestConcurrentAdmissionAndOverLimitResolution(t *testing.T) {
	s := Store{BuiltinDir: filepath.Join(t.TempDir(), "absent"), SavedDir: filepath.Join(t.TempDir(), "saved")}
	seed(t, s, Limit-1)
	start := make(chan struct{})
	results := make(chan error, 16)
	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			independent := Store{BuiltinDir: s.BuiltinDir, SavedDir: s.SavedDir}
			results <- independent.Publish(t.Context(), definition(1000+i))
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	admitted := 0
	for e := range results {
		if e == nil {
			admitted++
		} else if !errors.Is(e, ErrCapacity) {
			t.Fatal(e)
		}
	}
	if admitted != 1 {
		t.Fatal("concurrent admission exceeded capacity", admitted)
	}
	entries, e := s.List(t.Context())
	if e != nil || len(entries) != Limit {
		t.Fatal(len(entries), e)
	}
	if e = s.Publish(t.Context(), definition(0)); e != nil {
		t.Fatal("duplicate at capacity refused", e)
	}
	// An external legacy writer created an already over-limit store. Never remove
	// it; exact saved IDs and verified duplicate retry remain available.
	d := definition(2000)
	b, _ := yaml.Marshal(d)
	file := filepath.Join(s.SavedDir, measurements.Hash(d)+".yaml")
	os.WriteFile(file, b, 0600)
	if _, e = s.List(t.Context()); !errors.Is(e, ErrCapacity) {
		t.Fatal(e)
	}
	if got, e := s.Resolve(t.Context(), measurements.Hash(d)); e != nil || got.Revision != measurements.Hash(d) {
		t.Fatal(got, e)
	}
	if e = s.Publish(t.Context(), d); e != nil {
		t.Fatal(e)
	}
	if e = s.Publish(t.Context(), definition(3000)); !errors.Is(e, ErrCapacity) {
		t.Fatal(e)
	}
	os.WriteFile(file, []byte("corrupt"), 0600)
	if e = s.Publish(t.Context(), d); e == nil {
		t.Fatal("corrupt duplicate accepted")
	}
	if _, e = s.Resolve(t.Context(), measurements.Hash(d)); e == nil {
		t.Fatal("corrupt exact lookup accepted")
	}
	got, _ := os.ReadFile(file)
	if string(got) != "corrupt" {
		t.Fatal("corrupt winner changed")
	}
}
func TestBuiltinAliasesPhysicalBudgetAndReadOnlyMissing(t *testing.T) {
	root := t.TempDir()
	s := Store{BuiltinDir: filepath.Join(root, "builtin"), SavedDir: filepath.Join(root, "saved")}
	os.Mkdir(s.BuiltinDir, 0700)
	d := definition(0)
	b, _ := yaml.Marshal(d)
	os.WriteFile(filepath.Join(s.BuiltinDir, "builtin.yaml"), b, 0600)
	entries, e := s.List(t.Context())
	if e != nil || len(entries) != 1 || !entries[0].Builtin {
		t.Fatal(entries, e)
	}
	if _, e = os.Stat(s.SavedDir); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("GET created saved directory", e)
	}
	if e = s.Publish(t.Context(), d); e != nil {
		t.Fatal(e)
	}
	entries, e = s.List(t.Context())
	if e != nil || len(entries) != 1 || !entries[0].Builtin {
		t.Fatal("alias not deduplicated", entries, e)
	}
	for i := range Limit - 1 {
		if e = os.WriteFile(filepath.Join(s.SavedDir, fmt.Sprint("alias-", i, ".yaml")), b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	entries, e = s.List(t.Context())
	if e != nil || len(entries) != 1 {
		t.Fatal(entries, e)
	}
	if e = s.Publish(t.Context(), definition(1)); !errors.Is(e, ErrCapacity) {
		t.Fatal("physical aliases bypassed admission", e)
	}
	if e = s.Publish(t.Context(), d); e != nil {
		t.Fatal(e)
	}
}
func TestDirectoryLockCancellation(t *testing.T) {
	path := t.TempDir()
	release, e := lockDirectory(t.Context(), path, true)
	if e != nil {
		t.Fatal(e)
	}
	defer release()
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	if _, e := lockDirectory(ctx, path, false); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal("lock wait ignored cancellation", e)
	}
}

func TestBuiltinUnionAdmission(t *testing.T) {
	root := t.TempDir()
	s := Store{BuiltinDir: filepath.Join(root, "builtin"), SavedDir: filepath.Join(root, "saved")}
	seed(t, s, Limit-1)
	os.Mkdir(s.BuiltinDir, 0700)
	d := definition(1000)
	b, _ := yaml.Marshal(d)
	os.WriteFile(filepath.Join(s.BuiltinDir, "builtin.yaml"), b, 0600)
	if e := s.Publish(t.Context(), definition(2000)); !errors.Is(e, ErrCapacity) {
		t.Fatal("combined logical capacity bypassed", e)
	}
	if e := s.Publish(t.Context(), d); e != nil {
		t.Fatal("existing builtin could not be preserved", e)
	}
	all, e := s.List(t.Context())
	if e != nil || len(all) != Limit {
		t.Fatal(len(all), e)
	}
}

func TestNativeCatalogProcessHelper(t *testing.T) {
	dir := os.Getenv("GOO_NATIVE_CATALOG_TEST_DIR")
	if dir == "" {
		return
	}
	i, e := strconv.Atoi(os.Getenv("GOO_NATIVE_CATALOG_TEST_INDEX"))
	if e != nil {
		os.Exit(2)
	}
	s := Store{SavedDir: dir, BuiltinDir: filepath.Join(dir, "absent-builtin")}
	e = s.Publish(context.Background(), definition(i))
	if e == nil {
		os.Exit(0)
	}
	if errors.Is(e, ErrCapacity) {
		os.Exit(3)
	}
	fmt.Fprintln(os.Stderr, e)
	os.Exit(2)
}
func TestNativeCatalogCrossProcessAdmission(t *testing.T) {
	s := Store{SavedDir: t.TempDir(), BuiltinDir: filepath.Join(t.TempDir(), "absent")}
	seed(t, s, Limit-1)
	release, e := lockDirectory(t.Context(), s.SavedDir, true)
	if e != nil {
		t.Fatal(e)
	}
	defer func() { release() }()
	children := []*exec.Cmd{}
	for i := range 2 {
		child := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestNativeCatalogProcessHelper$")
		child.Env = append(os.Environ(), "GOO_NATIVE_CATALOG_TEST_DIR="+s.SavedDir, fmt.Sprint("GOO_NATIVE_CATALOG_TEST_INDEX=", 1000+i))
		if e = child.Start(); e != nil {
			t.Fatal(e)
		}
		children = append(children, child)
	}
	release()
	release = func() {}
	admitted, full := 0, 0
	for _, child := range children {
		e := child.Wait()
		if e == nil {
			admitted++
		} else if child.ProcessState.ExitCode() == 3 {
			full++
		} else {
			t.Fatal(e)
		}
	}
	if admitted != 1 || full != 1 {
		t.Fatal("cross-process capacity exceeded", admitted, full)
	}
	all, e := s.List(t.Context())
	if e != nil || len(all) != Limit {
		t.Fatal(len(all), e)
	}
}

func TestCanonicalSizeAndCorruptBuiltinWinner(t *testing.T) {
	root := t.TempDir()
	s := Store{BuiltinDir: filepath.Join(root, "builtin"), SavedDir: filepath.Join(root, "saved")}
	d := definition(0)
	d.Source.Attributes = map[string]string{}
	for i := range 16 {
		d.Source.Attributes[fmt.Sprint("key", i)] = strings.Repeat("\x01", 512)
	}
	if e := d.Validate(); e != nil {
		t.Fatal(e)
	}
	b, e := yaml.Marshal(d)
	if e != nil || len(b) <= FileBytes {
		t.Fatal("size fixture does not exceed reader limit", len(b), e)
	}
	if e = s.Publish(t.Context(), d); e == nil {
		t.Fatal("accepted unbrowseable YAML")
	}
	files, e := os.ReadDir(s.SavedDir)
	if e != nil || len(files) != 0 {
		t.Fatal("rejection published bytes", files, e)
	}
	d = definition(1)
	b, _ = yaml.Marshal(d)
	os.Mkdir(s.BuiltinDir, 0700)
	os.WriteFile(filepath.Join(s.BuiltinDir, "builtin.yaml"), b, 0600)
	file := filepath.Join(s.SavedDir, measurements.Hash(d)+".yaml")
	os.WriteFile(file, []byte("corrupt"), 0600)
	if _, e = s.Resolve(t.Context(), measurements.Hash(d)); e == nil {
		t.Fatal("builtin hid corrupt saved winner")
	}
	if e = s.Publish(t.Context(), d); e == nil {
		t.Fatal("corrupt builtin duplicate accepted")
	}
}
