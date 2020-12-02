package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValidInferenceURL(t *testing.T) {
	testCases := []struct {
		desc              string
		url               string
		inferenceName     string
		validInferenceURL string
	}{
		{
			desc:              "Should return valid inferenceURL without appending suffix",
			url:               "http://sklearn.default.domain.com/v1/models/sklearn",
			inferenceName:     "sklearn",
			validInferenceURL: "http://sklearn.default.domain.com/v1/models/sklearn",
		},
		{
			desc:              "Should return valid inferenceURL with appending suffix",
			url:               "http://sklearn.default.domain.com",
			inferenceName:     "sklearn",
			validInferenceURL: "http://sklearn.default.domain.com/v1/models/sklearn",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res := GetValidInferenceURL(tC.url, tC.inferenceName)
			assert.Equal(t, tC.validInferenceURL, res)
		})
	}
}
