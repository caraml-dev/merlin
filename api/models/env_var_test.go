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

package models

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestEnvVars_CheckForProtectedEnvVars(t *testing.T) {
	tests := []struct {
		name    string
		evs     EnvVars
		wantErr bool
	}{
		{
			"no protected",
			EnvVars{
				EnvVar{"foo", "bar"},
			},
			false,
		},
		{
			"protected exist",
			EnvVars{
				EnvVar{envModelName, "test"},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.evs.CheckForProtectedEnvVars(); (err != nil) != tt.wantErr {
				t.Errorf("EnvVars.CheckForProtectedEnvVars() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvVars_ToKubernetesEnvVars(t *testing.T) {
	tests := []struct {
		name string
		errRaised    EnvVars
		want []v1.EnvVar
	}{
		{
			"empty",
			EnvVars{},
			[]v1.EnvVar{},
		},
		{
			"1",
			EnvVars{
				EnvVar{Name: "foo", Value: "bar"},
			},
			[]v1.EnvVar{
				v1.EnvVar{Name: "foo", Value: "bar"},
			},
		},
		{
			"2",
			EnvVars{
				EnvVar{Name: "1", Value: "1"},
				EnvVar{Name: "2", Value: "2"},
			},
			[]v1.EnvVar{
				v1.EnvVar{Name: "1", Value: "1"},
				v1.EnvVar{Name: "2", Value: "2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errRaised.ToKubernetesEnvVars(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EnvVars.ToKubernetesEnvVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeEnvVars(t *testing.T) {
	type args struct {
		left  EnvVars
		right EnvVars
	}
	tests := []struct {
		name string
		args args
		want EnvVars
	}{
		{
			"both empty",
			args{
				EnvVars{},
				EnvVars{},
			},
			EnvVars{},
		},
		{
			"right empty",
			args{
				EnvVars{
					EnvVar{"left_foo_1", "left_bar_1"},
				},
				EnvVars{},
			},
			EnvVars{
				EnvVar{"left_foo_1", "left_bar_1"},
			},
		},
		{
			"left empty",
			args{
				EnvVars{},
				EnvVars{
					EnvVar{"right_foo_1", "right_bar_1"},
				},
			},
			EnvVars{
				EnvVar{"right_foo_1", "right_bar_1"},
			},
		},
		{
			"no overlap",
			args{
				EnvVars{
					EnvVar{"left_foo_1", "left_bar_1"},
				},
				EnvVars{
					EnvVar{"right_foo_1", "right_bar_1"},
				},
			},
			EnvVars{
				EnvVar{"left_foo_1", "left_bar_1"},
				EnvVar{"right_foo_1", "right_bar_1"},
			},
		},
		{
			"one overlap - 2",
			args{
				EnvVars{
					EnvVar{"left_foo_1", "left_bar_1"},
				},
				EnvVars{
					EnvVar{"left_foo_1", "right_bar_1"},
				},
			},
			EnvVars{
				EnvVar{"left_foo_1", "right_bar_1"},
			},
		},
		{
			"one overlap - 2",
			args{
				EnvVars{
					EnvVar{"left_foo_1", "left_bar_1"},
				},
				EnvVars{
					EnvVar{"left_foo_1", "right_bar_1"},
					EnvVar{"right_foo_1", "right_bar_1"},
				},
			},
			EnvVars{
				EnvVar{"left_foo_1", "right_bar_1"},
				EnvVar{"right_foo_1", "right_bar_1"},
			},
		},
		{
			"overlap workers",
			args{
				EnvVars{
					EnvVar{"MODEL_NAME", "TEST"},
					EnvVar{"WORKERS", "2"},
					EnvVar{"REPLICAS", "10"},
				},
				EnvVars{
					EnvVar{"WORKERS", "4"},
					EnvVar{"CPU_REQUEST", "2"},
					EnvVar{"MEMORY_REQUEST", "2Gi"},
				},
			},
			EnvVars{
				EnvVar{"MODEL_NAME", "TEST"},
				EnvVar{"WORKERS", "4"},
				EnvVar{"REPLICAS", "10"},
				EnvVar{"CPU_REQUEST", "2"},
				EnvVar{"MEMORY_REQUEST", "2Gi"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeEnvVars(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeEnvVars() = %v, want %v", got, tt.want)
			}
		})
	}
}
