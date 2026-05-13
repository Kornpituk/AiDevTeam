package tool

import (
	"encoding/json"
	"fmt"
)

// ToolDefinition describes a registered tool.
type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"input_schema"`
}

// toolRegistry maps tool names to their execution functions.
type toolFunc func(json.RawMessage, string, ToolOptions) (any, error)

// ToolOptions contains configuration for tool execution.
type ToolOptions struct {
	WorkspaceRoot     string
	ReadMaxBytes      int64
	SearchMaxResults  int
	MaxToolIterations int
	RequireApproval   []string
	ToolChoice        string
}

// RequiresApproval returns true if the given tool name requires human approval before execution.
func (opts ToolOptions) RequiresApproval(toolName string) bool {
	for _, name := range opts.RequireApproval {
		if name == toolName {
			return true
		}
	}
	return false
}

// RegisteredTool holds a tool definition and its implementation.
type RegisteredTool struct {
	Definition ToolDefinition
	Execute    toolFunc
}

var registry map[string]*RegisteredTool

func init() {
	registry = make(map[string]*RegisteredTool)
	registerTool(listFilesDef, executeListFiles)
	registerTool(readFileDef, executeReadFile)
	registerTool(searchCodeDef, executeSearchCode)
}

func registerTool(def ToolDefinition, fn toolFunc) {
	registry[def.Name] = &RegisteredTool{
		Definition: def,
		Execute:    fn,
	}
}

// GetTool returns a registered tool by name.
func GetTool(name string) (*RegisteredTool, error) {
	t, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return t, nil
}

// ListTools returns all registered tool definitions.
func ListTools() []ToolDefinition {
	defs := make([]ToolDefinition, 0, len(registry))
	for _, t := range registry {
		defs = append(defs, t.Definition)
	}
	return defs
}

var listFilesDef = ToolDefinition{
	Name:        "list_files",
	Description: "List files and directories at a given path within the workspace. Returns file names, sizes, and whether each entry is a directory.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Relative path to list (e.g. '.' for root, 'src' for src directory)",
			},
			"depth": map[string]any{
				"type":        "integer",
				"description": "Depth of recursion. 1 = shallow (default), 2+ = deeper recursion",
			},
		},
	},
}

var readFileDef = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a file within the workspace. Text files only; binary files are rejected.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Relative path to the file",
			},
		},
		"required": []string{"path"},
	},
}

var searchCodeDef = ToolDefinition{
	Name:        "search_code",
	Description: "Search for a pattern in code files within the workspace. Supports plain text and regex patterns.",
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Text or regex pattern to search for",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Optional: restrict search to a subdirectory",
			},
			"is_regex": map[string]any{
				"type":        "boolean",
				"description": "If true, treat pattern as a regular expression",
			},
		},
		"required": []string{"pattern"},
	},
}

func executeListFiles(input json.RawMessage, workspaceRoot string, opts ToolOptions) (any, error) {
	var in ListFilesInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("invalid list_files input: %s", err.Error())
	}
	result, err := ListFiles(in, workspaceRoot)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func executeReadFile(input json.RawMessage, workspaceRoot string, opts ToolOptions) (any, error) {
	var in ReadFileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("invalid read_file input: %s", err.Error())
	}
	result, err := ReadFile(in, workspaceRoot, opts.ReadMaxBytes)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func executeSearchCode(input json.RawMessage, workspaceRoot string, opts ToolOptions) (any, error) {
	var in SearchCodeInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("invalid search_code input: %s", err.Error())
	}
	result, err := SearchCode(in, workspaceRoot, opts.SearchMaxResults)
	if err != nil {
		return nil, err
	}
	return result, nil
}
