package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*mcp.ClientSession, func()) {
	t.Helper()

	db := entity.NewTestDB()
	tmpDB, err := db.DB()
	require.NoError(t, err)

	stubMCP := gostub.StubFunc(&getDB, db)

	srv := New()
	client, server := mcp.NewInMemoryTransports()

	go srv.srv.Run(context.Background(), server) //nolint:errcheck

	cc, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil).
		Connect(context.Background(), client, nil)
	require.NoError(t, err)

	return cc, func() {
		cc.Close()
		stubMCP.Reset()
		tmpDB.Close()
	}
}

func callTool(t *testing.T, cc *mcp.ClientSession, name string, args any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(args)
	require.NoError(t, err)
	res, err := cc.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: raw,
	})
	require.NoError(t, err)
	require.False(t, res.IsError, "tool error: %s", getText(t, res))
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(getText(t, res)), &out))
	return out
}

func callToolArray(t *testing.T, cc *mcp.ClientSession, name string, args any) []any {
	t.Helper()
	raw, err := json.Marshal(args)
	require.NoError(t, err)
	res, err := cc.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: raw,
	})
	require.NoError(t, err)
	require.False(t, res.IsError, "tool error: %s", getText(t, res))
	var out []any
	require.NoError(t, json.Unmarshal([]byte(getText(t, res)), &out))
	return out
}

func callToolRaw(t *testing.T, cc *mcp.ClientSession, name string, args any) (string, bool) {
	t.Helper()
	raw, err := json.Marshal(args)
	require.NoError(t, err)
	res, err := cc.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: raw,
	})
	require.NoError(t, err)
	return getText(t, res), res.IsError
}

func getText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	t.Fatal("no text content")
	return ""
}

func createFlag(t *testing.T, cc *mcp.ClientSession, key string) float64 {
	t.Helper()
	f := callTool(t, cc, "create_flag", map[string]any{"key": key})
	return f["id"].(float64)
}

func createFlagWithSegment(t *testing.T, cc *mcp.ClientSession) float64 {
	t.Helper()
	flagID := createFlag(t, cc, "flag_seg")
	callTool(t, cc, "create_segment", map[string]any{
		"flag_id": int(flagID), "rollout_percent": 100,
	})
	return flagID
}
