package entity

import (
	"fmt"
	"sort"

	"gorm.io/gorm"
)

// AppendTagValueToFlag finds or creates a tag by value and associates it with the flag.
func AppendTagValueToFlag(tx *gorm.DB, flagID uint, value string) error {
	t := &Tag{Value: value}
	if err := tx.Where("value = ?", value).FirstOrCreate(t).Error; err != nil {
		return err
	}
	flagRef := &Flag{}
	flagRef.ID = flagID
	return tx.Model(flagRef).Association("Tags").Append(t)
}

// SimpleBooleanFlagTemplate is the built-in starter template (variant "on", 100% rollout).
// Only nested Variants/Segments are set; ApplyFlagTemplate ignores flag-level scalars and IDs.
func SimpleBooleanFlagTemplate() Flag {
	return Flag{
		Variants: []Variant{{Key: "on"}},
		Segments: []Segment{{
			RolloutPercent: 100,
			Rank:           SegmentDefaultRank,
			Distributions:  []Distribution{{VariantKey: "on", Percent: 100}},
		}},
	}
}

// SourceFlagTemplate builds a template from an existing flag for cloning onto another flag.
// Variant keys and distribution variant keys are preserved; entity IDs and flag scalars are omitted.
func SourceFlagTemplate(source *Flag) Flag {
	out := Flag{
		Variants: make([]Variant, 0, len(source.Variants)),
		Tags:     make([]Tag, 0, len(source.Tags)),
	}
	for _, sv := range source.Variants {
		out.Variants = append(out.Variants, Variant{
			Key:        sv.Key,
			Attachment: sv.Attachment,
		})
	}

	segments := append([]Segment(nil), source.Segments...)
	sort.SliceStable(segments, func(i, j int) bool {
		if segments[i].Rank != segments[j].Rank {
			return segments[i].Rank < segments[j].Rank
		}
		return segments[i].ID < segments[j].ID
	})

	for _, ss := range segments {
		seg := Segment{
			Description:    ss.Description,
			Rank:           ss.Rank,
			RolloutPercent: ss.RolloutPercent,
		}
		for _, sc := range ss.Constraints {
			seg.Constraints = append(seg.Constraints, Constraint{
				Property: sc.Property,
				Operator: sc.Operator,
				Value:    sc.Value,
			})
		}
		for _, sd := range ss.Distributions {
			seg.Distributions = append(seg.Distributions, Distribution{
				VariantKey: sd.VariantKey,
				Percent:    sd.Percent,
				Bitmap:     sd.Bitmap,
			})
		}
		out.Segments = append(out.Segments, seg)
	}

	for _, st := range source.Tags {
		out.Tags = append(out.Tags, Tag{Value: st.Value})
	}
	return out
}

// ApplyFlagTemplate persists a template's variants, segments, constraints, distributions, and tags onto flagID.
// Used for create templates (e.g. simple_boolean_flag) and for duplicating an existing flag's graph.
func ApplyFlagTemplate(tx *gorm.DB, flagID uint, template Flag) error {
	variantByKey := make(map[string]uint, len(template.Variants))
	for _, vs := range template.Variants {
		nv := &Variant{
			FlagID:     flagID,
			Key:        vs.Key,
			Attachment: vs.Attachment,
		}
		if err := nv.Validate(); err != nil {
			return err
		}
		if err := tx.Create(nv).Error; err != nil {
			return err
		}
		variantByKey[vs.Key] = nv.ID
	}

	for _, ss := range template.Segments {
		ns := &Segment{
			FlagID:         flagID,
			Description:    ss.Description,
			Rank:           ss.Rank,
			RolloutPercent: ss.RolloutPercent,
		}
		if err := tx.Create(ns).Error; err != nil {
			return err
		}
		for _, sc := range ss.Constraints {
			nc := &Constraint{
				SegmentID: ns.ID,
				Property:  sc.Property,
				Operator:  sc.Operator,
				Value:     sc.Value,
			}
			if err := nc.Validate(); err != nil {
				return err
			}
			if err := tx.Create(nc).Error; err != nil {
				return err
			}
		}
		for _, sd := range ss.Distributions {
			vid, ok := variantByKey[sd.VariantKey]
			if !ok {
				return fmt.Errorf("distribution references unknown variant key %q", sd.VariantKey)
			}
			nd := &Distribution{
				SegmentID:  ns.ID,
				VariantID:  vid,
				VariantKey: sd.VariantKey,
				Percent:    sd.Percent,
				Bitmap:     sd.Bitmap,
			}
			if err := tx.Create(nd).Error; err != nil {
				return err
			}
		}
	}

	for _, st := range template.Tags {
		if err := AppendTagValueToFlag(tx, flagID, st.Value); err != nil {
			return err
		}
	}
	return nil
}