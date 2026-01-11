package workspace

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed assets
var assets embed.FS

const (
	AppName     = "b75"
	ProblemsDir = "problems"
)

// GetProblemPath returns the absolute path to a specific problem directory in the user's workspace.
func GetProblemPath(problemSlug string) (string, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, ProblemsDir, problemSlug), nil
}

// EnsureProblem verifies that the problem files exist in the user's workspace.
// If not, it copies them from the embedded assets.
func EnsureProblem(problemSlug string) error {
	destDir, err := GetProblemPath(problemSlug)
	if err != nil {
		return err
	}

	if _, err := os.Stat(destDir); !os.IsNotExist(err) {
		// Problem directory already exists, assume it's initialized
		return nil
	}

	// Create the directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create problem directory: %w", err)
	}

	// Copy assets
	srcRoot := "assets/problems/" + problemSlug

	return fs.WalkDir(assets, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from srcRoot
		relPath, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(destDir, relPath)

		// Rename go.mod.tpl to go.mod
		if filepath.Base(path) == "go.mod.tpl" {
			destPath = filepath.Join(filepath.Dir(destPath), "go.mod")
		}

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read file content
		content, err := assets.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, content, 0644)
	})
}

func getDataDir() (string, error) {
	// XDG_DATA_HOME logic
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, AppName), nil
}
