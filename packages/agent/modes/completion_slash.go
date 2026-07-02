package modes

import "github.com/patriceckhart/zot/packages/tui"

const slashCompletionProviderName = "slash"

type slashCompletionProvider struct{ s *slashSuggester }

func newSlashCompletionProvider(s *slashSuggester) *slashCompletionProvider {
	return &slashCompletionProvider{s: s}
}

func (p *slashCompletionProvider) Name() string                   { return slashCompletionProviderName }
func (p *slashCompletionProvider) setCompletionCursor(cursor int) { p.s.cursor = cursor }

func (p *slashCompletionProvider) Suggestions(ctx completionContext) completionResult {
	m := p.s.matches(ctx.Input)
	items := make([]completionItem, 0, len(m))
	for _, c := range m {
		items = append(items, completionItem{Value: c.Name, Label: c.Name, Description: c.Desc, Header: c.Header})
	}
	return completionResult{Provider: p.Name(), Items: items}
}

func (p *slashCompletionProvider) Apply(ctx completionContext, result completionResult, item completionItem) (string, bool) {
	if item.Value == "" || item.Header {
		return "", false
	}
	p.s.Reset()
	return item.Value, true
}

func (p *slashCompletionProvider) Cancel(ctx completionContext, result completionResult) (string, bool) {
	p.s.Reset()
	return "", true
}

func (p *slashCompletionProvider) RenderCompletion(ctx completionContext, th tui.Theme, width int) []string {
	return p.s.Render(ctx.Input, th, width)
}
