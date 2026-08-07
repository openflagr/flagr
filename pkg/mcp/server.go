package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/handler"
	"github.com/sirupsen/logrus"
)

// getDB is the database accessor. Overridable in tests.
var getDB = entity.GetDB

// Server is the Flagr MCP server.
type Server struct {
	srv *mcp.Server
}

// New creates a new MCP server with all Flagr tools registered.
func New() *Server {
	s := &Server{
		srv: mcp.NewServer(
			&mcp.Implementation{
				Name:    "flagr-mcp",
				Version: "0.1.0",
			},
			nil,
		),
	}
	s.registerTools()
	return s
}

// Run starts the MCP server over stdio. Blocks until ctx is cancelled or
// stdin is closed. Call as a goroutine.
func (s *Server) Run(ctx context.Context) error {
	logrus.Info("mcp server starting (stdio transport)")
	t := &mcp.IOTransport{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}
	err := s.srv.Run(ctx, t)
	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("mcp server stopped: %w", err)
	}
	logrus.Info("mcp server stopped")
	return nil
}

// Handler returns an http.Handler serving the MCP server over Streamable HTTP.
// Mount at any path (e.g. /mcp).
func (s *Server) Handler() http.Handler {
	return mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return s.srv },
		&mcp.StreamableHTTPOptions{
			JSONResponse: true,
		},
	)
}

// jsonText returns a CallToolResult with indented JSON text content.
func jsonText(v any) *mcp.CallToolResult {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("marshal error: %v", err)}},
			IsError: true,
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
	}
}

// errResult returns a CallToolResult signalling an error.
func errResult(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}},
		IsError: true,
	}
}

// MCPEnabled returns true when FLAGR_MCP_ENABLED=true.
func MCPEnabled() bool {
	return config.Config.MCPEnabled
}

// StartEvalCache initializes and starts the in-memory evaluation cache.
// Must be called before any evaluate_flag tool calls.
func StartEvalCache() {
	handler.GetEvalCache().Start()
}

// tool adds a tool to the server (convenience wrapper).
func (s *Server) tool(name, desc string, schema json.RawMessage, h mcp.ToolHandler) {
	s.srv.AddTool(&mcp.Tool{
		Name:        name,
		Description: desc,
		InputSchema: schema,
	}, h)
}

// decodeArgs unmarshals the tool call arguments. The MCP transport may
// base64-encode json.RawMessage fields, so we try raw first, then base64.
func decodeArgs(raw json.RawMessage, v any) error {
	if err := json.Unmarshal(raw, v); err == nil {
		return nil
	}
	// The raw message may be a JSON string containing base64 data.
	// Unquote it first.
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
			return json.Unmarshal(decoded, v)
		}
	}
	return fmt.Errorf("invalid arguments")
}
