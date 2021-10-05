package feast

import (
	"testing"

	"github.com/gojek/merlin/pkg/transformer/spec"
)

func TestGetTableName(t *testing.T) {
	type args struct {
		featureTableSpec *spec.FeatureTable
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty table name and default project name",
			args: args{
				featureTableSpec: &spec.FeatureTable{
					Project:   "default",
					TableName: "",
					Entities: []*spec.Entity{
						{
							Name: "entity_1",
						},
						{
							Name: "entity_2",
						},
					},
				},
			},
			want: "entity_1_entity_2",
		},
		{
			name: "empty table name and non default project name",
			args: args{
				featureTableSpec: &spec.FeatureTable{
					Project:   "my-project",
					TableName: "",
					Entities: []*spec.Entity{
						{
							Name: "entity_1",
						},
						{
							Name: "entity_2",
						},
					},
				},
			},
			want: "my-project_entity_1_entity_2",
		},
		{
			name: "table name is defined",
			args: args{
				featureTableSpec: &spec.FeatureTable{
					Project:   "my-project",
					TableName: "my-table",
					Entities: []*spec.Entity{
						{
							Name: "entity_1",
						},
						{
							Name: "entity_2",
						},
					},
				},
			},
			want: "my-table",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTableName(tt.args.featureTableSpec); got != tt.want {
				t.Errorf("GetTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}
