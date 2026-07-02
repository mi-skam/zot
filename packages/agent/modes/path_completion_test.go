package modes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patriceckhart/zot/packages/tui"
)

func TestPathTabCompleteAmbiguousDotSlashStaysPut(t *testing.T) {
	tmp := t.TempDir()
	mustWriteFile(t, filepath.Join(tmp, "alpha.txt"))
	mustWriteFile(t, filepath.Join(tmp, "beta.txt"))
	ed := tui.NewEditor("")
	ed.SetValue("./")
	if !tryPathTabCompleteEditor(ed, tmp) {
		t.Fatal("./ tab should be consumed")
	}
	if got := ed.Value(); got != "./" {
		t.Fatalf("value = %q, want ./", got)
	}
}

func TestPathTabCompleteUniqueRelative(t *testing.T) {
	tmp := t.TempDir()
	mustWriteFile(t, filepath.Join(tmp, "unique.txt"))
	ed := tui.NewEditor("")
	ed.SetValue("./uni")
	if !tryPathTabCompleteEditor(ed, tmp) {
		t.Fatal("expected completion")
	}
	if got := ed.Value(); got != "./unique.txt" {
		t.Fatalf("value = %q", got)
	}
}

func TestPathTabCompletePreservesParentDisplay(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "cwd")
	dir := filepath.Join(root, "dir")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(dir, "file.txt"))
	ed := tui.NewEditor("")
	ed.SetValue("../dir/fi")
	if !tryPathTabCompleteEditor(ed, cwd) {
		t.Fatal("expected completion")
	}
	if got := ed.Value(); got != "../dir/file.txt" {
		t.Fatalf("value = %q", got)
	}
}

func TestPathTabCompletePreservesHomeDisplay(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("no home dir")
	}
	name := "zot_path_completion_test_unique_file"
	path := filepath.Join(home, name)
	if _, err := os.Stat(path); err == nil {
		t.Skipf("%s already exists", path)
	}
	mustWriteFile(t, path)
	defer os.Remove(path)
	ed := tui.NewEditor("")
	ed.SetValue("~/zot_path_completion_test_unique_")
	if !tryPathTabCompleteEditor(ed, t.TempDir()) {
		t.Fatal("expected completion")
	}
	if got := ed.Value(); got != "~/"+name {
		t.Fatalf("value = %q", got)
	}
}

func TestPathTabCompletePlainWordNoop(t *testing.T) {
	ed := tui.NewEditor("")
	ed.SetValue("hello")
	if tryPathTabCompleteEditor(ed, t.TempDir()) {
		t.Fatal("plain word should not be consumed")
	}
	if got := ed.Value(); got != "hello" {
		t.Fatalf("value = %q", got)
	}
}

func TestPathTabCompleteDotfilesHiddenUnlessTyped(t *testing.T) {
	tmp := t.TempDir()
	mustWriteFile(t, filepath.Join(tmp, ".secret"))
	mustWriteFile(t, filepath.Join(tmp, "visible"))
	ed := tui.NewEditor("")
	ed.SetValue("./")
	tryPathTabCompleteEditor(ed, tmp)
	if strings.Contains(ed.Value(), ".secret") {
		t.Fatalf("dotfile completed unexpectedly: %q", ed.Value())
	}
	ed.SetValue("./.sec")
	if !tryPathTabCompleteEditor(ed, tmp) {
		t.Fatal("expected dotfile completion")
	}
	if got := ed.Value(); got != "./.secret" {
		t.Fatalf("value = %q", got)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}
