package modes

import (
	"strings"

	"github.com/patriceckhart/zot/packages/tui"
)

const fileCompletionProviderName = "file"

type fileCompletionProvider struct{ s *fileSuggester }

func newFileCompletionProvider(s *fileSuggester) *fileCompletionProvider {
	return &fileCompletionProvider{s: s}
}

func (p *fileCompletionProvider) Name() string                   { return fileCompletionProviderName }
func (p *fileCompletionProvider) setCompletionCursor(cursor int) { p.s.cursor = cursor }

func (p *fileCompletionProvider) Suggestions(ctx completionContext) completionResult {
	p.s.SetCWD(ctx.CWD)
	m := p.s.matches(ctx.Input)
	p.s.lastMatches = m
	items := make([]completionItem, 0, len(m))
	for _, e := range m {
		label := e.name
		if e.isDir {
			label += "/"
		}
		items = append(items, completionItem{Value: e.rel, Label: label, IsDir: e.isDir})
	}
	return completionResult{Provider: p.Name(), Items: items}
}

func (p *fileCompletionProvider) Apply(ctx completionContext, result completionResult, item completionItem) (string, bool) {
	chip := "[file:" + item.Value + "]"
	if item.IsDir {
		chip = "[dir:" + item.Value + "/]"
	}
	val := ctx.Input
	if idx := strings.LastIndex(val, "@"); idx >= 0 {
		val = val[:idx]
	}
	out := val + chip + " "
	p.s.Reset()
	return out, true
}

func (p *fileCompletionProvider) Cancel(ctx completionContext, result completionResult) (string, bool) {
	val := ctx.Input
	if idx := strings.LastIndex(val, "@"); idx >= 0 {
		val = val[:idx]
	}
	p.s.Reset()
	return val, true
}

func (p *fileCompletionProvider) RenderCompletion(ctx completionContext, th tui.Theme, width int) []string {
	p.s.SetCWD(ctx.CWD)
	return p.s.Render(ctx.Input, th, width)
}

func (p *fileCompletionProvider) BrowseRight() bool { return p.s.Right() }
func (p *fileCompletionProvider) BrowseLeft() bool  { return p.s.Left() }
