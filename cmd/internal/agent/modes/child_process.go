package modes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

type PythonResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

func RunPythonProcess(ctx context.Context, args string, rootDir string) (string, error) {
	script, err := scriptFromArgs(args)
	if err != nil {
		return pythonResultToJSON(&PythonResult{Success: false, Error: err.Error()}), nil
	}

	cmd := exec.CommandContext(ctx, "py", "-I", "-")
	cmd.Dir = rootDir
	cmd.Stdin = strings.NewReader(script)

	cmd.Env = []string{
		"SystemRoot=" + os.Getenv("SystemRoot"),
	}

	cmd.WaitDelay = 15 * time.Second

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := runInJobObject(cmd)

	if runErr == nil {

		if stdout.String() == "" {
			return buildUnsuccessResult("No output captured. Use `print`"), nil
		}

		return pythonResultToJSON(&PythonResult{Success: true, Output: stdout.String()}), nil
	}

	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		return buildUnsuccessResult(stderr.String()), nil
	}

	return buildUnsuccessResult(runErr.Error()), nil
}

func buildUnsuccessResult(err string) string {
	return pythonResultToJSON(&PythonResult{Success: false, Error: err})
}

func pythonResultToJSON(result *PythonResult) string {
	b, _ := json.Marshal(result)
	return string(b)
}

func scriptFromArgs(args string) (string, error) {
	var a struct {
		Script string `json:"script"`
	}

	err := json.Unmarshal([]byte(args), &a)
	if err != nil {
		return "", err
	}

	if a.Script == "" {
		return "", errors.New("missing `script` parameter")
	}

	return a.Script, nil
}
