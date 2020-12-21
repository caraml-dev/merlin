package service

import (
	"testing"

	"github.com/gojek/merlin/models"
	"github.com/stretchr/testify/assert"
)

func TestEncodeCursor(t *testing.T) {
	t.Run("Should success", func(t *testing.T) {
		versionID := models.ID(1)
		modelID := models.ID(2)

		encodedCursor := encodeCursor(versionID, modelID)
		assert.Equal(t, "MV8y", encodedCursor)
	})
}

func TestDecodeCursor(t *testing.T) {
	testCases := []struct {
		desc           string
		encodedCursor  string
		expectedCursor cursorPagination
		isValid        bool
	}{
		{
			desc:           "Should successfully decode valid cursor",
			encodedCursor:  "MV8y",
			expectedCursor: cursorPagination{versionID: models.ID(1), versionModelID: models.ID(2)},
			isValid:        true,
		},
		{
			desc:           "Should successfully decode, but invalid cursor(versionID is not number)",
			encodedCursor:  "MXhfNA==", //1x_4
			expectedCursor: cursorPagination{},
			isValid:        false,
		},
		{
			desc:           "Should successfully decode, but invalid cursor(modelID is not number)",
			encodedCursor:  "MV80eA==", // 1_4x
			expectedCursor: cursorPagination{},
			isValid:        false,
		},
		{
			desc:           "Should successfully decode, but invalid cursor(only has versionID)",
			encodedCursor:  "MTEx", // 111
			expectedCursor: cursorPagination{},
			isValid:        false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cursor, valid := decodeCursor(tC.encodedCursor)
			assert.Equal(t, tC.expectedCursor, cursor)
			assert.Equal(t, tC.isValid, valid)
		})
	}
}
