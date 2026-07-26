package mcp

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openflagr/flagr/pkg/entity"
)

func (s *Server) registerTagTools() {
	s.tool("list_tags", "List tags on a flag.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"}
		},
		"required": ["flag_id"]
	}`), s.handleListTags)

	s.tool("add_tag", "Add a tag to a flag. Tags are used to filter flags in batch evaluations.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"value": {"type": "string", "description": "Tag value (max 64 chars, unique per flag)"}
		},
		"required": ["flag_id", "value"]
	}`), s.handleAddTag)

	s.tool("delete_tag", "Remove a tag from a flag.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"tag_id": {"type": "integer"}
		},
		"required": ["flag_id", "tag_id"]
	}`), s.handleDeleteTag)

	s.tool("list_all_tags", "List all tags across all flags.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"value_like": {"type": "string", "description": "Substring filter"},
			"limit": {"type": "integer"},
			"offset": {"type": "integer"}
		}
	}`), s.handleListAllTags)
}

func (s *Server) handleListTags(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID uint `json:"flag_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	f := &entity.Flag{}
	f.ID = in.FlagID
	tags := []entity.Tag{}
	if err := db.Model(f).Association("Tags").Find(&tags); err != nil {
		return errResult("failed to list tags: %v", err), nil
	}

	result := make([]any, len(tags))
	for i := range tags {
		result[i] = map[string]any{"id": tags[i].ID, "value": tags[i].Value}
	}
	return jsonText(result), nil
}

func (s *Server) handleAddTag(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID uint   `json:"flag_id"`
		Value  string `json:"value"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	if err := entity.AppendTagValueToFlag(db, in.FlagID, in.Value); err != nil {
		return errResult("failed to add tag: %v", err), nil
	}

	t := &entity.Tag{}
	if err := db.Where("value = ?", in.Value).First(t).Error; err != nil {
		return errResult("tag added but failed to load: %v", err), nil
	}

	return jsonText(map[string]any{"id": t.ID, "value": t.Value}), nil
}

func (s *Server) handleDeleteTag(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID uint `json:"flag_id"`
		TagID  uint `json:"tag_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	f := &entity.Flag{}
	f.ID = in.FlagID
	t := &entity.Tag{}
	t.ID = in.TagID
	if err := db.Model(f).Association("Tags").Delete(t); err != nil {
		return errResult("failed to delete tag: %v", err), nil
	}

	return jsonText(map[string]any{"message": "tag removed"}), nil
}

func (s *Server) handleListAllTags(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		ValueLike string `json:"value_like"`
		Limit     *int   `json:"limit"`
		Offset    *int   `json:"offset"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	tags := []entity.Tag{}

	if in.ValueLike != "" {
		db = db.Where("lower(value) like ?", "%"+in.ValueLike+"%")
	}
	if in.Limit != nil {
		db = db.Limit(*in.Limit)
	}
	if in.Offset != nil {
		db = db.Offset(*in.Offset)
	}

	if err := db.Find(&tags).Error; err != nil {
		return errResult("failed to list tags: %v", err), nil
	}

	result := make([]any, len(tags))
	for i := range tags {
		result[i] = map[string]any{"id": tags[i].ID, "value": tags[i].Value}
	}
	return jsonText(result), nil
}
