package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	istiov1beta1 "istio.io/api/networking/v1beta1"
	v1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/protocol"
)

const (
	knativeIngressGateway          = "knative-serving/knative-ingress-gateway"
	defaultIstioIngressGatewayHost = "istio-ingressgateway.istio-system.svc.cluster.local"
)

type VirtualService struct {
	Name                    string
	Namespace               string
	VersionID               string
	RevisionID              models.ID
	Labels                  map[string]string
	Protocol                protocol.Protocol
	ModelVersionRevisionURL *url.URL
}

func NewVirtualService(modelService *models.Service, isvcURL string) (*VirtualService, error) {
	modelVersionRevisionURL, err := url.Parse(isvcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse model version revision url: %s", isvcURL)
	}

	if modelVersionRevisionURL.Scheme == "" {
		veURL := "//" + isvcURL
		modelVersionRevisionURL, err = url.Parse(veURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse model version revision url: %s", isvcURL)
		}
	}

	return &VirtualService{
		Name:                    fmt.Sprintf("%s-%s", modelService.ModelName, modelService.ModelVersion),
		Namespace:               modelService.Namespace,
		VersionID:               modelService.ModelVersion,
		RevisionID:              modelService.RevisionID,
		Labels:                  modelService.Metadata.ToLabel(),
		Protocol:                modelService.Protocol,
		ModelVersionRevisionURL: modelVersionRevisionURL,
	}, nil
}

func (cfg VirtualService) BuildVirtualServiceSpec() (*v1beta1.VirtualService, error) {
	modelVersionHost, err := cfg.getModelVersionHost()
	if err != nil {
		return nil, err
	}

	// We manually create model version revision to avoid Kserve's modification to transformer url if the name contains "transformer"
	// E.g. given a model version: std-transformer-s-157.namespace.caraml.dev
	// The isvc.Status.URL would be: std-s-157-3-transformer.namespace.caraml.dev
	// modelVersionRevisionHost, err := cfg.getModelVersionRevisionHost()
	// if err != nil {
	// 	return nil, err
	// }

	modelVersionRevisionHost := cfg.ModelVersionRevisionURL.Hostname()
	modelVersionRevisionPath := cfg.ModelVersionRevisionURL.Path

	vs := &v1beta1.VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1beta1",
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    cfg.Labels,
		},
		Spec: istiov1beta1.VirtualService{
			Gateways: []string{knativeIngressGateway},
			Hosts:    []string{modelVersionHost},
			Http:     cfg.createHttpRoutes(modelVersionRevisionHost, modelVersionRevisionPath),
		},
	}

	return vs, nil
}

// getModelVersionHost creates model version endpoint host based on version endpoint's url
func (cfg *VirtualService) getModelVersionHost() (string, error) {
	host := strings.Split(cfg.ModelVersionRevisionURL.Hostname(), fmt.Sprintf(".%s.", cfg.Namespace))
	if len(host) != 2 {
		return "", fmt.Errorf("invalid version endpoint url: %s. failed to split domain: %+v", cfg.ModelVersionRevisionURL, host)
	}

	domain := host[1]
	return fmt.Sprintf("%s.%s.%s", cfg.Name, cfg.Namespace, domain), nil
}

// getModelVersionRevisionHost creates model version endpoint host based on version endpoint's url
func (cfg *VirtualService) getModelVersionRevisionHost() (string, error) {
	host := strings.Split(cfg.ModelVersionRevisionURL.Hostname(), fmt.Sprintf(".%s.", cfg.Namespace))
	if len(host) != 2 {
		return "", fmt.Errorf("invalid version endpoint url: %s. failed to split domain: %+v", cfg.ModelVersionRevisionURL, host)
	}

	domain := host[1]
	return fmt.Sprintf("%s-%s.%s.%s", cfg.Name, cfg.RevisionID, cfg.Namespace, domain), nil
}

func (cfg *VirtualService) createHttpRoutes(modelVersionRevisionHost, modelVersionRevisionPath string) []*istiov1beta1.HTTPRoute {
	routeDestinations := []*istiov1beta1.HTTPRouteDestination{
		{
			Destination: &istiov1beta1.Destination{
				Host: defaultIstioIngressGatewayHost,
			},
			Headers: &istiov1beta1.Headers{
				Request: &istiov1beta1.Headers_HeaderOperations{
					Set: map[string]string{
						"Host": modelVersionRevisionHost,
					},
				},
			},
			Weight: 100,
		},
	}

	switch cfg.Protocol {
	case protocol.UpiV1:
		return []*istiov1beta1.HTTPRoute{
			{
				Route: routeDestinations,
			},
		}

	default:
		// TODO: Remove setting Content-Type header after all clients are updated to set the header
		routeDestinations[0].Headers.Request.Add = map[string]string{
			"Content-Type": "application/json",
		}

		return []*istiov1beta1.HTTPRoute{
			{
				Route: routeDestinations,
				Match: []*istiov1beta1.HTTPMatchRequest{
					{
						Uri: &istiov1beta1.StringMatch{
							MatchType: &istiov1beta1.StringMatch_Exact{
								Exact: fmt.Sprintf("/v1/models/%s:predict", cfg.Name),
							},
						},
					},
				},
				Rewrite: &istiov1beta1.HTTPRewrite{
					Uri: fmt.Sprintf("%s:predict", modelVersionRevisionPath),
				},
			},
			{
				Route: routeDestinations,
			},
		}
	}
}

func (c *controller) deployVirtualService(ctx context.Context, vsCfg *VirtualService) (*v1beta1.VirtualService, error) {
	vsSpec, err := vsCfg.BuildVirtualServiceSpec()
	if err != nil {
		return nil, err
	}

	vsJSON, err := json.Marshal(vsSpec)
	if err != nil {
		return nil, err
	}

	return c.istioClient.VirtualServices(vsSpec.Namespace).Patch(ctx, vsCfg.Name, types.ApplyPatchType, vsJSON, metav1.PatchOptions{FieldManager: "application/apply-patch"})
}

func (c *controller) deleteVirtualService(ctx context.Context, name, namespace string) error {
	return c.istioClient.VirtualServices(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
