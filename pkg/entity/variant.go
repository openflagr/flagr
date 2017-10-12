//go:generate goqueryset -in variant.go

package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/jinzhu/gorm"
)

// Variant is the struct that represent the experience/variant of the evaluation entity
// gen:qs
type Variant struct {
	gorm.Model
	FlagID     uint `gorm:"index:idx_flagid"`
	Key        string
	Attachment Attachment `sql:"type:text"`
}

// Attachment supports dynamic configuration in variant
type Attachment map[string]string

// Scan implements scanner interface
func (a *Attachment) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if b, ok := value.([]byte); ok {
		err := json.Unmarshal(b, a)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("Cannot scan %v into Attachment type", value)
}

// Value implements valuer interface
func (a Attachment) Value() (driver.Value, error) {
	bytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}
