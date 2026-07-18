package mode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

type PythonResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

func RunPythonProcess(ctx context.Context, args string) (string, error) {
	script, err := scriptFromArgs(args)
	if err != nil {
		return pythonResultToJSON(&PythonResult{Success: false, Error: err.Error()}), nil
	}

	cmd := exec.CommandContext(ctx, "py", "-")
	cmd.Stdin = strings.NewReader(script)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	if runErr == nil {
		return pythonResultToJSON(&PythonResult{Success: true, Output: stdout.String()}), nil
	}

	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	var exitErr *exec.ExitError
	if errors.As(runErr, &exitErr) {
		return pythonResultToJSON(&PythonResult{Success: false, Error: stderr.String()}), nil
	}

	return pythonResultToJSON(&PythonResult{Success: false, Error: runErr.Error()}), nil
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
