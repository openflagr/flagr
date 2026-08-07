package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/notification"
	"gorm.io/gorm"
)

func (s *Server) registerFlagTools() {
	s.tool("create_flag", "Create a new feature flag.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"description": {"type": "string", "description": "Human-readable description"},
			"key": {"type": "string", "description": "Unique key. Auto-generated if empty."},
			"enabled": {"type": "boolean", "description": "Enable on creation (default false)"},
			"data_records_enabled": {"type": "boolean", "description": "Enable data recording"},
			"entity_type": {"type": "string", "description": "Entity type for data recording"},
			"notes": {"type": "string", "description": "Freeform notes"}
		}
	}`), s.handleCreateFlag)

	s.tool("get_flag", "Get a flag by ID with full detail (segments, variants, tags).", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer", "description": "Flag ID"}
		},
		"required": ["flag_id"]
	}`), s.handleGetFlag)

	s.tool("list_flags", "List flags with optional filters.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"enabled": {"type": "boolean", "description": "Filter by enabled state"},
			"description_like": {"type": "string", "description": "Substring match on description"},
			"tags": {"type": "string", "description": "Comma-separated tag values"},
			"preload": {"type": "boolean", "description": "Include segments, variants, tags"},
			"limit": {"type": "integer", "description": "Max results"},
			"offset": {"type": "integer", "description": "Offset for pagination"}
		}
	}`), s.handleListFlags)

	s.tool("update_flag", "Update a flag's properties.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer", "description": "Flag ID"},
			"description": {"type": "string"},
			"key": {"type": "string"},
			"enabled": {"type": "boolean"},
			"data_records_enabled": {"type": "boolean"},
			"entity_type": {"type": "string"},
			"notes": {"type": "string"}
		},
		"required": ["flag_id"]
	}`), s.handleUpdateFlag)

	s.tool("set_flag_enabled", "Enable or disable a flag.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"},
			"enabled": {"type": "boolean"}
		},
		"required": ["flag_id", "enabled"]
	}`), s.handleSetFlagEnabled)

	s.tool("delete_flag", "Soft-delete a flag.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer"}
		},
		"required": ["flag_id"]
	}`), s.handleDeleteFlag)

	s.tool("duplicate_flag", "Duplicate an existing flag with all its segments, variants, tags, constraints, and distributions.", json.RawMessage(`{
		"type": "object",
		"properties": {
			"flag_id": {"type": "integer", "description": "Source flag ID to duplicate"}
		},
		"required": ["flag_id"]
	}`), s.handleDuplicateFlag)
}

type createFlagInput struct {
	Key                string `json:"key"`
	Description        string `json:"description"`
	Enabled            bool   `json:"enabled"`
	DataRecordsEnabled bool   `json:"data_records_enabled"`
	EntityType         string `json:"entity_type"`
	Notes              string `json:"notes"`
}

func (s *Server) handleCreateFlag(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in createFlagInput
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	key, err := entity.CreateFlagKey(in.Key)
	if err != nil {
		return errResult("%v", err), nil
	}

	f := &entity.Flag{
		Key:                key,
		Description:        in.Description,
		Enabled:            in.Enabled,
		DataRecordsEnabled: in.DataRecordsEnabled,
		EntityType:         in.EntityType,
		Notes:              in.Notes,
		CreatedBy:          "mcp",
		UpdatedBy:          "mcp",
	}

	db := getDB()
	if err := db.Create(f).Error; err != nil {
		return errResult("failed to create flag: %v", err), nil
	}

	if err := entity.PreloadSegmentsVariantsTags(db).First(f, f.ID).Error; err != nil {
		return errResult("flag created but failed to load: %v", err), nil
	}

	return jsonText(mapFlag(f)), nil
}

