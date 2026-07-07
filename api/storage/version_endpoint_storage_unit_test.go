package storage

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/caraml-dev/merlin/models"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeEndpoint_RemovesInvalidUTF8(t *testing.T) {
	endpoint := &models.VersionEndpoint{
		Message: "prefix" + string([]byte{0xe2, 0x94}) + "suffix",
	}

	sanitizeEndpoint(endpoint)

	assert.Equal(t, "prefixsuffix", endpoint.Message)
	assert.True(t, utf8.ValidString(endpoint.Message))
}

func TestSanitizeEndpoint_TruncatesToValidUTF8(t *testing.T) {
	message := strings.Repeat("a", maxMessageChar-1) + "e" + "\u0301"
	endpoint := &models.VersionEndpoint{
		Message: message,
	}

	sanitizeEndpoint(endpoint)

	assert.Equal(t, maxMessageChar, len(endpoint.Message))
	assert.True(t, utf8.ValidString(endpoint.Message))
}
