package runner

import (
	"bytes"
	"os/exec"

	"github.com/haljac/b75/internal/workspace"
)

type Result struct {
	Passed bool
	Output string
}

// RunTests executes 'go test' in the problem directory.
func RunTests(slug string) (Result, error) {
	path, err := workspace.GetProblemPath(slug)
	if err != nil {
		return Result{}, err
	}

	cmd := exec.Command("go", "test", "-v", ".")
	cmd.Dir = path

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()

	output := out.String()
	passed := err == nil

	return Result{
		Passed: passed,
		Output: output,
	}, nil
}
