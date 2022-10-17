package hystrix

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DoerFunc func(req *http.Request) (*http.Response, error)

func (fun DoerFunc) Do(req *http.Request) (*http.Response, error) { return fun(req) }

func Test_Do(t *testing.T) {
	testCases := []struct {
		desc    string
		doer    Doer
		want    *http.Response
		wantErr error
	}{
		{
			desc: "Should success without body",
			doer: DoerFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
				}, nil
			}),
			want: &http.Response{
				StatusCode: http.StatusOK,
			},
			wantErr: nil,
		},
		{
			desc: "Should success with body",
			doer: DoerFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}),
			want: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
			},
			wantErr: nil,
		},
		{
			desc: "Should not return error if got 4xx status code",
			doer: DoerFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusForbidden,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}),
			want: &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(strings.NewReader("")),
			},
			wantErr: nil,
		},
		{
			desc: "Should error if got 5xx status code",
			doer: DoerFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}),
			want: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("")),
			},
			wantErr: fmt.Errorf("got 5xx response code: %d", http.StatusInternalServerError),
		},
		{
			desc: "Should error when doer failed got response",
			doer: DoerFunc(func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("something went wrong")
			}),
			want:    nil,
			wantErr: errors.New("something went wrong"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
			require.NoError(t, err)
			require.NotNil(t, req)

			client := NewClient(tC.doer, nil, "call_api")
			resp, err := client.Do(req)
			if tC.wantErr != nil {
				assert.Error(t, tC.wantErr, err)
				assert.EqualError(t, err, tC.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tC.want, resp)
				if resp.Body != nil {
					assert.NoError(t, resp.Body.Close())
				}
			}
		})
	}
}