func (s *Server) handleGetFlag(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID uint `json:"flag_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	f := &entity.Flag{}
	db := getDB()
	result := entity.PreloadSegmentsVariantsTags(db).First(f, in.FlagID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errResult("flag %d not found", in.FlagID), nil
	}
	if result.Error != nil {
		return errResult("failed to get flag: %v", result.Error), nil
	}

	return jsonText(mapFlag(f)), nil
}

type listFlagsInput struct {
	Enabled         *bool  `json:"enabled"`
	DescriptionLike string `json:"description_like"`
	Tags            string `json:"tags"`
	Preload         *bool  `json:"preload"`
	Limit           *int   `json:"limit"`
	Offset          *int   `json:"offset"`
}

func (s *Server) handleListFlags(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in listFlagsInput
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB().Unscoped()
	flags := []entity.Flag{}

	if in.Enabled != nil {
		db = db.Where("enabled = ?", *in.Enabled)
	}
	if in.DescriptionLike != "" {
		db = db.Where("lower(description) like ?", fmt.Sprintf("%%%s%%", in.DescriptionLike))
	}
	if in.Limit != nil {
		db = db.Limit(*in.Limit)
	}
	if in.Offset != nil {
		db = db.Offset(*in.Offset)
	}

	if in.Tags != "" {
		tagValues := strings.Split(in.Tags, ",")
		for i := range tagValues {
			tagValues[i] = strings.TrimSpace(tagValues[i])
		}
		tags := []entity.Tag{}
		db.Where("value in (?)", tagValues).Find(&tags)
		if err := db.Model(&tags).Group("flags.id").Association("Flags").Find(&flags); err != nil {
			return errResult("failed to query flags by tags: %v", err), nil
		}
	} else {
		if in.Preload != nil && *in.Preload {
			db = entity.PreloadSegmentsVariantsTags(db)
		} else {
			db = entity.PreloadFlagTags(db)
		}
		db.Where("deleted_at is null").Order("id").Find(&flags)
	}

	resultFlags := make([]any, len(flags))
	for i := range flags {
		resultFlags[i] = mapFlag(&flags[i])
	}

	return jsonText(map[string]any{
		"flags": resultFlags,
		"count": len(resultFlags),
	}), nil
}

type updateFlagInput struct {
	FlagID             uint   `json:"flag_id"`
	Description        string `json:"description"`
	Key                string `json:"key"`
	Enabled            *bool  `json:"enabled"`
	DataRecordsEnabled *bool  `json:"data_records_enabled"`
	EntityType         string `json:"entity_type"`
	Notes              string `json:"notes"`
}

func (s *Server) handleUpdateFlag(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in updateFlagInput
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	f := &entity.Flag{}
	if err := db.First(f, in.FlagID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errResult("flag %d not found", in.FlagID), nil
		}
		return errResult("failed to get flag: %v", err), nil
	}

	if in.Description != "" {
		f.Description = in.Description
	}
	if in.Key != "" {
		key, err := entity.CreateFlagKey(in.Key)
		if err != nil {
			return errResult("%v", err), nil
		}
		f.Key = key
	}
	if in.Enabled != nil {
		f.Enabled = *in.Enabled
	}
	if in.DataRecordsEnabled != nil {
		f.DataRecordsEnabled = *in.DataRecordsEnabled
	}
	if in.EntityType != "" {
		f.EntityType = in.EntityType
	}
	if in.Notes != "" {
		f.Notes = in.Notes
	}
	f.UpdatedBy = "mcp"

	if err := db.Save(f).Error; err != nil {
		return errResult("failed to update flag: %v", err), nil
	}
	if err := entity.PreloadSegmentsVariantsTags(db).First(f, f.ID).Error; err != nil {
		return errResult("updated but failed to reload: %v", err), nil
	}

	return jsonText(mapFlag(f)), nil
}

func (s *Server) handleSetFlagEnabled(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID  uint `json:"flag_id"`
		Enabled bool `json:"enabled"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	f := &entity.Flag{}
	if err := db.First(f, in.FlagID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errResult("flag %d not found", in.FlagID), nil
		}
		return errResult("failed to get flag: %v", err), nil
	}

	f.Enabled = in.Enabled
	f.UpdatedBy = "mcp"
	if err := db.Save(f).Error; err != nil {
		return errResult("failed to update flag: %v", err), nil
	}
	if err := entity.PreloadSegmentsVariantsTags(db).First(f, f.ID).Error; err != nil {
		return errResult("updated but failed to reload: %v", err), nil
	}

	return jsonText(mapFlag(f)), nil
}

