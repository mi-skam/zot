# Completion Provider Refactor Implementation Plan

> **For agentic workers:** This is a large interactive-TUI refactor. Implement task-by-task with small, separately tested changes. Do not change user-facing completion behavior unless the task explicitly says so.

**Goal:** Replace zot's hardcoded autocomplete paths with a shared completion-provider architecture that supports slash commands, file mentions, shell-style path completion, command argument completion, and future extension-provided completions.


**Architecture:** Introduce a small completion core in `packages/agent/modes` with provider interfaces, a manager that picks the active provider by priority, and a reusable popup renderer/navigation model. Wrap existing slash and file suggesters first, then migrate tab path completion and selected slash-command argument completions. Existing dialogs remain unchanged unless replaced by an argument completion.

**Tech Stack:** Go 1.22, existing `packages/tui.Editor`, existing `slashSuggester`, existing `fileSuggester`, existing tests with `go test`.

---

## Current State

Relevant files:

- `packages/agent/modes/interactive.go`
  - Wires slash popup, file popup, key handling, and `tryPathTabCompleteEditor` directly.
- `packages/agent/modes/slash_suggest.go`
  - Owns slash command matching, navigation, rendering, and extension slash command rows.
- `packages/agent/modes/file_suggest.go`
  - Owns `@` file picker, recursive scans, `.gitignore`, chips, and popup rendering.
- `packages/tui/editor.go`
  - Owns editor state, paste placeholders, drag/drop file chips, and submit expansion.

Current behavior to preserve initially:

1. `/` opens slash-command suggestions.
2. `@` at token start opens file suggestions.
3. `@` picker inserts `[file:path]` / `[dir:path/]` chips.
4. `./foo<Tab>`, `../foo<Tab>`, `/foo<Tab>`, `~/foo<Tab>`, and `foo/bar<Tab>` use shell-style path completion.
5. Ambiguous `./<Tab>` consumes tab but does not open a popup.
6. Plain words do not complete on tab.
7. Extension slash commands still appear under the slash popup extension section.

---

## Target Design

### Core Types

Create `packages/agent/modes/completion.go`:

```go
package modes

import "github.com/patriceckhart/zot/packages/tui"

type completionTrigger int

const (
	completionTriggerAuto completionTrigger = iota
	completionTriggerTab
)

type completionContext struct {
	Input   string
	Cursor  int
	CWD     string
	Trigger completionTrigger
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
	Prefix   string
	Items    []completionItem
}

type completionProvider interface {
	Name() string
	Suggestions(ctx completionContext) completionResult
	Apply(ctx completionContext, result completionResult, item completionItem) (string, int, bool)
	Cancel(ctx completionContext, result completionResult) (string, int, bool)
}
```

Notes:

- `Cursor` is included even though zot currently keeps the cursor at the end after `SetValue`. This prevents baking in that limitation.
- `Apply` returns `(newInput, newCursor, ok)`.
- `Cancel` lets providers decide whether Esc removes a prefix (`@foo`) or just closes the popup.
- `Header` keeps slash extension section headers representable without special UI code.

### Completion Manager

Create `packages/agent/modes/completion_manager.go`:

```go
type completionManager struct {
	providers []completionProvider
	active    string
	cursor    int
	last      completionResult
}
```

Responsibilities:

1. Ask providers in priority order for suggestions.
2. Track the active provider and selected row.
3. Expose `Active()`, `Render()`, `Up()`, `Down()`, `PageUp()`, `PageDown()`, `Apply()`, `Cancel()`, and `Reset()`.
4. Skip header rows during navigation.
5. Render a shared popup shape using the current theme.

Initial provider priority:

1. Slash provider (`/` command names and command arguments)
2. File mention provider (`@` picker)
3. Optional future providers

Path-tab completion is initially not a popup provider. Keep it as a separate helper until the shell-style completion behavior is fully pinned by tests.

---

## Behavior Spec

### Global Completion Behavior

