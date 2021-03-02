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

func Test_generateLabelsWhereQuery(t *testing.T) {
	type args struct {
		freeTextQuery string
	}
	tests := []struct {
		name      string
		args      args
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name:      "empty query",
			args:      args{freeTextQuery: ""},
			wantQuery: "",
			wantArgs:  nil,
		},
		{
			name:      "empty label key, one label value",
			args:      args{freeTextQuery: " in (GO-RIDE)"},
			wantQuery: "",
			wantArgs:  nil,
		},
		{
			name:      "one label key, one label value",
			args:      args{freeTextQuery: "service_type in (GO-RIDE)"},
			wantQuery: "(labels @> ?)",
			wantArgs:  []interface{}{`{"service_type": "GO-RIDE"}`},
		},
		{
			name:      "one label key, one label value with unclosed parenthesis",
			args:      args{freeTextQuery: "service_type in (GO-RIDE"},
			wantQuery: "",
			wantArgs:  nil,
		},
		{
			name:      "one label key, empty value",
			args:      args{freeTextQuery: "service_type in ()"},
			wantQuery: "",
			wantArgs:  nil,
		},
		{
			name:      "one label key, one label value with invalid chars, should ignore parts before invalid chars",
			args:      args{freeTextQuery: "service%%__type in (GO-RIDE)"},
			wantQuery: "(labels @> ?)",
			wantArgs:  []interface{}{`{"__type": "GO-RIDE"}`},
		},
		{
			name:      "one label key, multiple label values",
			args:      args{freeTextQuery: "service_type in (GO-RIDE, GO-CAR)"},
			wantQuery: "(labels @> ? OR labels @> ?)",
			wantArgs:  []interface{}{`{"service_type": "GO-RIDE"}`, `{"service_type": "GO-CAR"}`},
		},
		{
			name:      "one label key, multiple label values with extra whitespaces and uppercase 'in'",
			args:      args{freeTextQuery: "   service_type IN   (GO-RIDE  ,   GO-CAR )"},
			wantQuery: "(labels @> ? OR labels @> ?)",
			wantArgs:  []interface{}{`{"service_type": "GO-RIDE"}`, `{"service_type": "GO-CAR"}`},
		},
		{
			name:      "multiple label keys and multiple label values",
			args:      args{freeTextQuery: "service_type in (GO-RIDE, GO-CAR), service2-area in (S-9_G3)"},
			wantQuery: "(labels @> ? OR labels @> ?) AND (labels @> ?)",
			wantArgs:  []interface{}{`{"service_type": "GO-RIDE"}`, `{"service_type": "GO-CAR"}`, `{"service2-area": "S-9_G3"}`},
		},
		{
			name:      "multiple label keys and multiple label values with empty keys",
			args:      args{freeTextQuery: "service_type in (GO-RIDE, GO-CAR), service-area in (SG), in (foo)"},
			wantQuery: "(labels @> ? OR labels @> ?) AND (labels @> ?)",
			wantArgs:  []interface{}{`{"service_type": "GO-RIDE"}`, `{"service_type": "GO-CAR"}`, `{"service-area": "SG"}`},
		},
		{
			name:      "multiple label keys and multiple label values with extra meaningless terms",
			args:      args{freeTextQuery: "service_type in (GO-RIDE, GO-CAR), service-area in (SG), quick, brown in fox, moomin in (troll)"},
			wantQuery: "(labels @> ? OR labels @> ?) AND (labels @> ?) AND (labels @> ?)",
			wantArgs:  []interface{}{`{"service_type": "GO-RIDE"}`, `{"service_type": "GO-CAR"}`, `{"service-area": "SG"}`, `{"moomin": "troll"}`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs := generateLabelsWhereQuery(tt.args.freeTextQuery)
			if gotQuery != tt.wantQuery {
				t.Errorf("generateLabelsWhereQuery() gotQuery = %v, want %v", gotQuery, tt.wantQuery)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("generateLabelsWhereQuery() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}