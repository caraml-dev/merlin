package models

import (
	"testing"

	"github.com/gojek/merlin/pkg/protocol"
	"github.com/stretchr/testify/assert"
	"knative.dev/pkg/apis"
)

func TestGetValidInferenceURL(t *testing.T) {
	testCases := []struct {
		desc          string
		url           string
		inferenceName string
		protocol      protocol.Protocol
		expectedUrl   string
	}{
		{
			desc:          "Should return valid inferenceURL without appending suffix",
			url:           "http://sklearn.default.domain.com/v1/models/sklearn",
			inferenceName: "sklearn",
			expectedUrl:   "http://sklearn.default.domain.com/v1/models/sklearn",
		},
		{
			desc:          "Should return valid inferenceURL with appending suffix",
			url:           "http://sklearn.default.domain.com",
			inferenceName: "sklearn",
			expectedUrl:   "http://sklearn.default.domain.com/v1/models/sklearn",
		},
		{
			desc:          "UPI V1 Protocol: should return hostname",
			url:           "http://sklearn.default.domain.com",
			inferenceName: "sklearn",
			expectedUrl:   "sklearn.default.domain.com",
			protocol:      protocol.UpiV1,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			url, _ := apis.ParseURL(tC.url)
			res := GetInferenceURL(url, tC.inferenceName, tC.protocol)
			assert.Equal(t, tC.expectedUrl, res)
		})
	}
}
