package cronjob

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c, err := New()
	assert.Nil(t, err)
	assert.NotNil(t, c)
}