1. Only one completion provider is active at a time.
2. Slash suggestions win over file suggestions when input starts with `/`.
3. File suggestions only activate for explicit `@` mentions at token boundaries.
4. `./<Tab>` must not open the file picker popup.
5. `Tab` while a popup is open applies the highlighted suggestion when that provider supports tab application.
6. `Enter` while a popup is open applies/runs according to existing behavior:
   - Slash provider: preserve current slash popup semantics.
   - File provider: insert chip.
7. `Esc` while a popup is open cancels that popup and must not cancel an active model turn.
8. Up/down/page navigation skips header rows.
9. Closing a popup still triggers the existing full redraw path to avoid stale overlay rows.

### File Mention Provider

1. `@` at input start opens the picker.
2. ` @` after whitespace opens the picker.
3. `foo@` does not open the picker.
4. `@foo` narrows file suggestions.
5. In flat mode, `right` opens a selected directory and clears the stale filter.
6. In flat mode, `left` returns to parent and clears the stale filter.
7. In recursive mode, `right` and `left` do nothing.
8. `Enter` inserts `[file:rel] ` or `[dir:rel/] `.
9. `Esc` removes the active `@query` and closes the picker.
10. Recursive mode continues to honor nested `.gitignore` and entry/depth caps.

### Path Tab Completion

1. `./foo<Tab>` completes by longest common prefix or single match.
2. `../foo<Tab>` completes by longest common prefix or single match.
3. `/foo<Tab>` completes by longest common prefix or single match.
4. `~/foo<Tab>` completes by longest common prefix or single match.
5. `foo/bar<Tab>` completes by longest common prefix or single match.
6. `./<Tab>` consumes tab and remains `./` when ambiguous.
7. Plain `hello<Tab>` is a no-op and does not consume special popup state.
8. Dotfiles are hidden unless the typed basename starts with `.`.
9. Future quoted path support must preserve quotes and escape behavior, but is not required in the first migration task.

### Slash Command Argument Completion

After the core migration, add opt-in argument completion for commands where it is low-risk:

1. `/study <path>` uses path/file completion.
2. `/session import <path>` uses path/file completion.
3. `/logout <provider>` completes logged-in provider ids.
4. `/model <query>` completes available model ids/providers if practical.

Argument completions must not replace existing dialogs. For example, `/model` with no args still opens the model picker.

### Extension Completion Future Hook

The core should leave room for extensions to register completions later, but the first refactor does not need to change the extension JSON-RPC protocol.

Future target shape:

```json
{
  "type": "completion",
  "command": "jira",
  "prefix": "MIS",
  "items": [
    { "value": "MIS-123", "label": "MIS-123", "description": "Fix login flow" }
  ]
}
```

---

## File Structure

Create:

- `packages/agent/modes/completion.go`
  - Shared types and helper functions.
- `packages/agent/modes/completion_manager.go`
  - Provider orchestration, navigation, and shared rendering.
- `packages/agent/modes/completion_manager_test.go`
  - Provider priority, navigation, header skipping, apply/cancel behavior.
- `packages/agent/modes/completion_slash.go`
  - Slash provider wrapper around existing slash command catalog logic.
- `packages/agent/modes/completion_file.go`
  - File provider wrapper around existing file suggester logic.
- `packages/agent/modes/completion_args.go`
  - Optional later task: slash argument completion helpers.

Modify:

- `packages/agent/modes/interactive.go`
  - Replace direct slash/file popup branching with `completionManager` once wrappers are in place.
  - Keep `tryPathTabCompleteEditor` until path completion migration is separately tested.
- `packages/agent/modes/slash_suggest.go`
  - Keep existing public behavior; extract reusable matching/catalog helpers if needed.
- `packages/agent/modes/file_suggest.go`
  - Keep scanner and chip expansion. Extract query parsing/apply helpers if needed.
- `packages/agent/modes/file_suggest_test.go`
  - Add regression coverage for provider wrapper behavior if not covered elsewhere.
- `README.md`
  - Update only after behavior changes are user-visible.

