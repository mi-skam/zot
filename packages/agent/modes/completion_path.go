package modes

const pathCompletionProviderName = "path"

type pathCompletionProvider struct {
	open       bool
	input      string
	tokenStart int
	items      []completionItem
}

func newPathCompletionProvider() *pathCompletionProvider { return &pathCompletionProvider{} }

func (p *pathCompletionProvider) Name() string { return pathCompletionProviderName }

func (p *pathCompletionProvider) Open(ctx completionContext) bool {
	res := pathTabComplete(ctx.Input, ctx.CWD)
	if !res.consumed || len(res.candidates) == 0 {
		p.Reset()
		return false
	}
	items := make([]completionItem, 0, len(res.candidates))
	for _, c := range res.candidates {
		label := c.name
		value := res.displayParent + c.name
		if c.isDir {
			label += "/"
			value += "/"
		}
		items = append(items, completionItem{Value: value, Label: label, IsDir: c.isDir})
	}
	p.open = true
	p.input = ctx.Input
	p.tokenStart = res.tokenStart
	p.items = items
	return true
}

func (p *pathCompletionProvider) Reset() {
	p.open = false
	p.input = ""
	p.tokenStart = 0
	p.items = nil
}

func (p *pathCompletionProvider) Suggestions(ctx completionContext) completionResult {
	if !p.open || ctx.Input != p.input {
		p.Reset()
		return completionResult{Provider: p.Name()}
	}
	return completionResult{Provider: p.Name(), Items: p.items}
}

func (p *pathCompletionProvider) Apply(ctx completionContext, result completionResult, item completionItem) (string, bool) {
	if item.Value == "" || p.tokenStart > len(ctx.Input) {
		return "", false
	}
	out := ctx.Input[:p.tokenStart] + item.Value
	p.Reset()
	return out, true
}

func (p *pathCompletionProvider) Cancel(ctx completionContext, result completionResult) (string, bool) {
	p.Reset()
	return ctx.Input, true
}
