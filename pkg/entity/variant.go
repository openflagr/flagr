package entity

import (
	"database/sql/driver"
	"fmt"

	jsoniter "github.com/json-iterator/go"

	"github.com/openflagr/flagr/pkg/util"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

// Variant is the struct that represent the experience/variant of the evaluation entity
type Variant struct {
	gorm.Model
	FlagID     uint `gorm:"index:idx_variant_flagid"`
	Key        string
	Attachment Attachment `gorm:"type:text"`
}

// Validate validates the Variant
func (v *Variant) Validate() error {
	ok, msg := util.IsSafeKey(v.Key)
	if !ok {
		return fmt.Errorf(msg)
	}
	return nil
}

// Attachment supports dynamic configuration in variant
type Attachment map[string]interface{}

// Scan implements scanner interface
func (a *Attachment) Scan(value interface{}) error {
	var json = jsoniter.ConfigFastest

	if value == nil {
		return nil
	}
	s := cast.ToString(value)
	if err := json.Unmarshal([]byte(s), a); err != nil {
		return fmt.Errorf("cannot scan %v into Attachment type. err: %v", value, err)
	}
	return nil
}

// Value implements valuer interface
func (a Attachment) Value() (driver.Value, error) {
	var json = jsoniter.ConfigFastest

	bytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}
