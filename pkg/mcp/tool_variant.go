package mcp

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/util"
	"gorm.io/gorm"
)

func (s *Server) registerVariantTools() {
	s.tool("create_variant", "Create a variant on a flag. Variants represent the possible outcomes of a flag evaluation.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"key": {"type": "string", "description": "Variant key (unique per flag)"},
			"attachment": {"type": "object", "description": "Dynamic configuration (key/value pairs)"}
		},
		"required": ["flag_id", "key"]
	}`), s.handleCreateVariant)

	s.tool("list_variants", "List all variants for a flag.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"}
		},
		"required": ["flag_id"]
	}`), s.handleListVariants)

	s.tool("update_variant", "Update a variant's key and attachment.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"variant_id": {"type": "integer"},
			"key": {"type": "string"},
			"attachment": {"type": "object"}
		},
		"required": ["flag_id", "variant_id"]
	}`), s.handleUpdateVariant)

	s.tool("delete_variant", "Delete a variant from a flag.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"variant_id": {"type": "integer"}
		},
		"required": ["flag_id", "variant_id"]
	}`), s.handleDeleteVariant)
}

type createVariantInput struct {
	FlagID     uint           `json:"flag_id"`
	Key        string         `json:"key"`
	Attachment map[string]any `json:"attachment"`
}

func (s *Server) handleCreateVariant(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in createVariantInput
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	v := &entity.Variant{
		FlagID:     in.FlagID,
		Key:        util.SafeString(in.Key),
		Attachment: entity.Attachment(in.Attachment),
	}
	if err := v.Validate(); err != nil {
		return errResult("%v", err), nil
	}

	db := getDB()
	if err := db.Create(v).Error; err != nil {
		return errResult("failed to create variant: %v", err), nil
	}

	return jsonText(mapVariant(v)), nil
}

func (s *Server) handleListVariants(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID uint `json:"flag_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	variants := []entity.Variant{}
	if err := db.Order("id").Where("flag_id = ?", in.FlagID).Find(&variants).Error; err != nil {
		return errResult("failed to list variants: %v", err), nil
	}

	result := make([]any, len(variants))
	for i := range variants {
		result[i] = mapVariant(&variants[i])
	}
	return jsonText(result), nil
}

type updateVariantInput struct {
	FlagID     uint           `json:"flag_id"`
	VariantID  uint           `json:"variant_id"`
	Key        string         `json:"key"`
	Attachment map[string]any `json:"attachment"`
}

func (s *Server) handleUpdateVariant(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in updateVariantInput
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	v := &entity.Variant{}
	if err := db.First(v, in.VariantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errResult("variant %d not found", in.VariantID), nil
		}
		return errResult("failed to get variant: %v", err), nil
	}

	if in.Key != "" {
		v.Key = util.SafeString(in.Key)
	}
	if in.Attachment != nil {
		v.Attachment = entity.Attachment(in.Attachment)
	}

	if err := v.Validate(); err != nil {
		return errResult("%v", err), nil
	}
	if err := db.Save(v).Error; err != nil {
		return errResult("failed to update variant: %v", err), nil
	}

	return jsonText(mapVariant(v)), nil
}

func (s *Server) handleDeleteVariant(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID    uint `json:"flag_id"`
		VariantID uint `json:"variant_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	if err := db.Delete(&entity.Variant{}, in.VariantID).Error; err != nil {
		return errResult("failed to delete variant: %v", err), nil
	}

	return jsonText(map[string]any{"message": "variant deleted"}), nil
}

func mapVariant(v *entity.Variant) map[string]any {
	return map[string]any{
		"id":         v.ID,
		"flag_id":    v.FlagID,
		"key":        v.Key,
		"attachment": v.Attachment,
	}
}
