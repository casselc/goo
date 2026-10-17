package annotations

import (
	"os"
	"path/filepath"
	"testing"
)

func scanText(t *testing.T, source string) (*Manifest, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "example.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return ScanDir(dir)
}
func TestManifestDeterminismAndUnits(t *testing.T) {
	source := `package checkout
import "context"
//goo:operation name:checkout.reserve role:server
func Reserve(ctx context.Context) (err error){return nil}
//goo:event name:reservation.accepted subject:id when:ok
func Accept(ctx context.Context,id string) (err error){return nil}
`
	a, err := scanText(t, source)
	if err != nil {
		t.Fatal(err)
	}
	b, err := scanText(t, source)
	if err != nil {
		t.Fatal(err)
	}
	if a.Revision != b.Revision || a.Declarations[0].Unit != "subject-event" || len(a.Signals) != 4 {
		t.Fatal("invalid contract")
	}
	for _, signal := range a.Signals {
		for _, dimension := range signal.Dimensions {
			if dimension == "goo.subject.id" {
				t.Fatal("high cardinality metric dimension")
			}
		}
	}
}
func TestRejectsUnsafeAndUnsupportedAnnotations(t *testing.T) {
	for _, source := range []string{
		`//goo:operation name:checkout role:server surprise:x
func F(ctx context.Context)(err error){return nil}`,
		`//goo:event name:accepted subject:lookup() when:ok
func F(ctx context.Context)(err error){return nil}`,
		`//goo:event name:accepted subject:missing.ID when:ok
func F(ctx context.Context)(err error){return nil}`,
		`//goo:event name:accepted subject:id when:always
func F(ctx context.Context,id string)(err error){return nil}`,
		`//goo:operation name:checkout role:server
func F(ctx context.Context)error{return nil}`,
		`//goo:operation name:checkout role:server
func F(){}`,
		`//goo:operation name:checkout role:server
func F(ctx context.Context, other context.Context){}`,
		`//goo:event name:accepted subject:id when:ok
//goo:operation name:checkout role:server
func F(ctx context.Context,id string)(err error){return nil}`,
	} {
		if _, err := scanText(t, "package p\nimport \"context\"\n"+source); err == nil {
			t.Fatal("accepted unsupported declaration", source)
		}
	}
}
