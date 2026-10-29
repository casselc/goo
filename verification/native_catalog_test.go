package verification

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/goo/internal/nativecatalog"
	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
	"hegel.dev/go/hegel"
)

type nativeCatalogMachine struct {
	store    nativecatalog.Store
	base     measurements.Definition
	expected map[string]bool
}

func (m *nativeCatalogMachine) definition(i int) measurements.Definition {
	d := m.base
	d.Name = fmt.Sprintf("native_%d", i)
	return d
}
func (m *nativeCatalogMachine) RulePublish(tc hegel.TestCase) {
	d := m.definition(hegel.Draw(tc, hegel.Integers(0, 260)))
	id := measurements.Hash(d)
	known := m.expected[id]
	e := m.store.Publish(context.Background(), d)
	if known || len(m.expected) < nativecatalog.Limit {
		if e != nil {
			panic(e)
		}
		m.expected[id] = true
	} else if !errors.Is(e, nativecatalog.ErrCapacity) {
		panic("full store admitted distinct revision")
	}
}
func (m *nativeCatalogMachine) RuleResolve(tc hegel.TestCase) {
	d := m.definition(hegel.Draw(tc, hegel.Integers(0, 253)))
	id := measurements.Hash(d)
	got, e := m.store.Resolve(context.Background(), id)
	if e != nil || got.Revision != id || measurements.Hash(got.Definition) != id {
		panic("known revision changed or disappeared")
	}
}
func (m *nativeCatalogMachine) InvariantCatalog(tc hegel.TestCase) {
	all, e := m.store.List(context.Background())
	if e != nil || len(all) != len(m.expected) {
		panic("accepted publication disabled or changed catalog")
	}
	for _, entry := range all {
		if !m.expected[entry.Revision] {
			panic("catalog invented revision")
		}
	}
}
func TestHegelNativeCatalogStateSwarm(t *testing.T) {
	raw, e := os.ReadFile("../config/measurements/worker-active.yaml")
	if e != nil {
		t.Fatal(e)
	}
	base, e := measurements.Read(string(raw))
	if e != nil {
		t.Fatal(e)
	}
	hegel.Test(t, func(tc *hegel.T) {
		root := tc.TempDir()
		saved := filepath.Join(root, "saved")
		if e := os.Mkdir(saved, 0700); e != nil {
			tc.Fatal(e)
		}
		m := &nativeCatalogMachine{store: nativecatalog.Store{BuiltinDir: filepath.Join(root, "missing"), SavedDir: saved}, base: base, expected: map[string]bool{}}
		for i := range 254 {
			d := m.definition(i)
			id := measurements.Hash(d)
			b, _ := yaml.Marshal(d)
			if e := os.WriteFile(filepath.Join(saved, id+".yaml"), b, 0600); e != nil {
				tc.Fatal(e)
			}
			m.expected[id] = true
		}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(10), hegel.WithAlwaysCheckInvariants("InvariantCatalog"))
	}, options(20)...)
}
