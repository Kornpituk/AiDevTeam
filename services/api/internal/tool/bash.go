package tool

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const maxOutputBytes = 100 * 1024 // 100KB

// BashInput represents the input for the bash tool.
type BashInput struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"` // seconds, default from opts
}

// BashOutput represents the output from the bash tool.
type BashOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// Bash executes a shell command within the workspace.
func Bash(input BashInput, workspaceRoot string, timeout int, blockedCommands []string) (*BashOutput, error) {
	if input.Command == "" {
		return nil, fmt.Errorf("command is empty")
	}

	// Check against blocked commands list
	if err := IsBashCommandAllowed(input.Command); err != nil {
		return nil, err
	}

	// Check additional blocked commands from config
	for _, blocked := range blockedCommands {
		if blocked != "" && strings.Contains(strings.ToLower(input.Command), strings.ToLower(blocked)) {
			return nil, fmt.Errorf("command contains blocked pattern from configuration: %s", blocked)
		}
	}

	// Use provided timeout or default
	cmdTimeout := timeout
	if input.Timeout > 0 {
		cmdTimeout = input.Timeout
	}
	if cmdTimeout <= 0 {
		cmdTimeout = 30 // default 30 seconds
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cmdTimeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", input.Command)
	cmd.Dir = workspaceRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("command timed out after %d seconds", cmdTimeout)
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute command: %s", err.Error())
		}
	}

	// Truncate output if it exceeds limits
	stdoutStr := truncateOutput(stdout.String(), "stdout")
	stderrStr := truncateOutput(stderr.String(), "stderr")

	return &BashOutput{
		Stdout:   stdoutStr,
		Stderr:   stderrStr,
		ExitCode: exitCode,
	}, nil
}

func truncateOutput(output string, label string) string {
	if len(output) > maxOutputBytes {
		return output[:maxOutputBytes] + fmt.Sprintf("\n... [%s truncated at %d bytes]", label, maxOutputBytes)
	}
	return output
}
