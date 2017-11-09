package entity

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	t.Run("empty case", func(t *testing.T) {
		v := Variant{}
		err := v.Validate()
		assert.Error(t, err)
	})
	t.Run("happy code path", func(t *testing.T) {
		v := Variant{
			Model:      gorm.Model{},
			FlagID:     0,
			Key:        "a123",
			Attachment: nil,
		}
		err := v.Validate()
		assert.NoError(t, err)
	})
}
