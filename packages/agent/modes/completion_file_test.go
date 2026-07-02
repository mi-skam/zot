package modes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompletionFileApplyInsertsChip(t *testing.T) {
	tmp := t.TempDir()
	mustWriteFile(t, filepath.Join(tmp, "README.md"))
	fs := newFileSuggester()
	p := newFileCompletionProvider(fs)
	ctx := completionContext{Input: "read @read", CWD: tmp}
	res := p.Suggestions(ctx)
	if len(res.Items) == 0 {
		t.Fatal("expected file completion items")
	}
	got, ok := p.Apply(ctx, res, res.Items[0])
	if !ok || got != "read [file:README.md] " {
		t.Fatalf("Apply = %q ok=%v", got, ok)
	}
}

func TestCompletionFileCancelRemovesQuery(t *testing.T) {
	p := newFileCompletionProvider(newFileSuggester())
	got, ok := p.Cancel(completionContext{Input: "read @foo"}, completionResult{})
	if !ok || got != "read " {
		t.Fatalf("Cancel = %q ok=%v", got, ok)
	}
}

func TestCompletionFileRightLeftFlatOnly(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	fs := newFileSuggester()
	fs.SetCWD(tmp)
	p := newFileCompletionProvider(fs)
	_ = p.Suggestions(completionContext{Input: "@dir", CWD: tmp})
	fs.lastMatches = fs.matches("@dir")
	if !p.BrowseRight() {
		t.Fatal("Right should enter dir in flat mode")
	}
	if !p.BrowseLeft() {
		t.Fatal("Left should leave dir in flat mode")
	}
	fs.SetRecursive(true)
	_ = p.Suggestions(completionContext{Input: "@dir", CWD: tmp})
	fs.lastMatches = fs.matches("@dir")
	if p.BrowseRight() {
		t.Fatal("Right should be disabled in recursive mode")
	}
	if p.BrowseLeft() {
		t.Fatal("Left should be disabled in recursive mode")
	}
}
