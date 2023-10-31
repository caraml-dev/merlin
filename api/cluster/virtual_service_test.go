package cluster

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/protocol"
	istiov1beta1 "istio.io/api/networking/v1beta1"
)

func TestVirtualService_getModelVersionHost(t *testing.T) {
	defaultModelVersionRevisionURL, _ := url.Parse("http://test-model-1-1.test-namespace.caraml.dev")

	type fields struct {
		Name                    string
		Namespace               string
		ModelName               string
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
				ModelName:               "test-model",
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
				ModelName:               "test-model",
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
				ModelName:               tt.fields.ModelName,
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

func TestVirtualService_createHttpRoutes(t *testing.T) {
	type fields struct {
		Name                    string
		Namespace               string
		ModelName               string
		VersionID               string
		RevisionID              models.ID
		Labels                  map[string]string
		Protocol                protocol.Protocol
		ModelVersionRevisionURL *url.URL
	}
	type args struct {
		modelVersionRevisionHost string
		modelVersionRevisionPath string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*istiov1beta1.HTTPRoute
	}{
		{
			name: "http",
			fields: fields{
				Name:      "test-model-1",
				ModelName: "test-model",
				VersionID: "1",
				Protocol:  protocol.HttpJson,
			},
			args: args{
				modelVersionRevisionHost: "test-model-1-1.test-namespace.caraml.dev",
				modelVersionRevisionPath: "/v1/models/test-model-1-1",
			},
			want: []*istiov1beta1.HTTPRoute{
				{
					Match: []*istiov1beta1.HTTPMatchRequest{
						{
							Uri: &istiov1beta1.StringMatch{
								MatchType: &istiov1beta1.StringMatch_Exact{
									Exact: "/v1/models/test-model-1:predict",
								},
							},
							Headers: map[string]*istiov1beta1.StringMatch{
								"content-type": {},
							},
						},
					},
					Route: []*istiov1beta1.HTTPRouteDestination{
						{
							Destination: &istiov1beta1.Destination{
								Host: defaultIstioIngressGatewayHost,
							},
							Headers: &istiov1beta1.Headers{
								Request: &istiov1beta1.Headers_HeaderOperations{
									Set: map[string]string{
										"Content-Type": "application/json",
										"Host":         "test-model-1-1.test-namespace.caraml.dev",
									},
								},
							},
							Weight: 100,
						},
					},
					Rewrite: &istiov1beta1.HTTPRewrite{
						Uri: fmt.Sprintf("%s:predict", "/v1/models/test-model-1-1"),
					},
				},
				{
					Match: []*istiov1beta1.HTTPMatchRequest{
						{
							Uri: &istiov1beta1.StringMatch{
								MatchType: &istiov1beta1.StringMatch_Exact{
									Exact: "/v1/models/test-model-1:predict",
								},
							},
						},
					},
					Route: []*istiov1beta1.HTTPRouteDestination{
						{
							Destination: &istiov1beta1.Destination{
								Host: defaultIstioIngressGatewayHost,
							},
							Headers: &istiov1beta1.Headers{
								Request: &istiov1beta1.Headers_HeaderOperations{
									Set: map[string]string{
										"Host": "test-model-1-1.test-namespace.caraml.dev",
									},
								},
							},
							Weight: 100,
						},
					},
					Rewrite: &istiov1beta1.HTTPRewrite{
						Uri: fmt.Sprintf("%s:predict", "/v1/models/test-model-1-1"),
					},
				},
				{
					Route: []*istiov1beta1.HTTPRouteDestination{
						{
							Destination: &istiov1beta1.Destination{
								Host: defaultIstioIngressGatewayHost,
							},
							Headers: &istiov1beta1.Headers{
								Request: &istiov1beta1.Headers_HeaderOperations{
									Set: map[string]string{
										"Host": "test-model-1-1.test-namespace.caraml.dev",
									},
								},
							},
							Weight: 100,
						},
					},
				},
			},
		},
		{
			name: "upi",
			fields: fields{
				Name:      "test-model-1",
				ModelName: "test-model",
				VersionID: "1",
				Protocol:  protocol.UpiV1,
			},
			args: args{
				modelVersionRevisionHost: "test-model-1-1.test-namespace.caraml.dev",
				modelVersionRevisionPath: "/v1/models/test-model-1-1",
			},
			want: []*istiov1beta1.HTTPRoute{
				{
					Route: []*istiov1beta1.HTTPRouteDestination{
						{
							Destination: &istiov1beta1.Destination{
								Host: defaultIstioIngressGatewayHost,
							},
							Headers: &istiov1beta1.Headers{
								Request: &istiov1beta1.Headers_HeaderOperations{
									Set: map[string]string{
										"Host": "test-model-1-1.test-namespace.caraml.dev",
									},
								},
							},
							Weight: 100,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &VirtualService{
				Name:                    tt.fields.Name,
				Namespace:               tt.fields.Namespace,
				ModelName:               tt.fields.ModelName,
				VersionID:               tt.fields.VersionID,
				RevisionID:              tt.fields.RevisionID,
				Labels:                  tt.fields.Labels,
				Protocol:                tt.fields.Protocol,
				ModelVersionRevisionURL: tt.fields.ModelVersionRevisionURL,
			}
			if got := cfg.createHttpRoutes(tt.args.modelVersionRevisionHost, tt.args.modelVersionRevisionPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VirtualService.createHttpRoutes() =\n%v\n, want\n%v", got, tt.want)
			}
		})
	}
}
