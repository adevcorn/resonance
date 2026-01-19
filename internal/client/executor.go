package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adevcorn/ensemble/internal/client/tools"
	"github.com/adevcorn/ensemble/internal/protocol"
)

// Executor executes tools locally
type Executor struct {
	projectPath string
	checker     *Checker
}

// NewExecutor creates a new local tool executor
func NewExecutor(projectPath string, checker *Checker) *Executor {
	return &Executor{
		projectPath: projectPath,
		checker:     checker,
	}
}

// Execute executes a tool call
func (e *Executor) Execute(ctx context.Context, call protocol.ToolCall) (protocol.ToolResult, error) {
	result := protocol.ToolResult{
		CallID: call.ID,
	}

	// Route to appropriate tool handler
	switch call.ToolName {
	case "read_file":
		output, err := e.executeReadFile(ctx, call.Arguments)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Result = output
		}

	case "write_file":
		output, err := e.executeWriteFile(ctx, call.Arguments)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Result = output
		}

	case "list_directory":
		output, err := e.executeListDirectory(ctx, call.Arguments)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Result = output
		}

	case "execute_command":
		output, err := e.executeCommand(ctx, call.Arguments)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Result = output
		}

	default:
		result.Error = fmt.Sprintf("unknown tool: %s", call.ToolName)
	}

	return result, nil
}

// executeReadFile executes the read_file tool
func (e *Executor) executeReadFile(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input tools.ReadFileInput
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Check permissions
	if err := e.checker.CheckFilePath(input.Path); err != nil {
		return nil, err
	}

	// Execute tool
	output, err := tools.ReadFile(ctx, e.projectPath, input)
	if err != nil {
		return nil, err
	}

	return json.Marshal(output)
}

// executeWriteFile executes the write_file tool
func (e *Executor) executeWriteFile(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input tools.WriteFileInput
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Check permissions
	if err := e.checker.CheckFilePath(input.Path); err != nil {
		return nil, err
	}

	// Execute tool
	output, err := tools.WriteFile(ctx, e.projectPath, input)
	if err != nil {
		return nil, err
	}

	return json.Marshal(output)
}

// executeListDirectory executes the list_directory tool
func (e *Executor) executeListDirectory(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input tools.ListDirectoryInput
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Check permissions
	if err := e.checker.CheckFilePath(input.Path); err != nil {
		return nil, err
	}

	// Execute tool
	output, err := tools.ListDirectory(ctx, e.projectPath, input)
	if err != nil {
		return nil, err
	}

	return json.Marshal(output)
}

// executeCommand executes the execute_command tool
func (e *Executor) executeCommand(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input tools.ExecuteCommandInput
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	// Check permissions
	if err := e.checker.CheckCommand(input.Command, input.Args); err != nil {
		return nil, err
	}

	// Execute tool
	output, err := tools.ExecuteCommand(ctx, e.projectPath, input)
	if err != nil {
		return nil, err
	}

	return json.Marshal(output)
}
