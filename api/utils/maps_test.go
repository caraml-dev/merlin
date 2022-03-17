package utils

import (
	"reflect"
	"testing"
)

func TestMergeMaps(t *testing.T) {
	type args struct {
		left  map[string]string
		right map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "both maps has different keys",
			args: args{
				left: map[string]string{
					"key-1": "value-1",
				},
				right: map[string]string{
					"key-2": "value-2",
				},
			},
			want: map[string]string{
				"key-1": "value-1",
				"key-2": "value-2",
			},
		},
		{
			name: "duplicate key",
			args: args{
				left: map[string]string{
					"key-1": "value-1",
				},
				right: map[string]string{
					"key-1": "value-11",
					"key-2": "value-2",
				},
			},
			want: map[string]string{
				"key-1": "value-11",
				"key-2": "value-2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeMaps(tt.args.left, tt.args.right); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeMaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExcludeKeys(t *testing.T) {
	type args struct {
		srcMap        map[string]string
		keysToExclude []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "all keys not in map",
			args: args{
				srcMap: map[string]string{
					"key-1": "value-1",
					"key-2": "value-2",
					"key-3": "value-3",
					"key-4": "value-4",
				},
				keysToExclude: []string{
					"key-not-found",
				},
			},
			want: map[string]string{
				"key-1": "value-1",
				"key-2": "value-2",
				"key-3": "value-3",
				"key-4": "value-4",
			},
		},
		{
			name: "all keys in map",
			args: args{
				srcMap: map[string]string{
					"key-1": "value-1",
					"key-2": "value-2",
					"key-3": "value-3",
					"key-4": "value-4",
				},
				keysToExclude: []string{
					"key-2", "key-3",
				},
			},
			want: map[string]string{
				"key-1": "value-1",
				"key-4": "value-4",
			},
		},
		{
			name: "some keys in map",
			args: args{
				srcMap: map[string]string{
					"key-1": "value-1",
					"key-2": "value-2",
					"key-3": "value-3",
					"key-4": "value-4",
				},
				keysToExclude: []string{
					"key-2", "key-3", "key-not-found",
				},
			},
			want: map[string]string{
				"key-1": "value-1",
				"key-4": "value-4",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExcludeKeys(tt.args.srcMap, tt.args.keysToExclude); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExcludeKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
