package modes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patriceckhart/zot/packages/tui"
)

type stubCompletionProvider struct {
	name          string
	provider      string
	items         []completionItem
	renderRows    []string
	browserCursor *int
	applied       bool
	cancelled     bool
	applyValue    string
}

func (p *stubCompletionProvider) Name() string { return p.name }
func (p *stubCompletionProvider) Suggestions(ctx completionContext) completionResult {
	provider := p.provider
	if provider == "" {
		provider = p.name
	}
	return completionResult{Provider: provider, Items: p.items}
}
func (p *stubCompletionProvider) Apply(ctx completionContext, result completionResult, item completionItem) (string, bool) {
	p.applied = true
	if p.applyValue != "" {
		return p.applyValue, true
	}
	return item.Value, true
}
func (p *stubCompletionProvider) Cancel(ctx completionContext, result completionResult) (string, bool) {
	p.cancelled = true
	return "cancelled", true
}
func (p *stubCompletionProvider) setCompletionCursor(cursor int) {
	if p.browserCursor != nil {
		*p.browserCursor = cursor
	}
}
func (p *stubCompletionProvider) RenderCompletion(ctx completionContext, th tui.Theme, width int) []string {
	return p.renderRows
}

func TestCompletionManagerPriority(t *testing.T) {
	low := &stubCompletionProvider{name: "low", items: []completionItem{{Value: "low", Label: "low"}}}
	high := &stubCompletionProvider{name: "high", items: []completionItem{{Value: "high", Label: "high"}}}
	m := newCompletionManager(high, low)
	if !m.Refresh(completionContext{}) {
		t.Fatal("expected active completion")
	}
	if got := m.ActiveProvider(); got != "high" {
		t.Fatalf("active provider = %q, want high", got)
	}
}

func TestCompletionManagerIgnoresInactiveProviders(t *testing.T) {
	empty := &stubCompletionProvider{name: "empty"}
	full := &stubCompletionProvider{name: "full", items: []completionItem{{Value: "x", Label: "x"}}}
	m := newCompletionManager(empty, full)
	m.Refresh(completionContext{})
	if got := m.ActiveProvider(); got != "full" {
		t.Fatalf("active provider = %q, want full", got)
	}
}

func TestCompletionManagerNavigationSkipsHeaders(t *testing.T) {
	p := &stubCompletionProvider{name: "p", items: []completionItem{
		{Value: "a", Label: "a"},
		{Label: "group", Header: true},
		{Value: "b", Label: "b"},
	}}
	m := newCompletionManager(p)
	m.Refresh(completionContext{})
	m.Down()
	item, ok := m.Selected()
	if !ok || item.Value != "b" {
		t.Fatalf("selected = %#v ok=%v, want b", item, ok)
	}
	m.Up()
	item, ok = m.Selected()
	if !ok || item.Value != "a" {
		t.Fatalf("selected = %#v ok=%v, want a", item, ok)
	}
}

func TestCompletionManagerApplyCancelReset(t *testing.T) {
	p := &stubCompletionProvider{name: "p", items: []completionItem{{Value: "x", Label: "x"}}}
	m := newCompletionManager(p)
	m.Refresh(completionContext{})
	if got, ok := m.Apply(completionContext{}); !ok || got != "x" || !p.applied {
		t.Fatalf("Apply = %q ok=%v applied=%v", got, ok, p.applied)
	}
	if got, ok := m.Cancel(completionContext{}); !ok || got != "cancelled" || !p.cancelled {
		t.Fatalf("Cancel = %q ok=%v cancelled=%v", got, ok, p.cancelled)
	}
	m.Reset()
	if m.Active() || m.cursor != 0 || len(m.last.Items) != 0 {
		t.Fatalf("Reset left state: %#v", m)
	}
}

func TestCompletionManagerBrowseResetsCursor(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "b_dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(tmp, "a.txt"))
	mustWriteFile(t, filepath.Join(tmp, "c.txt"))
	mustWriteFile(t, filepath.Join(tmp, "b_dir", "inside.txt"))

	fs := newFileSuggester()
	p := newFileCompletionProvider(fs)
	m := newCompletionManager(p)
	ctx := completionContext{Input: "@", CWD: tmp}
	if !m.Refresh(ctx) {
		t.Fatal("expected active file completion")
	}
	for item, ok := m.Selected(); !ok || !item.IsDir; item, ok = m.Selected() {
		m.Down()
		if m.cursor == len(m.last.Items)-1 {
			break
		}
	}
	if item, ok := m.Selected(); !ok || !item.IsDir {
		t.Fatalf("selected before browse = %#v ok=%v, want directory", item, ok)
	}
	if !m.BrowseRight() {
		t.Fatal("BrowseRight should enter selected directory")
	}
	if m.cursor != 0 || fs.cursor != 0 {
		t.Fatalf("cursor after browse = manager %d suggester %d, want both 0", m.cursor, fs.cursor)
	}
	if !m.Refresh(ctx) {
		t.Fatal("expected active completion inside directory")
	}
	item, ok := m.Selected()
	if !ok || item.Value != filepath.Join("b_dir", "inside.txt") {
		t.Fatalf("selected after browse = %#v ok=%v", item, ok)
	}
}
