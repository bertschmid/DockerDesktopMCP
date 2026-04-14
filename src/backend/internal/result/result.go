// Package result defines the shared MCP tool result types used by both the
// docker and mcp packages. Keeping them here breaks the import cycle.
package result

// CallToolResult is the MCP tools/call response payload.
type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content is a single piece of content inside a CallToolResult.
type Content struct {
	Type string `json:"type"` // "text" | "image" | "resource"
	Text string `json:"text,omitempty"`
}

// Text creates a successful text tool result.
func Text(text string) *CallToolResult {
	return &CallToolResult{Content: []Content{{Type: "text", Text: text}}}
}

// Error creates an error tool result.
func Error(msg string) *CallToolResult {
	return &CallToolResult{
		IsError: true,
		Content: []Content{{Type: "text", Text: msg}},
	}
}