func (s *Server) handleDeleteFlag(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID uint `json:"flag_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	f := &entity.Flag{}
	if err := db.First(f, in.FlagID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errResult("flag %d not found", in.FlagID), nil
		}
		return errResult("failed to get flag: %v", err), nil
	}

	if err := db.Delete(&entity.Flag{}, in.FlagID).Error; err != nil {
		return errResult("failed to delete flag: %v", err), nil
	}

	return jsonText(map[string]any{
		"message": fmt.Sprintf("flag %d deleted", in.FlagID),
	}), nil
}

func (s *Server) handleDuplicateFlag(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var in struct {
		FlagID uint `json:"flag_id"`
	}
	if err := decodeArgs(req.Params.Arguments, &in); err != nil {
		return errResult("invalid arguments: %v", err), nil
	}

	db := getDB()
	src := &entity.Flag{}
	if err := entity.PreloadSegmentsVariantsTags(db).First(src, in.FlagID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errResult("flag %d not found", in.FlagID), nil
		}
		return errResult("failed to load source flag: %v", err), nil
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	newFlag := &entity.Flag{
		Key:                src.Key + "_copy",
		Description:        src.Description,
		Enabled:            false,
		DataRecordsEnabled: src.DataRecordsEnabled,
		EntityType:         src.EntityType,
		Notes:              src.Notes,
		CreatedBy:          "mcp",
		UpdatedBy:          "mcp",
	}
	if err := tx.Create(newFlag).Error; err != nil {
		tx.Rollback()
		return errResult("failed to create duplicate flag: %v", err), nil
	}

	// Map source variant IDs to new variant IDs.
	variantIDMap := map[uint]uint{}
	for _, v := range src.Variants {
		nv := entity.Variant{
			FlagID:     newFlag.ID,
			Key:        v.Key,
			Attachment: v.Attachment,
		}
		if err := tx.Create(&nv).Error; err != nil {
			tx.Rollback()
			return errResult("failed to duplicate variant: %v", err), nil
		}
		variantIDMap[v.ID] = nv.ID
	}

	for _, seg := range src.Segments {
		ns := entity.Segment{
			FlagID:         newFlag.ID,
			Description:    seg.Description,
			Rank:           seg.Rank,
			RolloutPercent: seg.RolloutPercent,
		}
		if err := tx.Create(&ns).Error; err != nil {
			tx.Rollback()
			return errResult("failed to duplicate segment: %v", err), nil
		}

		for _, c := range seg.Constraints {
			nc := entity.Constraint{
				SegmentID: ns.ID,
				Property:  c.Property,
				Operator:  c.Operator,
				Value:     c.Value,
			}
			if err := tx.Create(&nc).Error; err != nil {
				tx.Rollback()
				return errResult("failed to duplicate constraint: %v", err), nil
			}
		}

		for _, d := range seg.Distributions {
			newVariantID := variantIDMap[d.VariantID]
			nd := entity.Distribution{
				SegmentID:  ns.ID,
				VariantID:  newVariantID,
				VariantKey: d.VariantKey,
				Percent:    d.Percent,
			}
			if err := tx.Create(&nd).Error; err != nil {
				tx.Rollback()
				return errResult("failed to duplicate distribution: %v", err), nil
			}
		}
	}

	for _, t := range src.Tags {
		if err := entity.AppendTagValueToFlag(tx, newFlag.ID, t.Value); err != nil {
			tx.Rollback()
			return errResult("failed to duplicate tag: %v", err), nil
		}
	}

	// Write snapshot.
	snap, err := entity.WriteFlagSnapshotTx(tx, newFlag.ID, "mcp")
	if err != nil {
		tx.Rollback()
		return errResult("failed to write snapshot: %v", err), nil
	}
	if err := tx.Commit().Error; err != nil {
		return errResult("failed to commit duplicate: %v", err), nil
	}
	snap.NotifyAfterCommit(newFlag.ID, "mcp", notification.OperationCreate, notification.ComponentFlag, newFlag.ID, newFlag.Key)

	if err := entity.PreloadSegmentsVariantsTags(db).First(newFlag, newFlag.ID).Error; err != nil {
		return errResult("duplicated but failed to load: %v", err), nil
	}

	return jsonText(mapFlag(newFlag)), nil
}

// mapFlag maps an entity.Flag to a JSON-friendly map.
func mapFlag(f *entity.Flag) map[string]any {
	segments := make([]any, len(f.Segments))
	for i, seg := range f.Segments {
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
		segments[i] = map[string]any{
			"id":              seg.ID,
			"flag_id":         seg.FlagID,
			"description":     seg.Description,
			"rank":            seg.Rank,
			"rollout_percent": seg.RolloutPercent,
			"constraints":     constraints,
			"distributions":   distributions,
		}
	}

	variants := make([]any, len(f.Variants))
	for i, v := range f.Variants {
		variants[i] = map[string]any{
			"id":         v.ID,
			"flag_id":    v.FlagID,
			"key":        v.Key,
			"attachment": v.Attachment,
		}
	}

	tags := make([]any, len(f.Tags))
	for i, t := range f.Tags {
		tags[i] = map[string]any{
			"id":    t.ID,
			"value": t.Value,
		}
	}

	return map[string]any{
		"id":                   f.ID,
		"key":                  f.Key,
		"description":          f.Description,
		"enabled":              f.Enabled,
		"data_records_enabled": f.DataRecordsEnabled,
		"entity_type":          f.EntityType,
		"notes":                f.Notes,
		"created_by":           f.CreatedBy,
		"updated_by":           f.UpdatedBy,
		"created_at":           f.CreatedAt,
		"updated_at":           f.UpdatedAt,
		"segments":             segments,
		"variants":             variants,
		"tags":                 tags,
	}
}
