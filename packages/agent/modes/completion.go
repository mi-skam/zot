package modes

import "github.com/patriceckhart/zot/packages/tui"

type completionContext struct {
	Input string
	CWD   string
}

type completionItem struct {
	Value       string
	Label       string
	Description string
	IsDir       bool
	Header      bool
}

type completionResult struct {
	Provider string
	Items    []completionItem
}

type completionProvider interface {
	Name() string
	Suggestions(ctx completionContext) completionResult
	Apply(ctx completionContext, result completionResult, item completionItem) (string, bool)
	Cancel(ctx completionContext, result completionResult) (string, bool)
}

// completionRenderableProvider lets existing suggesters keep their exact
// popup layout while the manager owns provider priority and navigation.
type completionRenderableProvider interface {
	RenderCompletion(ctx completionContext, th tui.Theme, width int) []string
}
