package workbench

import (
	"testing"
)

func TestViewContentIDMatchesExistingCatalogIdentity(t *testing.T) {
	view := compositionView()
	// digest is the pre-existing catalog recipe; identity must not incorporate
	// the saved ID or portable descriptive labels.
	want, err := digest(view)
	if err != nil {
		t.Fatal(err)
	}
	id, err := ViewContentID(view)
	if err != nil || id != want {
		t.Fatal("content identity changed", id, want, err)
	}
	store := Store{Dir: t.TempDir()}
	saved, err := store.CreateExplorationView(view)
	if err != nil || saved.ID != id {
		t.Fatal("save identity differs", saved, err)
	}
	loaded, err := store.LoadExplorationView(id)
	if err != nil || loaded.ID != id {
		t.Fatal("load integrity changed", loaded, err)
	}
	if _, err = ViewContentID(*loaded); err == nil {
		t.Fatal("saved ID admitted into content hash")
	}
	view.Title = "A different fixed view"
	changed, err := ViewContentID(view)
	if err != nil || changed == id {
		t.Fatal("content change did not change identity", err)
	}
}
