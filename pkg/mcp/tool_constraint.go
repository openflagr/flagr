package mcp

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openflagr/flagr/pkg/entity"
	"gorm.io/gorm"
)

func (s *Server) registerConstraintTools() {
	s.tool("create_constraint", `Create a constraint on a segment. Constraints determine which entities match the segment.

Supported operators: EQ, NEQ, LT, LTE, GT, GTE, EREG, NEREG, IN, NOTIN, CONTAINS, NOTCONTAINS.

Value format: strings must be quoted ("foo"), arrays as ("a","b"), regex as ("pattern").`, json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"segment_id": {"type": "integer"},
			"property": {"type": "string", "description": "Entity context key to match"},
			"operator": {"type": "string", "description": "Comparison operator"},
			"value": {"type": "string", "description": "Value to compare against"}
		},
		"required": ["flag_id", "segment_id", "property", "operator", "value"]
	}`), s.handleCreateConstraint)

	s.tool("list_constraints", "List constraints for a segment.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"segment_id": {"type": "integer"}
		},
		"required": ["flag_id", "segment_id"]
	}`), s.handleListConstraints)

	s.tool("update_constraint", "Update a constraint's property, operator, or value.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"constraint_id": {"type": "integer"},
			"property": {"type": "string"},
			"operator": {"type": "string"},
			"value": {"type": "string"}
		},
		"required": ["flag_id", "constraint_id"]
	}`), s.handleUpdateConstraint)

	s.tool("delete_constraint", "Delete a constraint from a segment.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"constraint_id": {"type": "integer"}
		},
		"required": ["flag_id", "constraint_id"]
	}`), s.handleDeleteConstraint)

	s.tool("list_distributions", "List distributions for a segment.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"segment_id": {"type": "integer"}
		},
		"required": ["flag_id", "segment_id"]
	}`), s.handleListDistributions)

	s.tool("set_distributions", `Overwrite all distributions for a segment. Each entry maps a variant to a rollout percent. Percents must sum to 100.`, json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"segment_id": {"type": "integer"},
			"distributions": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"variant_id": {"type": "integer"},
						"variant_key": {"type": "string"},
						"percent": {"type": "integer", "description": "0-100"}
					},
					"required": ["variant_id", "variant_key", "percent"]
				}
			}
		},
		"required": ["flag_id", "segment_id", "distributions"]
	}`), s.handleSetDistributions)
}

func (s *Server) handleCreateConstraint(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID    uint   `json:"flag_id"`
		SegmentID uint   `json:"segment_id"`
		Property  string `json:"property"`
		Operator  string `json:"operator"`
		Value     string `json:"value"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	c := &entity.Constraint{
		SegmentID: in.SegmentID,
		Property:  in.Property,
		Operator:  in.Operator,
		Value:     in.Value,
	}
	if err := c.Validate(); err != nil {
		return errResult("invalid constraint: %v", err), nil
	}

	db := getDB()
	if err := db.Create(c).Error; err != nil {
		return errResult("failed to create constraint: %v", err), nil
	}

	return jsonText(mapConstraint(c)), nil
}

func (s *Server) handleListConstraints(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		SegmentID uint `json:"segment_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	constraints := []entity.Constraint{}
	if err := db.Order("created_at").Where("segment_id = ?", in.SegmentID).Find(&constraints).Error; err != nil {
		return errResult("failed to list constraints: %v", err), nil
	}

	result := make([]any, len(constraints))
	for i := range constraints {
		result[i] = mapConstraint(&constraints[i])
	}
	return jsonText(result), nil
}

func (s *Server) handleUpdateConstraint(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		ConstraintID uint   `json:"constraint_id"`
		Property     string `json:"property"`
		Operator     string `json:"operator"`
		Value        string `json:"value"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	c := &entity.Constraint{}
	if err := db.First(c, in.ConstraintID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errResult("constraint %d not found", in.ConstraintID), nil
		}
		return errResult("failed to get constraint: %v", err), nil
	}

	if in.Property != "" {
		c.Property = in.Property
	}
	if in.Operator != "" {
		c.Operator = in.Operator
	}
	if in.Value != "" {
		c.Value = in.Value
	}

	if err := c.Validate(); err != nil {
		return errResult("invalid constraint: %v", err), nil
	}
	if err := db.Save(c).Error; err != nil {
		return errResult("failed to update constraint: %v", err), nil
	}

	return jsonText(mapConstraint(c)), nil
}

func (s *Server) handleDeleteConstraint(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		ConstraintID uint `json:"constraint_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	if err := db.Delete(&entity.Constraint{}, in.ConstraintID).Error; err != nil {
		return errResult("failed to delete constraint: %v", err), nil
	}

	return jsonText(map[string]any{"message": "constraint deleted"}), nil
}

func (s *Server) handleListDistributions(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		SegmentID uint `json:"segment_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	distributions := []entity.Distribution{}
	if err := db.Order("variant_id").Where("segment_id = ?", in.SegmentID).Find(&distributions).Error; err != nil {
		return errResult("failed to list distributions: %v", err), nil
	}

	result := make([]any, len(distributions))
	for i := range distributions {
		result[i] = mapDistribution(&distributions[i])
	}
	return jsonText(result), nil
}

type setDistributionsInput struct {
	FlagID        uint `json:"flag_id"`
	SegmentID     uint `json:"segment_id"`
	Distributions []struct {
		VariantID  uint   `json:"variant_id"`
		VariantKey string `json:"variant_key"`
		Percent    uint   `json:"percent"`
	} `json:"distributions"`
}

func (s *Server) handleSetDistributions(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in setDistributionsInput
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()

	// Delete existing distributions.
	if err := db.Where("segment_id = ?", in.SegmentID).Delete(&entity.Distribution{}).Error; err != nil {
		return errResult("failed to clear distributions: %v", err), nil
	}

	// Create new ones.
	for _, d := range in.Distributions {
		dist := entity.Distribution{
			SegmentID:  in.SegmentID,
			VariantID:  d.VariantID,
			VariantKey: d.VariantKey,
			Percent:    d.Percent,
		}
		if err := db.Create(&dist).Error; err != nil {
			return errResult("failed to create distribution: %v", err), nil
		}
	}

	return jsonText(map[string]any{"message": "distributions updated"}), nil
}

func mapConstraint(c *entity.Constraint) map[string]any {
	return map[string]any{
		"id":         c.ID,
		"segment_id": c.SegmentID,
		"property":   c.Property,
		"operator":   c.Operator,
		"value":      c.Value,
	}
}

func mapDistribution(d *entity.Distribution) map[string]any {
	return map[string]any{
		"id":          d.ID,
		"segment_id":  d.SegmentID,
		"variant_id":  d.VariantID,
		"variant_key": d.VariantKey,
		"percent":     d.Percent,
	}
}
