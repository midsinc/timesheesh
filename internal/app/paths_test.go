package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAppPathPrefersExistingFileNearExecutable(t *testing.T) {
	root := t.TempDir()
	external := filepath.Join(t.TempDir(), "outside")
	if err := os.MkdirAll(external, 0o755); err != nil {
		t.Fatalf("mkdir external: %v", err)
	}

	dbFile := filepath.Join(root, "timesheesh.db")
	if err := os.WriteFile(dbFile, []byte("db"), 0o644); err != nil {
		t.Fatalf("write db: %v", err)
	}

	resolved := resolveAppPath("timesheesh.db", external, filepath.Join(root, "timesheesh"))
	if resolved != dbFile {
		t.Fatalf("expected executable-adjacent db %q, got %q", dbFile, resolved)
	}
}

func TestResolveAppPathFallsBackToExecutableDirWhenNothingExists(t *testing.T) {
	root := t.TempDir()
	external := filepath.Join(t.TempDir(), "outside")
	if err := os.MkdirAll(external, 0o755); err != nil {
		t.Fatalf("mkdir external: %v", err)
	}

	expected := filepath.Join(root, "timesheesh.db")
	resolved := resolveAppPath("timesheesh.db", external, filepath.Join(root, "timesheesh"))
	if resolved != expected {
		t.Fatalf("expected fallback db path %q, got %q", expected, resolved)
	}
}

func TestResolveAppPathPrefersExecutableTreeOverLaunchDirectoryCopy(t *testing.T) {
	root := t.TempDir()
	launchDir := filepath.Join(t.TempDir(), "outside")
	if err := os.MkdirAll(launchDir, 0o755); err != nil {
		t.Fatalf("mkdir launch dir: %v", err)
	}

	execDB := filepath.Join(root, "timesheesh.db")
	if err := os.WriteFile(execDB, []byte("real"), 0o644); err != nil {
		t.Fatalf("write exec db: %v", err)
	}
	launchDB := filepath.Join(launchDir, "timesheesh.db")
	if err := os.WriteFile(launchDB, []byte("empty"), 0o644); err != nil {
		t.Fatalf("write launch db: %v", err)
	}

	resolved := resolveAppPath("timesheesh.db", launchDir, filepath.Join(root, "timesheesh"))
	if resolved != execDB {
		t.Fatalf("expected executable db %q, got %q", execDB, resolved)
	}
}

func TestResolveAppPathFindsAssetsInWorkingTreeAncestors(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "cmd", "timesheesh")
	if err := os.MkdirAll(filepath.Join(root, "static"), 0o755); err != nil {
		t.Fatalf("mkdir static: %v", err)
	}
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	indexFile := filepath.Join(root, "static", "index.html")
	if err := os.WriteFile(indexFile, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	resolved := resolveAppPath(filepath.Join("static", "index.html"), nested, filepath.Join(t.TempDir(), "timesheesh"))
	if resolved != indexFile {
		t.Fatalf("expected ancestor asset %q, got %q", indexFile, resolved)
	}
}
