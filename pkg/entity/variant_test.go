package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
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

func TestVariantScan(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		a := &Attachment{}
		err := a.Scan([]byte(`{"key": "value"}`))
		assert.NoError(t, err)
	})

	t.Run("nil value", func(t *testing.T) {
		a := &Attachment{}
		err := a.Scan(nil)
		assert.NoError(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		a := &Attachment{}
		err := a.Scan([]byte(`{`))
		assert.Error(t, err)
	})

	t.Run("invalid value type", func(t *testing.T) {
		a := &Attachment{}
		err := a.Scan(123)
		assert.Error(t, err)
	})
}
