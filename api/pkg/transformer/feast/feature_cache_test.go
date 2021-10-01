package feast

import (
	"testing"
	"time"

	feast "github.com/feast-dev/feast/sdk/go"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/pkg/transformer/types"
)

func TestFetchFeatureTable(t *testing.T) {
	type cacheConfig struct {
		ttl      time.Duration
		sizeInMB int
	}
	type args struct {
		entities    []feast.Row
		columnNames []string
		project     string
	}
	tests := []struct {
		name           string
		cacheConfig    cacheConfig
		args           args
		valueInCache   *internalFeatureTable
		want           *internalFeatureTable
		missedEntities []feast.Row
	}{
		{
			name: "single entity, no value in cache",
			cacheConfig: cacheConfig{
				ttl:      10 * time.Minute,
				sizeInMB: 10,
			},
			args: args{
				entities: []feast.Row{
					{
						"driver_id": feast.StrVal("1001"),
					},
				},
				columnNames: []string{"feature1", "feature2"},
				project:     "my-project",
			},
			valueInCache: nil,
			want: &internalFeatureTable{
				entities:    nil,
				columnNames: []string{"feature1", "feature2"},
				columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_INVALID, feastTypes.ValueType_INVALID},
				valueRows:   nil,
			},
			missedEntities: []feast.Row{
				{
					"driver_id": feast.StrVal("1001"),
				},
			},
		},
		{
			name: "two entities, no value in cache",
			cacheConfig: cacheConfig{
				ttl:      10 * time.Minute,
				sizeInMB: 10,
			},
			args: args{
				entities: []feast.Row{
					{
						"driver_id": feast.StrVal("1001"),
					},
					{
						"driver_id": feast.StrVal("1002"),
					},
				},
				columnNames: []string{"feature1", "feature2"},
				project:     "my-project",
			},
			valueInCache: nil,
			want: &internalFeatureTable{
				entities:    nil,
				columnNames: []string{"feature1", "feature2"},
				columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_INVALID, feastTypes.ValueType_INVALID},
				valueRows:   nil,
			},
			missedEntities: []feast.Row{
				{
					"driver_id": feast.StrVal("1001"),
				},
				{
					"driver_id": feast.StrVal("1002"),
				},
			},
		},
		{
			name: "two entities, only one has value in cache",
			cacheConfig: cacheConfig{
				ttl:      10 * time.Minute,
				sizeInMB: 10,
			},
			args: args{
				entities: []feast.Row{
					{
						"driver_id": feast.StrVal("1001"),
					},
					{
						"driver_id": feast.StrVal("1002"),
					},
				},
				columnNames: []string{"feature1", "feature2"},
				project:     "my-project",
			},
			valueInCache: &internalFeatureTable{
				entities: []feast.Row{
					{
						"driver_id": feast.StrVal("1001"),
					},
				},
				columnNames: []string{"feature1", "feature2"},
				columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_STRING},
				valueRows: types.ValueRows{
					types.ValueRow{
						"val1",
						"val2",
					},
				},
			},
			want: &internalFeatureTable{
				entities: []feast.Row{
					{
						"driver_id": feast.StrVal("1001"),
					},
				},
				columnNames: []string{"feature1", "feature2"},
				columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_STRING},
				valueRows: types.ValueRows{
					types.ValueRow{
						"val1",
						"val2",
					},
				},
			},
			missedEntities: []feast.Row{
				{
					"driver_id": feast.StrVal("1002"),
				},
			},
		},
		{
			name: "two entities, both have value in cache",
			cacheConfig: cacheConfig{
				ttl:      10 * time.Minute,
				sizeInMB: 10,
			},
			args: args{
				entities: []feast.Row{
					{
						"driver_id": feast.StrVal("1001"),
					},
					{
						"driver_id": feast.StrVal("1002"),
					},
				},
				columnNames: []string{"feature1", "feature2"},
				project:     "my-project",
			},
			valueInCache: &internalFeatureTable{
				entities: []feast.Row{
					{
						"driver_id": feast.StrVal("1001"),
					},
					{
						"driver_id": feast.StrVal("1002"),
					},
				},
				columnNames: []string{"feature1", "feature2"},
				columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_STRING},
				valueRows: types.ValueRows{
					types.ValueRow{
						"val11",
						"val12",
					},
					types.ValueRow{
						"val21",
						"val22",
					},
				},
			},
			want: &internalFeatureTable{
				entities: []feast.Row{
					{
						"driver_id": feast.StrVal("1001"),
					},
					{
						"driver_id": feast.StrVal("1002"),
					},
				},
				columnNames: []string{"feature1", "feature2"},
				columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_STRING},
				valueRows: types.ValueRows{
					types.ValueRow{
						"val11",
						"val12",
					},
					types.ValueRow{
						"val21",
						"val22",
					},
				},
			},
			missedEntities: nil,
		},
		{
			name: "one entity, but cache contain different list of feature than requested",
			cacheConfig: cacheConfig{
				ttl:      10 * time.Minute,
				sizeInMB: 10,
			},
			args: args{
				entities: []feast.Row{
					{
						"driver_id": feast.StrVal("1001"),
					},
				},
				columnNames: []string{"feature1", "feature2"},
				project:     "my-project",
			},
			valueInCache: &internalFeatureTable{
				entities: []feast.Row{
					{
						"driver_id": feast.StrVal("1001"),
					},
				},
				columnNames: []string{"feature3", "feature4"},
				columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_STRING},
				valueRows: types.ValueRows{
					types.ValueRow{
						"val11",
						"val12",
					},
					types.ValueRow{
						"val21",
						"val22",
					},
				},
			},
			want: &internalFeatureTable{
				entities:    nil,
				columnNames: []string{"feature1", "feature2"},
				columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_INVALID, feastTypes.ValueType_INVALID},
				valueRows:   nil,
			},
			missedEntities: []feast.Row{
				{
					"driver_id": feast.StrVal("1001"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := newFeatureCache(tt.cacheConfig.ttl, tt.cacheConfig.sizeInMB)
			if tt.valueInCache != nil {
				err := fc.insertFeatureTable(tt.valueInCache, tt.args.project)
				if err != nil {
					t.Fatalf("unable to pre-populate cache: %v", err)
				}
			}
			got, missedEntities := fc.fetchFeatureTable(tt.args.entities, tt.args.columnNames, tt.args.project)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.missedEntities, missedEntities)
		})
	}
}

