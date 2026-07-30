package handler

import (
	"net/http"
	"slices"
	"strings"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/mapper/entity_restapi/e2r"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/constraint"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/distribution"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/tag"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/variant"

	"github.com/go-openapi/runtime/middleware"
)

// readOnlyDenyMsg explains why write operations are rejected in eval-only mode.
const readOnlyDenyMsg = "Flagr is running in read-only (eval-only) mode: " +
	"flags are managed via the JSON source (FLAGR_DB_DBDRIVER=json_file/json_http), " +
	"write APIs are disabled"

// NewReadOnlyCRUD creates the CRUD implementation used in eval-only mode
// (json_file / json_http). Read operations are served from the EvalCache so
// the UI can display the current flag state without a database; write
// operations return 403 because the JSON source is the only write path.
func NewReadOnlyCRUD() CRUD {
	return &readOnlyCRUD{}
}

type readOnlyCRUD struct{}

// cachedFlag looks up a flag by numeric ID in the EvalCache. Returns nil when
// missing. Uses GetByFlagID (not GetByFlagKeyOrID) so a flag whose key is a
// numeric string can never shadow another flag's ID.
func cachedFlag(flagID int64) *entity.Flag {
	return GetEvalCache().GetByFlagID(util.SafeUint(flagID))
}

// cachedSegment looks up a segment by ID within a cached flag.
// Returns nil when the flag or the segment is missing.
func cachedSegment(flagID int64, segmentID int64) *entity.Segment {
	f := cachedFlag(flagID)
	if f == nil {
		return nil
	}
	for i := range f.Segments {
		if f.Segments[i].ID == uint(segmentID) {
			return &f.Segments[i]
		}
	}
	return nil
}

// filterFlags applies the FindFlags query filters to the cached flags,
// mirroring the DB-backed semantics: exact match on key/description,
// case-insensitive substring on description_like, ANY on comma-separated
// tags, then offset/limit.
func filterFlags(fs []entity.Flag, params flag.FindFlagsParams) []entity.Flag {
	var tags []string
	if params.Tags != nil {
		tags = strings.Split(*params.Tags, ",")
	}

	filtered := make([]entity.Flag, 0, len(fs))
	for _, f := range fs {
		if params.Enabled != nil && f.Enabled != *params.Enabled {
			continue
		}
		if params.Key != nil && f.Key != *params.Key {
			continue
		}
		if params.Description != nil && f.Description != *params.Description {
			continue
		}
		if params.DescriptionLike != nil &&
			!strings.Contains(strings.ToLower(f.Description), strings.ToLower(*params.DescriptionLike)) {
			continue
		}
		if tags != nil && !slices.ContainsFunc(f.Tags, func(t entity.Tag) bool {
			return slices.Contains(tags, t.Value)
		}) {
			continue
		}
		filtered = append(filtered, f)
	}

	if params.Offset != nil {
		filtered = filtered[min(max(int(*params.Offset), 0), len(filtered)):]
	}
	if params.Limit != nil && int(*params.Limit) < len(filtered) {
		filtered = filtered[:max(int(*params.Limit), 0)]
	}
	return filtered
}

func (c *readOnlyCRUD) FindFlags(params flag.FindFlagsParams) middleware.Responder {
	fs := []entity.Flag{}
	// The JSON source has no soft-deleted flags — deleted=true is always empty.
	if params.Deleted == nil || !*params.Deleted {
		fs = filterFlags(GetEvalCache().GetAllFlags(), params)
	}

	payload, err := e2rMapFlags(fs)
	if err != nil {
		return flag.NewFindFlagsDefault(http.StatusInternalServerError).WithPayload(
			ErrorMessage("cannot map flags. %s", err))
	}
	return flag.NewFindFlagsOK().WithPayload(payload)
}

func (c *readOnlyCRUD) GetFlag(params flag.GetFlagParams) middleware.Responder {
	f := cachedFlag(params.FlagID)
	if f == nil {
		return flag.NewGetFlagDefault(http.StatusNotFound).WithPayload(
			ErrorMessage("unable to find flag %v in the eval cache", params.FlagID))
	}
	payload, err := e2rMapFlag(f)
	if err != nil {
		return flag.NewGetFlagDefault(http.StatusInternalServerError).WithPayload(
			ErrorMessage("cannot map flag %v. %s", params.FlagID, err))
	}
	return flag.NewGetFlagOK().WithPayload(payload)
}

func (c *readOnlyCRUD) GetFlagSnapshots(params flag.GetFlagSnapshotsParams) middleware.Responder {
	// Change history for JSON-sourced flags lives in Git, not in Flagr.
	return flag.NewGetFlagSnapshotsOK().WithPayload([]*models.FlagSnapshot{})
}

func (c *readOnlyCRUD) GetFlagEntityTypes(params flag.GetFlagEntityTypesParams) middleware.Responder {
	// Entity types are recorded on evaluation into the DB; none exists here.
	return flag.NewGetFlagEntityTypesOK().WithPayload([]string{})
}

func (c *readOnlyCRUD) GetFlagSnapshotMaxID(params flag.GetFlagSnapshotMaxIDParams) middleware.Responder {
	// No flag_snapshot rows in eval-only mode. Serve the EvalCache content
	// fingerprint instead — clients (the UI list cache) only compare this
	// value for equality to decide whether to refetch.
	return flag.NewGetFlagSnapshotMaxIDOK().WithPayload(
		&models.FlagSnapshotMaxID{MaxID: GetEvalCache().ContentFingerprint()})
}

