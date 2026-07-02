package modes

import (
	"os"
	"sort"
	"strings"
)

type pathCompletionCandidate struct {
	name  string
	isDir bool
}

type pathTabCompletionResult struct {
	consumed      bool
	tokenStart    int
	replacement   string
	changed       bool
	displayParent string
	candidates    []pathCompletionCandidate
}

// slashArgumentCompletion rewrites low-risk slash command arguments using the
// same shell-style path completion policy as editor Tab completion.
func slashArgumentCompletion(input, cwd string) (string, bool) {
	cmd, restStart, ok := splitSlashCommandInput(input)
	if !ok {
		return "", false
	}
	rest := input[restStart:]
	switch cmd {
	case "/study":
		if rest == "" || trailingWhitespace(rest) {
			return "", false
		}
		completed, changed := completePathText(rest, cwd)
		if !changed {
			return "", false
		}
		return input[:restStart] + completed, true
	case "/session":
		importStart := firstFieldStart(rest)
		if importStart < 0 || rest[importStart:] != "import" && !strings.HasPrefix(rest[importStart:], "import ") && !strings.HasPrefix(rest[importStart:], "import\t") {
			return "", false
		}
		importEnd := importStart + len("import")
		if importEnd < len(rest) && rest[importEnd] != ' ' && rest[importEnd] != '\t' {
			return "", false
		}
		pathStart := importEnd
		for pathStart < len(rest) && (rest[pathStart] == ' ' || rest[pathStart] == '\t') {
			pathStart++
		}
		if pathStart == len(rest) || trailingWhitespace(rest[pathStart:]) {
			return "", false
		}
		completed, changed := completePathText(rest[pathStart:], cwd)
		if !changed {
			return "", false
		}
		return input[:restStart+pathStart] + completed, true
	}
	return "", false
}

func splitSlashCommandInput(input string) (cmd string, restStart int, ok bool) {
	if !strings.HasPrefix(input, "/") {
		return "", 0, false
	}
	if idx := strings.IndexAny(input, " \t"); idx >= 0 {
		restStart = idx
		for restStart < len(input) && (input[restStart] == ' ' || input[restStart] == '\t') {
			restStart++
		}
		return input[:idx], restStart, true
	}
	return input, len(input), true
}

func firstFieldStart(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' {
			return i
		}
	}
	return -1
}

func trailingWhitespace(s string) bool {
	if s == "" {
		return false
	}
	last := s[len(s)-1]
	return last == ' ' || last == '\t' || last == '\n'
}

func completePathText(text, cwd string) (string, bool) {
	res := pathTabComplete(text, cwd)
	if !res.consumed || !res.changed {
		return "", false
	}
	return text[:res.tokenStart] + res.replacement, true
}

func pathTabComplete(text, cwd string) pathTabCompletionResult {
	start := len(text)
	for start > 0 {
		r := text[start-1]
		if r == ' ' || r == '\t' || r == '\n' {
			break
		}
		start--
	}
	token := text[start:]
	res := completePathTokenDetailed(token, cwd)
	res.tokenStart = start
	return res
}

func completePathTokenDetailed(token, cwd string) pathTabCompletionResult {
	res := pathTabCompletionResult{replacement: token}
	if token == "" || !looksLikePathToken(token) {
		return res
	}
	res.consumed = true
	parentAbs, basePrefix, displayParent, ok := resolvePathTabToken(token, cwd)
	res.displayParent = displayParent
	if !ok {
		return res
	}
	entries, err := os.ReadDir(parentAbs)
	if err != nil {
		return res
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, basePrefix) {
			continue
		}
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(basePrefix, ".") {
			continue
		}
		res.candidates = append(res.candidates, pathCompletionCandidate{name: name, isDir: e.IsDir()})
	}
	if len(res.candidates) == 0 {
		return res
	}
	sort.Slice(res.candidates, func(i, j int) bool { return res.candidates[i].name < res.candidates[j].name })
	completed := res.candidates[0].name
	completedIsDir := res.candidates[0].isDir
	if len(res.candidates) > 1 {
		names := make([]string, len(res.candidates))
		for i, m := range res.candidates {
			names[i] = m.name
		}
		completed = longestCommonPrefix(names)
		completedIsDir = false
		if completed == basePrefix {
			return res
		}
	}
	res.replacement = displayParent + completed
	if len(res.candidates) == 1 && completedIsDir {
		res.replacement += "/"
	}
	res.changed = res.replacement != token
	return res
}
