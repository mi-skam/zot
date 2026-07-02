package modes

import (
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/patriceckhart/zot/packages/tui"
)

type completionCursorSync interface {
	setCompletionCursor(int)
}

type completionBrowser interface {
	BrowseRight() bool
	BrowseLeft() bool
}

type completionManager struct {
	providers []completionProvider
	active    string
	cursor    int
	last      completionResult
}

func newCompletionManager(providers ...completionProvider) *completionManager {
	return &completionManager{providers: providers}
}

func (m *completionManager) Refresh(ctx completionContext) bool {
	for _, p := range m.providers {
		res := p.Suggestions(ctx)
		if len(res.Items) == 0 {
			continue
		}
		if res.Provider == "" {
			panic("completion provider " + p.Name() + " returned empty Provider")
		}
		if m.active != res.Provider {
			m.cursor = 0
		}
		m.active = res.Provider
		m.last = res
		m.clampCursor()
		return true
	}
	m.Reset()
	return false
}

func (m *completionManager) Active() bool           { return m.active != "" && len(m.last.Items) > 0 }
func (m *completionManager) ActiveProvider() string { return m.active }

func (m *completionManager) Reset() {
	m.active = ""
	m.cursor = 0
	m.last = completionResult{}
}

func (m *completionManager) provider(name string) completionProvider {
	for _, p := range m.providers {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

func (m *completionManager) selectable(i int) bool {
	return i >= 0 && i < len(m.last.Items) && !m.last.Items[i].Header
}

func (m *completionManager) clampCursor() {
	n := len(m.last.Items)
	if n == 0 {
		m.cursor = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= n {
		m.cursor = n - 1
	}
	if m.selectable(m.cursor) {
		return
	}
	for i := m.cursor + 1; i < n; i++ {
		if !m.last.Items[i].Header {
			m.cursor = i
			return
		}
	}
	for i := m.cursor - 1; i >= 0; i-- {
		if !m.last.Items[i].Header {
			m.cursor = i
			return
		}
	}
	m.cursor = 0
}

func (m *completionManager) move(step int) {
	if !m.Active() {
		return
	}
	n := len(m.last.Items)
	idx := m.cursor + step
	for idx >= 0 && idx < n && m.last.Items[idx].Header {
		idx += step
	}
	if idx < 0 {
		for i, it := range m.last.Items {
			if !it.Header {
				m.cursor = i
				return
			}
		}
		return
	}
	if idx >= n {
		for i := n - 1; i >= 0; i-- {
			if !m.last.Items[i].Header {
				m.cursor = i
				return
			}
		}
		return
	}
	m.cursor = idx
}

func (m *completionManager) syncCursor() {
	if p := m.provider(m.active); p != nil {
		if cp, ok := p.(completionCursorSync); ok {
			cp.setCompletionCursor(m.cursor)
		}
	}
}

func (m *completionManager) Up()   { m.move(-1); m.syncCursor() }
func (m *completionManager) Down() { m.move(+1); m.syncCursor() }
func (m *completionManager) PageUp() {
	m.cursor -= slashSuggestPageSize
	m.clampCursor()
	m.syncCursor()
}
func (m *completionManager) PageDown() {
	m.cursor += slashSuggestPageSize
	m.clampCursor()
	m.syncCursor()
}

func (m *completionManager) BrowseRight() bool {
	p := m.provider(m.active)
	bp, ok := p.(completionBrowser)
	if !ok || !bp.BrowseRight() {
		return false
	}
	m.cursor = 0
	m.syncCursor()
	return true
}

func (m *completionManager) BrowseLeft() bool {
	p := m.provider(m.active)
	bp, ok := p.(completionBrowser)
	if !ok || !bp.BrowseLeft() {
		return false
	}
	m.cursor = 0
	m.syncCursor()
	return true
}

func (m *completionManager) Selected() (completionItem, bool) {
	if !m.Active() {
		return completionItem{}, false
	}
	m.clampCursor()
	if !m.selectable(m.cursor) {
		return completionItem{}, false
	}
	return m.last.Items[m.cursor], true
}

func (m *completionManager) Apply(ctx completionContext) (string, bool) {
	item, ok := m.Selected()
	if !ok {
		return "", false
	}
	p := m.provider(m.active)
	if p == nil {
		return "", false
	}
	return p.Apply(ctx, m.last, item)
}

func (m *completionManager) Cancel(ctx completionContext) (string, bool) {
	p := m.provider(m.active)
	if p == nil {
		return "", false
	}
	return p.Cancel(ctx, m.last)
}

func (m *completionManager) Render(ctx completionContext, th tui.Theme, width int) []string {
	if !m.Refresh(ctx) {
		return nil
	}
	m.syncCursor()
	if p := m.provider(m.active); p != nil {
		if rp, ok := p.(completionRenderableProvider); ok {
			rows := rp.RenderCompletion(ctx, th, width)
			if len(rows) == 0 {
				panic("completion provider " + p.Name() + " rendered zero rows for active suggestions")
			}
			return rows
		}
	}
	return m.renderGeneric(th, width)
}

func (m *completionManager) renderGeneric(th tui.Theme, width int) []string {
	m.clampCursor()
	nameWidth := 10
	for _, it := range m.last.Items {
		if it.Header {
			continue
		}
		if n := runewidth.StringWidth(it.Label); n > nameWidth {
			nameWidth = n
		}
	}
	var lines []string
	for idx, it := range m.last.Items {
		if it.Header {
			lines = append(lines, "")
			rule := strings.Repeat("─", width)
			label := "── " + it.Label + " "
			if lw := runewidth.StringWidth(label); lw < width {
				rule = label + strings.Repeat("─", width-lw)
			}
			lines = append(lines, th.FG256(th.Muted, rule), "")
			continue
		}
		label := it.Label
		if w := runewidth.StringWidth(label); w < nameWidth {
			label += strings.Repeat(" ", nameWidth-w)
		}
		plain := "  " + label
		if it.Description != "" {
			plain += "  " + it.Description
		}
		if idx == m.cursor {
			lines = append(lines, th.PadHighlight(plain, width))
		} else {
			lines = append(lines, th.FG256(th.Muted, plain))
		}
	}
	lines = append(lines, "", th.FG256(th.Muted, "  ↑/↓ navigate - tab complete - enter select"), "")
	return lines
}
