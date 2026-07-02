package modes

import "testing"

func TestCompletionSlashProviderPreservesExtensionsHeader(t *testing.T) {
	s := newSlashSuggester()
	s.SetExtra([]slashCommand{{Name: "/jira", Desc: "Jira"}})
	p := newSlashCompletionProvider(s)
	res := p.Suggestions(completionContext{Input: "/"})
	var sawHeader, sawJira bool
	for _, it := range res.Items {
		if it.Header && it.Label == "extensions" {
			sawHeader = true
		}
		if it.Value == "/jira" && it.Description == "Jira" {
			sawJira = true
		}
	}
	if !sawHeader || !sawJira {
		t.Fatalf("items missing extension header/jira: %#v", res.Items)
	}
}

func TestCompletionSlashApplyAndCancel(t *testing.T) {
	p := newSlashCompletionProvider(newSlashSuggester())
	res := p.Suggestions(completionContext{Input: "/he"})
	if len(res.Items) == 0 {
		t.Fatal("expected slash items")
	}
	got, ok := p.Apply(completionContext{}, res, res.Items[0])
	if !ok || got != "/help" {
		t.Fatalf("Apply = %q ok=%v", got, ok)
	}
	got, ok = p.Cancel(completionContext{}, res)
	if !ok || got != "" {
		t.Fatalf("Cancel = %q ok=%v", got, ok)
	}
}