---

## Migration Plan

### Task 1: Pin existing completion behavior with tests

**Files:**

- Modify: `packages/agent/modes/file_suggest_test.go`
- Create or modify: `packages/agent/modes/path_completion_test.go`

- [ ] Add tests for `tryPathTabCompleteEditor`:
  - `./<Tab>` ambiguous stays `./` and returns true.
  - `./uni<Tab>` completes to `./unique.txt`.
  - `../dir/fi<Tab>` preserves `../` display form.
  - `~/fi<Tab>` preserves `~/` display form.
  - `hello<Tab>` returns false and leaves text unchanged.
  - dotfiles are hidden unless basename starts with `.`.

- [ ] Add tests for file picker activation:
  - `@` active.
  - `hello @` active.
  - `hello@` inactive.
  - `@foo bar` inactive after whitespace in query.

Run:

```bash
go test ./packages/agent/modes -run 'Test.*Path.*|TestFileSuggester'
```

### Task 2: Add completion core and manager

**Files:**

- Create: `packages/agent/modes/completion.go`
- Create: `packages/agent/modes/completion_manager.go`
- Create: `packages/agent/modes/completion_manager_test.go`

- [ ] Implement core types.
- [ ] Implement manager provider priority.
- [ ] Implement cursor navigation with header skipping.
- [ ] Implement apply/cancel dispatch.
- [ ] Implement shared popup rendering or a minimal render abstraction that can reproduce current rows.

Tests:

- Higher-priority provider wins.
- Inactive providers are ignored.
- Up/down skip header rows.
- Apply calls the active provider with the selected item.
- Cancel calls the active provider.
- Reset clears active result and cursor.

Run:

```bash
go test ./packages/agent/modes -run TestCompletionManager
```

### Task 3: Wrap slash suggestions as a provider

**Files:**

- Create: `packages/agent/modes/completion_slash.go`
- Modify: `packages/agent/modes/slash_suggest.go` if extraction is needed.
- Modify: `packages/agent/modes/completion_manager_test.go` or create `completion_slash_test.go`.

- [ ] Implement `slashCompletionProvider` using existing slash catalog/matching behavior.
- [ ] Preserve extension command section headers.
- [ ] Preserve `/help` style descriptions.
- [ ] Preserve tab/enter completion semantics for partial slash prefixes.

Do not wire into `interactive.go` yet unless tests prove rendered output is equivalent.

Run:

```bash
go test ./packages/agent/modes -run 'TestSlash|TestCompletionSlash'
```

### Task 4: Wrap file suggestions as a provider

**Files:**

- Create: `packages/agent/modes/completion_file.go`
- Modify: `packages/agent/modes/file_suggest.go` if helper extraction is needed.
- Create: `packages/agent/modes/completion_file_test.go`

- [ ] Implement `fileCompletionProvider` using the existing `fileSuggester` scanner and state.
- [ ] Preserve flat browse mode.
- [ ] Preserve recursive fuzzy mode.
- [ ] Preserve `.gitignore` and nested `.gitignore` behavior.
- [ ] Preserve chip insertion.
- [ ] Preserve Esc removal of the active `@query`.
- [ ] Preserve Right/Left directory navigation.

Run:

```bash
go test ./packages/agent/modes -run 'TestFileSuggester|TestCompletionFile'
```

### Task 5: Wire completion manager into interactive mode

**Files:**

- Modify: `packages/agent/modes/interactive.go`

- [ ] Add `completion *completionManager` to `Interactive`.
- [ ] Initialize it with slash and file providers.
- [ ] Replace render-time slash/file popup branching with manager rendering.
- [ ] Replace key handling blocks for slash/file popup with manager navigation/apply/cancel.
- [ ] Preserve special slash Enter behavior for running ambiguous partial slash command completions.
- [ ] Preserve `right`/`left` behavior for file provider.
- [ ] Keep path-tab completion guard:

