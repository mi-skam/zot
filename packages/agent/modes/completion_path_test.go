package modes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathCompletionProviderOpensOnAmbiguousDirectoryToken(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "examples", "echo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "examples", "eval"), 0o755); err != nil {
		t.Fatal(err)
	}
	p := newPathCompletionProvider()
	if !p.Open(completionContext{Input: "update ./examples/e", CWD: tmp}) {
		t.Fatal("expected path popup to open")
	}
	res := p.Suggestions(completionContext{Input: "update ./examples/e", CWD: tmp})
	if len(res.Items) != 2 {
		t.Fatalf("items = %#v", res.Items)
	}
	got, ok := p.Apply(completionContext{Input: "update ./examples/e", CWD: tmp}, res, res.Items[0])
	if !ok || got != "update ./examples/echo/" {
		t.Fatalf("Apply = %q ok=%v", got, ok)
	}
}

func TestPathCompletionProviderOpensOnDotSlash(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "examples"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(tmp, "README.md"))
	p := newPathCompletionProvider()
	if !p.Open(completionContext{Input: "./", CWD: tmp}) {
		t.Fatal("expected ./ popup to open")
	}
	res := p.Suggestions(completionContext{Input: "./", CWD: tmp})
	if len(res.Items) != 2 {
		t.Fatalf("items = %#v", res.Items)
	}
}
