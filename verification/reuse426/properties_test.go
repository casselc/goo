package reuse426

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"hegel.dev/go/hegel"
)

type sample struct {
	Kind, Signal, Policy, Environment, Tenant string
	From                                      int64
	Success                                   bool
}

func TestGeneratedCustodyAuthorityTransitions(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 65535))
		samples := []sample{}
		for i := range 8 {
			kind, signal, policy := "attribute", "logs", "effective"
			if i%2 == 1 {
				kind = "resource"
				signal = "spans"
			}
			if i%3 == 0 {
				policy = "current-only"
			}
			samples = append(samples, sample{kind, signal, policy, fmt.Sprintf("prod-%d", seed), fmt.Sprintf("tenant-雪-%d", seed), int64(1000 + seed*1000 + i*60000), (seed+i)%3 != 0})
		}
		raw, err := json.Marshal(samples)
		if err != nil {
			tc.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "node", "bridge.mjs")
		cmd.Stdin = bytes.NewReader(raw)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			tc.Fatalf("production counted custody transitions: %v %s", err, stderr.String())
		}
		var count int
		if err = json.Unmarshal(out, &count); err != nil || count != len(samples) {
			tc.Fatalf("bridge lost transitions: %s %v", out, err)
		}
	}, hegel.WithTestCases(24), hegel.WithSeed(426), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
