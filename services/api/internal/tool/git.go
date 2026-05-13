package tool

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

const gitTimeout = 60 // seconds

// GitInput represents the input for the git tool.
type GitInput struct {
	Args []string `json:"args"`
}

// GitOutput represents the output from the git tool.
type GitOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// Git executes a git command within the workspace.
func Git(input GitInput, workspaceRoot string) (*GitOutput, error) {
	if err := IsGitCommandAllowed(input.Args); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", input.Args...)
	cmd.Dir = workspaceRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("git command timed out after %d seconds", gitTimeout)
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute git command: %s", err.Error())
		}
	}

	stdoutStr := truncateOutput(stdout.String(), "stdout")
	stderrStr := truncateOutput(stderr.String(), "stderr")

	return &GitOutput{
		Stdout:   stdoutStr,
		Stderr:   stderrStr,
		ExitCode: exitCode,
	}, nil
}