func TestInsertFeatureTable(t *testing.T) {
	type cacheConfig struct {
		ttl      time.Duration
		sizeInMB int
	}
	type args struct {
		featureTable *internalFeatureTable
		project      string
	}
	tests := []struct {
		name         string
		cacheConfig  cacheConfig
		args         args
		wantErr      bool
		wantErrorMsg string
	}{
		{
			name: "insert table containing one entity",
			cacheConfig: cacheConfig{
				ttl:      10 * time.Minute,
				sizeInMB: 10,
			},
			args: args{
				featureTable: &internalFeatureTable{
					entities: []feast.Row{
						{
							"driver_id": feast.StrVal("1001"),
						},
					},
					columnNames: []string{"feature3", "feature4"},
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_STRING},
					valueRows: types.ValueRows{
						types.ValueRow{
							"val11",
							"val12",
						},
					},
				},
				project: "my-project",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := newFeatureCache(tt.cacheConfig.ttl, tt.cacheConfig.sizeInMB)
			err := fc.insertFeatureTable(tt.args.featureTable, tt.args.project)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error = %v", err)
					return
				}

				assert.EqualError(t, err, tt.wantErrorMsg)
			}

			got, _ := fc.fetchFeatureTable(tt.args.featureTable.entities, tt.args.featureTable.columnNames, tt.args.project)
			if err != nil {
				t.Errorf("unexpected error returned from fetchFeatureTable = %v", err)
			}

			assert.Equal(t, tt.args.featureTable, got)
		})
	}
}
