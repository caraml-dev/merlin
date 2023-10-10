package cluster

import (
	"net/url"
	"testing"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/protocol"
)

func TestVirtualService_getModelVersionHost(t *testing.T) {
	defaultModelVersionRevisionURL, _ := url.Parse("http://test-model-1-1.test-namespace.caraml.dev")

	type fields struct {
		Name                    string
		Namespace               string
		VersionID               string
		RevisionID              models.ID
		Labels                  map[string]string
		Protocol                protocol.Protocol
		ModelVersionRevisionURL *url.URL
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "1",
			fields: fields{
				Name:                    "test-model-1",
				Namespace:               "test-namespace",
				VersionID:               "1",
				RevisionID:              models.ID(1),
				Labels:                  map[string]string{},
				Protocol:                protocol.HttpJson,
				ModelVersionRevisionURL: defaultModelVersionRevisionURL,
			},
			want:    "test-model-1.test-namespace.caraml.dev",
			wantErr: false,
		},
		{
			name: "2",
			fields: fields{
				Name:                    "test-model-1",
				Namespace:               "test-namespace",
				VersionID:               "1",
				RevisionID:              models.ID(1),
				Labels:                  map[string]string{},
				Protocol:                protocol.HttpJson,
				ModelVersionRevisionURL: defaultModelVersionRevisionURL,
			},
			want:    "test-model-1.test-namespace.caraml.dev",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &VirtualService{
				Name:                    tt.fields.Name,
				Namespace:               tt.fields.Namespace,
				VersionID:               tt.fields.VersionID,
				RevisionID:              tt.fields.RevisionID,
				Labels:                  tt.fields.Labels,
				Protocol:                tt.fields.Protocol,
				ModelVersionRevisionURL: tt.fields.ModelVersionRevisionURL,
			}
			got, err := cfg.getModelVersionHost()
			if (err != nil) != tt.wantErr {
				t.Errorf("VirtualService.getModelVersionHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VirtualService.getModelVersionHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVirtualService_getModelVersionRevisionHost(t *testing.T) {
	onlyHost, _ := url.Parse("http://test-model-1-1.test-namespace.caraml.dev")
	withPath, _ := url.Parse("http://test-model-1-1.test-namespace.caraml.dev/v1/models/test-model-1-1")

	type fields struct {
		Name                    string
		Namespace               string
		VersionID               string
		RevisionID              models.ID
		Labels                  map[string]string
		Protocol                protocol.Protocol
		ModelVersionRevisionURL *url.URL
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "1",
			fields: fields{
				Name:                    "test-model-1",
				Namespace:               "test-namespace",
				VersionID:               "1",
				RevisionID:              models.ID(1),
				Labels:                  map[string]string{},
				Protocol:                protocol.HttpJson,
				ModelVersionRevisionURL: onlyHost,
			},
			want:    "test-model-1-1.test-namespace.caraml.dev",
			wantErr: false,
		},
		{
			name: "2",
			fields: fields{
				Name:                    "test-model-1",
				Namespace:               "test-namespace",
				VersionID:               "1",
				RevisionID:              models.ID(1),
				Labels:                  map[string]string{},
				Protocol:                protocol.HttpJson,
				ModelVersionRevisionURL: withPath,
			},
			want:    "test-model-1-1.test-namespace.caraml.dev",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &VirtualService{
				Name:                    tt.fields.Name,
				Namespace:               tt.fields.Namespace,
				VersionID:               tt.fields.VersionID,
				RevisionID:              tt.fields.RevisionID,
				Labels:                  tt.fields.Labels,
				Protocol:                tt.fields.Protocol,
				ModelVersionRevisionURL: tt.fields.ModelVersionRevisionURL,
			}
			got, err := cfg.getModelVersionRevisionHost()
			if (err != nil) != tt.wantErr {
				t.Errorf("VirtualService.getModelVersionRevisionHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VirtualService.getModelVersionRevisionHost() = %v, want %v", got, tt.want)
			}
		})
	}
}