func (c *readOnlyCRUD) FindTags(params tag.FindTagsParams) middleware.Responder {
	ts := []entity.Tag{}
	if f := cachedFlag(params.FlagID); f != nil {
		ts = f.Tags
	}
	return tag.NewFindTagsOK().WithPayload(e2r.MapTags(ts))
}

func (c *readOnlyCRUD) FindAllTags(params tag.FindAllTagsParams) middleware.Responder {
	// In the JSON source, the same tag value on two flags is two entities —
	// dedupe by value to mirror the DB's many2many table.
	seen := map[string]bool{}
	ts := []entity.Tag{}
	for _, f := range GetEvalCache().GetAllFlags() {
		for _, t := range f.Tags {
			if seen[t.Value] {
				continue
			}
			if params.ValueLike != nil &&
				!strings.Contains(strings.ToLower(t.Value), strings.ToLower(*params.ValueLike)) {
				continue
			}
			seen[t.Value] = true
			ts = append(ts, t)
		}
	}

	if params.Offset != nil {
		ts = ts[min(max(int(*params.Offset), 0), len(ts)):]
	}
	if params.Limit != nil && int(*params.Limit) < len(ts) {
		ts = ts[:max(int(*params.Limit), 0)]
	}
	return tag.NewFindAllTagsOK().WithPayload(e2r.MapTags(ts))
}

func (c *readOnlyCRUD) FindSegments(params segment.FindSegmentsParams) middleware.Responder {
	ss := []entity.Segment{}
	if f := cachedFlag(params.FlagID); f != nil {
		ss = f.Segments
	}
	return segment.NewFindSegmentsOK().WithPayload(e2r.MapSegments(ss))
}

func (c *readOnlyCRUD) FindVariants(params variant.FindVariantsParams) middleware.Responder {
	vs := []entity.Variant{}
	if f := cachedFlag(params.FlagID); f != nil {
		vs = f.Variants
	}
	return variant.NewFindVariantsOK().WithPayload(e2r.MapVariants(vs))
}

func (c *readOnlyCRUD) FindConstraints(params constraint.FindConstraintsParams) middleware.Responder {
	cs := []entity.Constraint{}
	if s := cachedSegment(params.FlagID, params.SegmentID); s != nil {
		cs = s.Constraints
	}
	return constraint.NewFindConstraintsOK().WithPayload(e2r.MapConstraints(cs))
}

func (c *readOnlyCRUD) FindDistributions(params distribution.FindDistributionsParams) middleware.Responder {
	ds := []entity.Distribution{}
	if s := cachedSegment(params.FlagID, params.SegmentID); s != nil {
		ds = s.Distributions
	}
	return distribution.NewFindDistributionsOK().WithPayload(e2r.MapDistributions(ds))
}

// Write operations — all rejected with 403 in eval-only mode.

func (c *readOnlyCRUD) CreateFlag(flag.CreateFlagParams) middleware.Responder {
	return flag.NewCreateFlagDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) DuplicateFlag(flag.DuplicateFlagParams) middleware.Responder {
	return flag.NewDuplicateFlagDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) PutFlag(flag.PutFlagParams) middleware.Responder {
	return flag.NewPutFlagDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) DeleteFlag(flag.DeleteFlagParams) middleware.Responder {
	return flag.NewDeleteFlagDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) RestoreFlag(flag.RestoreFlagParams) middleware.Responder {
	return flag.NewRestoreFlagDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) SetFlagEnabledState(flag.SetFlagEnabledParams) middleware.Responder {
	return flag.NewSetFlagEnabledDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) CreateTag(tag.CreateTagParams) middleware.Responder {
	return tag.NewCreateTagDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) DeleteTag(tag.DeleteTagParams) middleware.Responder {
	return tag.NewDeleteTagDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) CreateSegment(segment.CreateSegmentParams) middleware.Responder {
	return segment.NewCreateSegmentDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) PutSegment(segment.PutSegmentParams) middleware.Responder {
	return segment.NewPutSegmentDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) DeleteSegment(segment.DeleteSegmentParams) middleware.Responder {
	return segment.NewDeleteSegmentDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) PutSegmentsReorder(segment.PutSegmentsReorderParams) middleware.Responder {
	return segment.NewPutSegmentsReorderDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) CreateConstraint(constraint.CreateConstraintParams) middleware.Responder {
	return constraint.NewCreateConstraintDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) PutConstraint(constraint.PutConstraintParams) middleware.Responder {
	return constraint.NewPutConstraintDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) DeleteConstraint(constraint.DeleteConstraintParams) middleware.Responder {
	return constraint.NewDeleteConstraintDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) PutDistributions(distribution.PutDistributionsParams) middleware.Responder {
	return distribution.NewPutDistributionsDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) CreateVariant(variant.CreateVariantParams) middleware.Responder {
	return variant.NewCreateVariantDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) PutVariant(variant.PutVariantParams) middleware.Responder {
	return variant.NewPutVariantDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}

func (c *readOnlyCRUD) DeleteVariant(variant.DeleteVariantParams) middleware.Responder {
	return variant.NewDeleteVariantDefault(http.StatusForbidden).WithPayload(ErrorMessage(readOnlyDenyMsg))
}
