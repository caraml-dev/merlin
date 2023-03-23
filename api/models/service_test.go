package models

import (
	"reflect"
	"testing"

	"github.com/gojek/merlin/mlp"
	"github.com/gojek/merlin/pkg/protocol"
	"github.com/stretchr/testify/assert"
	"knative.dev/pkg/apis"
)

func TestGetValidInferenceURL(t *testing.T) {
	testCases := []struct {
		desc          string
		url           string
		inferenceName string
		protocol      protocol.Protocol
		expectedUrl   string
	}{
		{
			desc:          "Should return valid inferenceURL without appending suffix",
			url:           "http://sklearn.default.domain.com/v1/models/sklearn",
			inferenceName: "sklearn",
			expectedUrl:   "http://sklearn.default.domain.com/v1/models/sklearn",
		},
		{
			desc:          "Should return valid inferenceURL with appending suffix",
			url:           "http://sklearn.default.domain.com",
			inferenceName: "sklearn",
			expectedUrl:   "http://sklearn.default.domain.com/v1/models/sklearn",
		},
		{
			desc:          "UPI V1 Protocol: should return hostname",
			url:           "http://sklearn.default.domain.com",
			inferenceName: "sklearn",
			expectedUrl:   "sklearn.default.domain.com",
			protocol:      protocol.UpiV1,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			url, _ := apis.ParseURL(tC.url)
			res := GetInferenceURL(url, tC.inferenceName, tC.protocol)
			assert.Equal(t, tC.expectedUrl, res)
		})
	}
}

func Test_mergeProjectVersionLabels(t *testing.T) {
	err := InitKubernetesLabeller("gojek.com/", environmentName)
	assert.NoError(t, err)

	defer func() {
		_ = InitKubernetesLabeller("", "")
	}()

	type args struct {
		projectLabels mlp.Labels
		versionLabels KV
	}
	tests := []struct {
		name string
		args args
		want mlp.Labels
	}{
		{
			"both maps has different keys",
			args{
				projectLabels: mlp.Labels{
					{Key: "gojek.com/key-1", Value: "value-1"},
				},
				versionLabels: KV{
					"key-2": "value-2",
				},
			},
			mlp.Labels{
				{Key: "gojek.com/key-1", Value: "value-1"},
				{Key: "key-2", Value: "value-2"},
			},
		},
		{
			"both maps has different keys",
			args{
				projectLabels: mlp.Labels{
					{Key: "gojek.com/key-1", Value: "value-1"},
					{Key: "key-1", Value: "value-1"},
				},
				versionLabels: KV{
					"key-1": "value-11",
					"key-2": "value-2",
				},
			},
			mlp.Labels{
				{Key: "gojek.com/key-1", Value: "value-1"},
				{Key: "key-1", Value: "value-11"},
				{Key: "key-2", Value: "value-2"},
			},
		},
		{
			"duplicate key name without prefix",
			args{
				projectLabels: mlp.Labels{
					{Key: "gojek.com/key-1", Value: "value-1"},
				},
				versionLabels: KV{
					"key-1": "value-11",
					"key-2": "value-2",
				},
			},
			mlp.Labels{
				{Key: "gojek.com/key-1", Value: "value-1"},
				{Key: "key-1", Value: "value-11"},
				{Key: "key-2", Value: "value-2"},
			},
		},
		{
			"only project labels",
			args{
				projectLabels: mlp.Labels{
					{Key: "gojek.com/key-1", Value: "value-1"},
				},
				versionLabels: nil,
			},
			mlp.Labels{
				{Key: "gojek.com/key-1", Value: "value-1"},
			},
		},
		{
			"only version labels",
			args{
				projectLabels: nil,
				versionLabels: KV{
					"key-2": "value-2",
				},
			},
			mlp.Labels{
				{Key: "key-2", Value: "value-2"},
			},
		},
		{
			"both empty",
			args{
				projectLabels: mlp.Labels{},
				versionLabels: KV{},
			},
			mlp.Labels{},
		},
		{
			"both nil",
			args{
				projectLabels: nil,
				versionLabels: nil,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeProjectVersionLabels(tt.args.projectLabels, tt.args.versionLabels)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestNewService(t *testing.T) {
	mlpLabels := mlp.Labels{
		{Key: "key-1", Value: "value-1"},
	}
	versionLabels := KV{
		"key-1": "value-11",
		"key-2": "value-2",
	}
	serviceLabels := mlp.Labels{
		{Key: "key-1", Value: "value-11"},
		{Key: "key-2", Value: "value-2"},
	}

	project := mlp.Project{Name: "project", Labels: mlpLabels}
	model := &Model{Name: "model", Project: project}
	version := &Version{ID: 1, Labels: versionLabels}
	endpoint := &VersionEndpoint{}

	type args struct {
		model    *Model
		version  *Version
		modelOpt *ModelOption
		endpoint *VersionEndpoint
	}
	tests := []struct {
		name string
		args args
		want *Service
	}{
		{
			name: "No model option",
			args: args{
				model:    model,
				version:  version,
				modelOpt: &ModelOption{},
				endpoint: endpoint,
			},
			want: &Service{
				Name:            CreateInferenceServiceName(model.Name, version.ID.String()),
				ModelName:       model.Name,
				ModelVersion:    version.ID.String(),
				Namespace:       model.Project.Name,
				ArtifactURI:     version.ArtifactURI,
				Type:            model.Type,
				Options:         &ModelOption{},
				ResourceRequest: endpoint.ResourceRequest,
				EnvVars:         endpoint.EnvVars,
				Metadata: Metadata{
					App:       model.Name,
					Component: ComponentModelVersion,
					Labels:    serviceLabels,
					Stream:    model.Project.Stream,
					Team:      model.Project.Team,
				},
				Transformer:       endpoint.Transformer,
				Logger:            endpoint.Logger,
				DeploymentMode:    endpoint.DeploymentMode,
				AutoscalingPolicy: endpoint.AutoscalingPolicy,
				Protocol:          endpoint.Protocol,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewService(tt.args.model, tt.args.version, tt.args.modelOpt, tt.args.endpoint); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewService() =\n\t%+v\n, want\n\t%+v", got, tt.want)
			}
		})
	}
}
