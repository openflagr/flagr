package entity

import "gorm.io/gorm"

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
