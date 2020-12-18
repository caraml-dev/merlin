// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mlflow

import "testing"

func Test_errorResponse_Error(t *testing.T) {
	type fields struct {
		ErrorCode string
		Message   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"1",
			fields{
				ErrorCode: "500",
				Message:   "Internal server error",
			},
			"500",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errRaised := &errorResponse{
				ErrorCode: tt.fields.ErrorCode,
				Message:   tt.fields.Message,
			}
			if got := errRaised.Error(); got != tt.want {
				t.Errorf("errorResponse.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
