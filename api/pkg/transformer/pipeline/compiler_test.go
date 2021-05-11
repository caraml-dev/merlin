package pipeline

import (
	"io/ioutil"
	"testing"

	feastSdk "github.com/feast-dev/feast/sdk/go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"

	"github.com/gojek/merlin/pkg/transformer/cache"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
)

func TestCompiler_Compile(t *testing.T) {
	type (
		fields struct {
			sr           symbol.Registry
			feastClient  feastSdk.Client
			feastOptions *feast.Options
			cacheOptions *cache.Options
			logger       *zap.Logger
		}

		want struct {
			expressions    []string
			jsonPaths      []string
			preprocessOps  []Op
			postprocessOps []Op
		}
	)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name             string
		fields           fields
		specYamlFilePath string
		want             want
		wantErr          bool
		expError         error
	}{
		{
			name: "preprocess input only",
			fields: fields{
				sr:          symbol.NewRegistry(),
				feastClient: &mocks.Client{},
				feastOptions: &feast.Options{
					CacheEnabled: true,
				},
				cacheOptions: &cache.Options{
					SizeInMB: 100,
				},
				logger: logger,
			},
			specYamlFilePath: "./testdata/valid_preprocess_input_only.yaml",
			want: want{
				expressions: []string{
					"Now()",
					"variable1",
				},
				jsonPaths: []string{
					"$.entity_1[*].id",
					"$.entity_2.id",
					"$.entity_3",
					"$.entity_2",
				},
				preprocessOps: []Op{
					&FeastOp{},
					&CreateTableOp{},
					&VariableDeclarationOp{},
				},
				postprocessOps: []Op{},
			},
			wantErr: false,
		},
		{
			name: "no pipeline",
			fields: fields{
				sr:          symbol.NewRegistry(),
				feastClient: &mocks.Client{},
				feastOptions: &feast.Options{
					CacheEnabled: true,
				},
				cacheOptions: &cache.Options{
					SizeInMB: 100,
				},
				logger: logger,
			},
			specYamlFilePath: "./testdata/valid_no_pipeline.yaml",
			want: want{
				expressions:    []string{},
				jsonPaths:      []string{},
				preprocessOps:  []Op{},
				postprocessOps: []Op{},
			},
			wantErr: false,
		},
		{
			name: "preprocess postprocess input only",
			fields: fields{
				sr:          symbol.NewRegistry(),
				feastClient: &mocks.Client{},
				feastOptions: &feast.Options{
					CacheEnabled: true,
				},
				cacheOptions: &cache.Options{
					SizeInMB: 100,
				},
				logger: logger,
			},
			specYamlFilePath: "./testdata/valid_input_only.yaml",
			want: want{
				expressions: []string{
					"Now()",
					"variable1",
				},
				jsonPaths: []string{
					"$.entity_1[*].id",
					"$.entity_2.id",
					"$.entity_3",
					"$.entity_2",
				},
				preprocessOps: []Op{
					&FeastOp{},
					&CreateTableOp{},
					&VariableDeclarationOp{},
				},
				postprocessOps: []Op{
					&FeastOp{},
					&CreateTableOp{},
					&VariableDeclarationOp{},
				},
			},
			wantErr: false,
		},
		{
			name: "preprocess with input and transformation",
			fields: fields{
				sr:          symbol.NewRegistry(),
				feastClient: &mocks.Client{},
				feastOptions: &feast.Options{
					CacheEnabled: true,
				},
				cacheOptions: &cache.Options{
					SizeInMB: 100,
				},
				logger: logger,
			},
			specYamlFilePath: "./testdata/valid_preprocess_input_and_transformation.yaml",
			want: want{
				expressions: []string{
					"Now()",
					"variable1",
					"entity_2_table.Col('col1')",
					"Now().Hour()",
				},
				jsonPaths: []string{
					"$.entity_1[*].id",
					"$.entity_2.id",
					"$.entity_3",
					"$.entity_2",
				},
				preprocessOps: []Op{
					&FeastOp{},
					&CreateTableOp{},
					&VariableDeclarationOp{},
					&TableTransformOp{},
					&TableJoinOp{},
				},
			},
			wantErr: false,
		},
		{
			name: "sequential table dependency",
			fields: fields{
				sr:          symbol.NewRegistry(),
				feastClient: &mocks.Client{},
				feastOptions: &feast.Options{
					CacheEnabled: true,
				},
				cacheOptions: &cache.Options{
					SizeInMB: 100,
				},
				logger: logger,
			},
			specYamlFilePath: "./testdata/valid_sequential_table_dependency.yaml",
			want: want{
				expressions: []string{
					"Now()",
					"variable1",
					"entity_2_table.Col('col1')",
					"Now().Hour()",
				},
				jsonPaths: []string{
					"$.entity_1[*].id",
					"$.entity_2.id",
					"$.entity_3",
					"$.entity_2",
				},
				preprocessOps: []Op{
					&FeastOp{},
					&CreateTableOp{},
					&VariableDeclarationOp{},
					&TableTransformOp{},
					&TableTransformOp{},
					&TableJoinOp{},
					&TableJoinOp{},
				},
			},
			wantErr: false,
		},
		{
			name: "preprocess - postprocess input and output - valid",
			fields: fields{
				sr:          symbol.NewRegistry(),
				feastClient: &mocks.Client{},
				feastOptions: &feast.Options{
					CacheEnabled: true,
				},
				cacheOptions: &cache.Options{
					SizeInMB: 100,
				},
				logger: logger,
			},
			specYamlFilePath: "./testdata/input_output.yaml",
			want: want{
				expressions: []string{
					"Now()",
					"variable1",
				},
				jsonPaths: []string{
					"$.entity_2.id",
					"$.entity_3",
					"$.entity_2",
					"$.path_1",
					"$.path_2",
				},
				preprocessOps: []Op{
					&CreateTableOp{},
					&VariableDeclarationOp{},
					&JsonOutputOp{},
				},
				postprocessOps: []Op{
					&CreateTableOp{},
					&VariableDeclarationOp{},
					&JsonOutputOp{},
				},
			},
			wantErr: false,
		},
		{
			name: "preprocess - postprocess input and output - invalid",
			fields: fields{
				sr:          symbol.NewRegistry(),
				feastClient: &mocks.Client{},
				feastOptions: &feast.Options{
					CacheEnabled: true,
				},
				cacheOptions: &cache.Options{
					SizeInMB: 100,
				},
				logger: logger,
			},
			specYamlFilePath: "./testdata/invalid_output.yaml",
			wantErr:          true,
			expError:         errors.New("unable to compile preprocessing pipeline: variable entity_5_table is not registered"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Compiler{
				sr:           tt.fields.sr,
				feastClient:  tt.fields.feastClient,
				feastOptions: tt.fields.feastOptions,
				cacheOptions: tt.fields.cacheOptions,
				logger:       tt.fields.logger,
			}

			yamlBytes, err := ioutil.ReadFile(tt.specYamlFilePath)
			assert.NoError(t, err)

			jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
			assert.NoError(t, err)

			var stdSpec spec.StandardTransformerConfig
			err = protojson.Unmarshal(jsonBytes, &stdSpec)
			assert.NoError(t, err)

			got, err := c.Compile(&stdSpec)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}
			assert.NoError(t, err)

			for _, jsonPath := range tt.want.jsonPaths {
				assert.NotNil(t, got.compiledJsonpath.Get(jsonPath), "json path not compiled", jsonPath)
			}

			for _, expression := range tt.want.expressions {
				assert.NotNil(t, got.compiledExpression.Get(expression), "expression not compiled", expression)
			}

			assert.Equal(t, len(tt.want.preprocessOps), len(got.preprocessOps))
			for i, op := range tt.want.preprocessOps {
				assert.IsType(t, op, got.preprocessOps[i], "different type")
			}

			assert.Equal(t, len(tt.want.postprocessOps), len(got.postprocessOps))
			for i, op := range tt.want.postprocessOps {
				assert.IsType(t, op, got.postprocessOps[i], "different type")
			}
		})
	}
}
