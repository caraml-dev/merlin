package series

import (
	"testing"

	"github.com/go-gota/gota/series"
	"github.com/stretchr/testify/assert"
)

func TestSeries_New(t *testing.T) {
	s := New([]interface{}{"1111", nil}, String, "string_col")

	assert.Equal(t, String, s.Type())
	assert.Equal(t, *s.Series(), series.New([]interface{}{"1111", nil}, series.String, "string_col"))

	gotaSeries := series.New([]interface{}{"1111", nil}, series.String, "string_col")
	s2 := NewSeries(&gotaSeries)
	assert.Equal(t, String, s2.Type())
	assert.Equal(t, *s2.Series(), gotaSeries)
}
