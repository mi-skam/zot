package modes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompletionArgsStudyPath(t *testing.T) {
	tmp := t.TempDir()
	mustWriteFile(t, filepath.Join(tmp, "path.txt"))
	got, ok := slashArgumentCompletion("/study ./pa", tmp)
	if !ok || got != "/study ./path.txt" {
		t.Fatalf("slashArgumentCompletion = %q ok=%v", got, ok)
	}
}

func TestCompletionArgsSessionImportPath(t *testing.T) {
	tmp := t.TempDir()
	mustWriteFile(t, filepath.Join(tmp, "session.zotsession"))
	got, ok := slashArgumentCompletion("/session import ./sess", tmp)
	if !ok || got != "/session import ./session.zotsession" {
		t.Fatalf("slashArgumentCompletion = %q ok=%v", got, ok)
	}
}

func TestCompletionArgsNoChangeLeavesPopupReachable(t *testing.T) {
	tmp := t.TempDir()
	mustWriteFile(t, filepath.Join(tmp, "alpha.txt"))
	mustWriteFile(t, filepath.Join(tmp, "alpine.txt"))
	if got, ok := slashArgumentCompletion("/study ./alp", tmp); ok || got != "" {
		t.Fatalf("unchanged ambiguous prefix should not consume Tab: %q ok=%v", got, ok)
	}
}

func TestCompletionArgsTrailingWhitespaceNoOp(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(tmp, "docs", "note.txt"))
	if got, ok := slashArgumentCompletion("/study ./docs/ ", tmp); ok || got != "" {
		t.Fatalf("trailing whitespace should end Tab token: %q ok=%v", got, ok)
	}
}

func TestCompletionArgsPreservesInternalWhitespace(t *testing.T) {
	tmp := t.TempDir()
	mustWriteFile(t, filepath.Join(tmp, "path.txt"))
	got, ok := slashArgumentCompletion("/study  ./pa", tmp)
	if !ok || got != "/study  ./path.txt" {
		t.Fatalf("slashArgumentCompletion = %q ok=%v", got, ok)
	}
}

func TestCompletionArgsModelUnchanged(t *testing.T) {
	if got, ok := slashArgumentCompletion("/model ./pa", t.TempDir()); ok || got != "" {
		t.Fatalf("/model should not complete here: %q ok=%v", got, ok)
	}
}
