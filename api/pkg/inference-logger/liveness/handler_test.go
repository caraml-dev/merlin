package liveness

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProbe_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		requestHeaders map[string]string
		wantCallNext   bool
		wantCallInner  bool
	}{
		{
			name: "liveness probe should be forwarded to inner",
			requestHeaders: map[string]string{
				"K-Kubelet-Probe": "queue",
				"User-Agent":      "kube-probe/1.22",
			},
			wantCallInner: true,
		},
		{
			name:           "other request should be forwarded to next",
			requestHeaders: map[string]string{},
			wantCallNext:   true,
		},
		{
			name: "next probe should be forwarded to next",
			requestHeaders: map[string]string{
				"K-Network-Probe": "queue",
				"K-Network-Hash":  "1234",
				"User-Agent":      "kube-probe/1.22",
			},
			wantCallNext: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			innerCalled := false
			inner := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				innerCalled = true
			})

			nextCalled := false
			health := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				nextCalled = true
			})
			p := NewProbe(inner, health)

			r := httptest.NewRequest("GET", "http://a", nil)
			for k, v := range tt.requestHeaders {
				r.Header.Set(k, v)
			}
			w := httptest.NewRecorder()

			p.ServeHTTP(w, r)

			assert.Equal(t, tt.wantCallNext, nextCalled)
			assert.Equal(t, tt.wantCallInner, innerCalled)
		})
	}
}
