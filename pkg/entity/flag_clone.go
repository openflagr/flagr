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

// CloneFlagGraph copies variants, segments (constraints, distributions), and tags from source onto dest inside tx.
func CloneFlagGraph(tx *gorm.DB, source *Flag, dest *Flag) error {
	variantMap := make(map[uint]uint, len(source.Variants))
	for _, sv := range source.Variants {
		nv := &Variant{
			FlagID:     dest.ID,
			Key:        sv.Key,
			Attachment: sv.Attachment,
		}
		if err := nv.Validate(); err != nil {
			return err
		}
		if err := tx.Create(nv).Error; err != nil {
			return err
		}
		variantMap[sv.ID] = nv.ID
	}

	segments := append([]Segment(nil), source.Segments...)
	sort.SliceStable(segments, func(i, j int) bool {
		if segments[i].Rank != segments[j].Rank {
			return segments[i].Rank < segments[j].Rank
		}
		return segments[i].ID < segments[j].ID
	})

	for _, ss := range segments {
		ns := &Segment{
			FlagID:         dest.ID,
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
			newVID, ok := variantMap[sd.VariantID]
			if !ok {
				return fmt.Errorf("distribution references unknown variant id %d", sd.VariantID)
			}
			nd := &Distribution{
				SegmentID:  ns.ID,
				VariantID:  newVID,
				VariantKey: sd.VariantKey,
				Percent:    sd.Percent,
				Bitmap:     sd.Bitmap,
			}
			if err := tx.Create(nd).Error; err != nil {
				return err
			}
		}
	}

	for _, st := range source.Tags {
		if err := AppendTagValueToFlag(tx, dest.ID, st.Value); err != nil {
			return err
		}
	}
	return nil
}