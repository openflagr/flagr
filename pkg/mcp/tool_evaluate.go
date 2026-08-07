package mcp

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openflagr/flagr/pkg/handler"
	"github.com/openflagr/flagr/swagger_gen/models"
)

func (s *Server) registerEvaluateTools() {
	s.tool("evaluate_flag", `Evaluate a flag for a given entity. Returns the matching variant.

Provide either flag_id or flag_key to identify the flag, and entity_id plus optional entity_context for the evaluation.

Example entity_context for a user: {"country": "US", "plan": "pro"}`, json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"flag_key": {"type": "string"},
			"entity_id": {"type": "string", "description": "Unique entity identifier"},
			"entity_type": {"type": "string"},
			"entity_context": {"type": "object", "description": "Key/value context for constraint evaluation"}
		},
		"required": ["entity_id"]
	}`), s.handleEvaluateFlag)

	s.tool("health", "Check Flagr server health.", json.RawMessage(`{"type": "object"}`), s.handleHealth)
}

type evaluateInput struct {
	FlagID        uint           `json:"flag_id"`
	FlagKey       string         `json:"flag_key"`
	EntityID      string         `json:"entity_id"`
	EntityType    string         `json:"entity_type"`
	EntityContext map[string]any `json:"entity_context"`
}

func (s *Server) handleEvaluateFlag(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in evaluateInput
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	if in.EntityID == "" {
		return errResult("entity_id is required"), nil
	}

	ec := models.EvalContext{
		EntityID:      in.EntityID,
		EntityType:    in.EntityType,
		EntityContext: in.EntityContext,
		EnableDebug:   true,
	}

	if in.FlagID != 0 {
		ec.FlagID = int64(in.FlagID)
	}
	if in.FlagKey != "" {
		ec.FlagKey = in.FlagKey
	}

	result := handler.EvalFlag(ec)
	return jsonText(mapEvalResult(result)), nil
}

func (s *Server) handleHealth(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return jsonText(map[string]any{"status": "OK"}), nil
}

func mapEvalResult(r *models.EvalResult) map[string]any {
	if r == nil {
		return map[string]any{"error": "nil result"}
	}

	out := map[string]any{
		"flag_id":              r.FlagID,
		"flag_key":             r.FlagKey,
		"flag_snapshot_id":     r.FlagSnapshotID,
		"variant_id":           r.VariantID,
		"variant_key":          r.VariantKey,
		"segment_id":           r.SegmentID,
		"timestamp":            r.Timestamp,
		"data_records_enabled": r.DataRecordsEnabled,
	}

	if r.VariantAttachment != nil {
		out["variant_attachment"] = r.VariantAttachment
	}
	if r.FlagTags != nil {
		out["flag_tags"] = r.FlagTags
	}

	if r.EvalDebugLog != nil {
		out["debug_msg"] = r.EvalDebugLog.Msg
	}

	return out
}
