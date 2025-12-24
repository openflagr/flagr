package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateDiff(t *testing.T) {
	t.Run("empty cases", func(t *testing.T) {
		assert.Empty(t, CalculateDiff("", ""))
		assert.Empty(t, CalculateDiff("a", ""))
		assert.Empty(t, CalculateDiff("", "b"))
	})

	t.Run("simple diff", func(t *testing.T) {
		pre := "line1\nline2\n"
		post := "line1\nline3\n"
		diff := CalculateDiff(pre, post)
		assert.NotEmpty(t, diff)
		assert.Contains(t, diff, "-line2")
		assert.Contains(t, diff, "+line3")
	})

	t.Run("JSON diff visibility", func(t *testing.T) {
		pre := `{"id":1,"key":"flag1","enabled":false}`
		post := `{"id":1,"key":"flag1","enabled":true}`
		diff := CalculateDiff(pre, post)
		t.Logf("Pretty JSON Diff:\n%s", diff)
		// Pretty JSON diff shows individual field changes
		assert.Contains(t, diff, "-  \"enabled\": false")
		assert.Contains(t, diff, "+  \"enabled\": true")
	})
}
