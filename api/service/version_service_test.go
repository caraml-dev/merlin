package service

import (
	"reflect"
	"testing"
)

func Test_getVersionSearchTerms(t *testing.T) {
	type args struct {
		query string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "empty query",
			args: args{""},
			want: map[string]string{},
		},
		{
			name: "empty query with whitespaces",
			args: args{"  "},
			want: map[string]string{},
		},
		{
			name: "single key empty value",
			args: args{"foo:"},
			want: map[string]string{"foo":""},
		},
		{
			name: "single key empty value with whitespaces",
			args: args{"foo :    "},
			want: map[string]string{"foo":""},
		},
		{
			name: "single key non-empty value",
			args: args{"foo: bar"},
			want: map[string]string{"foo": "bar"},
		},
		{
			name: "single key non-empty value with whitespaces",
			args: args{"foo     :bar"},
			want: map[string]string{"foo": "bar"},
		},
		{
			name: "multiple keys with empty key",
			args: args{"foo::bar"},
			want: map[string]string{"foo": ""},
		},
		{
			name: "multiple keys with empty key and whitespaces",
			args: args{"foo :     : bar"},
			want: map[string]string{"foo": ""},
		},
		{
			name: "multiple keys and some empty values",
			args: args{"foo:bar:baz"},
			want: map[string]string{"foo": "", "bar": "baz"},
		},
		{
			name: "multiple keys with empty key and values",
			args: args{"foo:bar::"},
			want: map[string]string{"foo": "", "bar": ""},
		},
		{
			name: "multiple keys and non-empty values",
			args: args{"foo:bar baz:qux quux "},
			want: map[string]string{"foo": "bar", "baz": "qux quux"},
		},
		{
			name: "duplicate keys and non-empty values",
			args: args{"foo:bar baz:qux quux foo: corge"},
			want: map[string]string{"foo": "corge", "baz": "qux quux"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getVersionSearchTerms(tt.args.query); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVersionSearchTerms() = %v, want %v", got, tt.want)
			}
		})
	}
}
