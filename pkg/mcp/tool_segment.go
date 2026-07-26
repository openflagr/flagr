package mcp

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openflagr/flagr/pkg/entity"
	"gorm.io/gorm"
)

func (s *Server) registerSegmentTools() {
	s.tool("create_segment", "Create a segment on a flag. Segments evaluate constraints in rank order to determine rollout.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"rollout_percent": {"type": "integer", "description": "0-100"},
			"description": {"type": "string"}
		},
		"required": ["flag_id", "rollout_percent"]
	}`), s.handleCreateSegment)

	s.tool("list_segments", "List segments for a flag, ordered by rank.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"}
		},
		"required": ["flag_id"]
	}`), s.handleListSegments)

	s.tool("update_segment", "Update a segment's rollout percent and description.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"segment_id": {"type": "integer"},
			"rollout_percent": {"type": "integer", "description": "0-100"},
			"description": {"type": "string"}
		},
		"required": ["flag_id", "segment_id"]
	}`), s.handleUpdateSegment)

	s.tool("delete_segment", "Delete a segment from a flag.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"segment_id": {"type": "integer"}
		},
		"required": ["flag_id", "segment_id"]
	}`), s.handleDeleteSegment)
}

type createSegmentInput struct {
	FlagID         uint   `json:"flag_id"`
	RolloutPercent uint   `json:"rollout_percent"`
	Description    string `json:"description"`
}

func (s *Server) handleCreateSegment(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in createSegmentInput
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	seg := &entity.Segment{
		FlagID:         in.FlagID,
		RolloutPercent: in.RolloutPercent,
		Description:    in.Description,
		Rank:           entity.SegmentDefaultRank,
	}

	if err := db.Create(seg).Error; err != nil {
		return errResult("failed to create segment: %v", err), nil
	}

	return jsonText(mapSegment(seg)), nil
}

func (s *Server) handleListSegments(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID uint `json:"flag_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	segments := []entity.Segment{}
	if err := entity.PreloadConstraintsDistribution(db).
		Order("segments.rank").Order("segments.id").
		Where("flag_id = ?", in.FlagID).
		Find(&segments).Error; err != nil {
		return errResult("failed to list segments: %v", err), nil
	}

	result := make([]any, len(segments))
	for i := range segments {
		result[i] = mapSegment(&segments[i])
	}
	return jsonText(result), nil
}

func (s *Server) handleUpdateSegment(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID         uint   `json:"flag_id"`
		SegmentID      uint   `json:"segment_id"`
		RolloutPercent uint   `json:"rollout_percent"`
		Description    string `json:"description"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	seg := &entity.Segment{}
	if err := entity.PreloadConstraintsDistribution(db).First(seg, in.SegmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errResult("segment %d not found", in.SegmentID), nil
		}
		return errResult("failed to get segment: %v", err), nil
	}

	seg.RolloutPercent = in.RolloutPercent
	if in.Description != "" {
		seg.Description = in.Description
	}

	if err := db.Save(seg).Error; err != nil {
		return errResult("failed to update segment: %v", err), nil
	}

	return jsonText(mapSegment(seg)), nil
}

func (s *Server) handleDeleteSegment(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID    uint `json:"flag_id"`
		SegmentID uint `json:"segment_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	if err := db.Delete(&entity.Segment{}, in.SegmentID).Error; err != nil {
		return errResult("failed to delete segment: %v", err), nil
	}

	return jsonText(map[string]any{"message": "segment deleted"}), nil
}

func mapSegment(seg *entity.Segment) map[string]any {
	constraints := make([]any, len(seg.Constraints))
	for j, c := range seg.Constraints {
		constraints[j] = map[string]any{
			"id":         c.ID,
			"segment_id": c.SegmentID,
			"property":   c.Property,
			"operator":   c.Operator,
			"value":      c.Value,
		}
	}
	distributions := make([]any, len(seg.Distributions))
	for j, d := range seg.Distributions {
		distributions[j] = map[string]any{
			"id":          d.ID,
			"segment_id":  d.SegmentID,
			"variant_id":  d.VariantID,
			"variant_key": d.VariantKey,
			"percent":     d.Percent,
		}
	}
	return map[string]any{
		"id":              seg.ID,
		"flag_id":         seg.FlagID,
		"description":     seg.Description,
		"rank":            seg.Rank,
		"rollout_percent": seg.RolloutPercent,
		"constraints":     constraints,
		"distributions":   distributions,
	}
}
