package app

import (
	"os"
	"path/filepath"
)

func ResolveAppPath(relPath string) string {
	if envPath := os.Getenv("TIMESHEESH_ROOT"); envPath != "" {
		return filepath.Join(envPath, relPath)
	}

	workingDir, _ := os.Getwd()
	execPath, _ := os.Executable()
	return resolveAppPath(relPath, workingDir, execPath)
}

func resolveAppPath(relPath string, workingDir string, execPath string) string {
	for _, root := range candidateRoots(workingDir, execPath) {
		candidate := filepath.Join(root, relPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	if execPath != "" {
		return filepath.Join(filepath.Dir(execPath), relPath)
	}
	if workingDir != "" {
		return filepath.Join(workingDir, relPath)
	}
	return relPath
}

func candidateRoots(workingDir string, execPath string) []string {
	roots := make([]string, 0)
	seen := make(map[string]struct{})

	addAncestors := func(start string) {
		if start == "" {
			return
		}
		for _, dir := range ancestorDirs(start) {
			if _, ok := seen[dir]; ok {
				continue
			}
			seen[dir] = struct{}{}
			roots = append(roots, dir)
		}
	}

	addAncestors(filepath.Dir(execPath))
	addAncestors(workingDir)

	return roots
}

func ancestorDirs(start string) []string {
	cleaned := filepath.Clean(start)
	if cleaned == "." || cleaned == "" {
		return nil
	}

	dirs := []string{cleaned}
	for {
		parent := filepath.Dir(cleaned)
		if parent == cleaned {
			break
		}
		dirs = append(dirs, parent)
		cleaned = parent
	}
	return dirs
}