```go
if k.Kind == tui.KeyTab && !i.completion.Active() {
	if i.tryPathTabComplete() { return false }
}
```

- [ ] Ensure busy-turn Esc behavior still does not cancel the turn when a completion popup is open.
- [ ] Ensure overlay close full redraw still triggers.

Run targeted tests:

```bash
go test ./packages/agent/modes
```

Manual smoke tests in tmux:

```bash
tmux new-session -d -s zot-complete -x 100 -y 30 './bin/zot --no-session'
tmux send-keys -t zot-complete '/'
tmux capture-pane -t zot-complete -p
tmux send-keys -t zot-complete Escape
tmux send-keys -t zot-complete '@'
tmux capture-pane -t zot-complete -p
tmux send-keys -t zot-complete Escape
tmux send-keys -t zot-complete './' Tab
tmux capture-pane -t zot-complete -p
tmux kill-session -t zot-complete
```

Expected:

- `/` shows slash suggestions.
- `@` shows file suggestions.
- `./<Tab>` does not show file suggestions.

### Task 6: Add command argument completion hooks

**Files:**

- Create: `packages/agent/modes/completion_args.go`
- Modify: `packages/agent/modes/completion_slash.go`
- Add tests in `packages/agent/modes/completion_args_test.go`

- [ ] Add an optional argument completion callback to slash command metadata, or a separate map keyed by command name.
- [ ] Implement path argument completion for `/study`.
- [ ] Implement path argument completion for `/session import`.
- [ ] Implement provider-id completion for `/logout`.
- [ ] Keep `/model` for a later task unless model registry access is straightforward without introducing tight coupling.

Behavior:

- `/study ./pa<Tab>` should path-complete.
- `/study @pa` may use file-chip completion if explicit `@` is present.
- `/session import ~/Downloads/foo<Tab>` should path-complete.
- `/logout ope<Tab>` should complete provider names.

Run:

```bash
go test ./packages/agent/modes -run 'TestCompletionArgs|TestSlash'
```

### Task 7: Document and clean up

**Files:**

- Modify: `README.md`
- Modify/remove old helper comments in `interactive.go`, `slash_suggest.go`, `file_suggest.go`.

- [ ] Document any new argument completions.
- [ ] Confirm no user-facing behavior changed unexpectedly.
- [ ] Delete dead code only after tests cover replacement behavior.

Run:

```bash
go test ./packages/agent/modes ./packages/tui
```

---

## Acceptance Criteria

1. Existing slash popup behavior is preserved.
2. Existing `@` file picker behavior is preserved.
3. Existing recursive and `.gitignore` file picker behavior is preserved.
4. Existing tab path completion behavior is preserved.
5. `./<Tab>` does not open a file picker popup.
6. Completion manager has unit tests for priority, navigation, apply, cancel, and reset.
7. Slash and file providers have focused tests independent of full interactive mode.
8. `/study` and `/session import` get path argument completion after the core migration.
9. Extension slash commands still appear in the popup.
10. No full test suite or build command is required unless requested; use targeted `go test` commands while implementing.

---

## Risks and Mitigations

### Risk: Regressing busy-turn Esc behavior

Mitigation: keep a test/manual smoke case where a popup is open during a busy turn and Esc closes the popup instead of cancelling the turn.

### Risk: Rendering differences create stale overlay rows

Mitigation: keep existing overlay-open tracking and force-clear behavior when completion popup closes.

### Risk: Header rows become selectable

Mitigation: central manager tests for header skipping.

### Risk: File picker loses nested `.gitignore` behavior

Mitigation: do not replace scanner. Wrap `fileSuggester` first; keep existing tests.

### Risk: Provider abstraction becomes too generic too early

Mitigation: keep the interface minimal. Add async/dynamic extension completions only after the local providers are migrated.

### Risk: Path completion and file picker semantics blur

Mitigation: maintain separate policies:

- `@` opens file picker.
- Path-like `<Tab>` performs shell-style completion.
- Ambiguous `./<Tab>` remains non-popup.
