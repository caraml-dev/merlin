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

	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/mitchellh/copystructure"
)

const (
	// TODO: Make these configurable
	knativeIngressGateway          = "knative-serving/knative-ingress-gateway"
	defaultIstioIngressGatewayHost = "istio-ingressgateway.istio-system.svc.cluster.local"
)

type VirtualService struct {
	Name                    string
	Namespace               string
	ModelName               string
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
		Name:                    fmt.Sprintf("%s-%s-%s", modelService.ModelName, modelService.ModelVersion, models.VirtualServiceComponentType),
		Namespace:               modelService.Namespace,
		ModelName:               modelService.ModelName,
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
	return fmt.Sprintf("%s-%s.%s.%s", cfg.ModelName, cfg.VersionID, cfg.Namespace, domain), nil
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
		// Default to application/json if Content-Type header is empty
		routeDestinationsWithContentType, err := copyRouteDestinations(routeDestinations)
		if err != nil {
			log.Errorf("failed to copy routeDestinations: %+v", err)
			return nil
		}
		routeDestinationsWithContentType[0].Headers.Request.Set["Content-Type"] = "application/json"

		uri := &istiov1beta1.StringMatch{
			MatchType: &istiov1beta1.StringMatch_Exact{
				Exact: fmt.Sprintf("/v1/models/%s-%s:predict", cfg.ModelName, cfg.VersionID),
			},
		}
		rewrite := &istiov1beta1.HTTPRewrite{
			Uri: fmt.Sprintf("%s:predict", modelVersionRevisionPath),
		}

		return []*istiov1beta1.HTTPRoute{
			{
				Match: []*istiov1beta1.HTTPMatchRequest{
					{
						Uri: uri,
						Headers: map[string]*istiov1beta1.StringMatch{
							"content-type": {},
						},
					},
				},
				Route:   routeDestinationsWithContentType,
				Rewrite: rewrite,
			},
			{
				Match: []*istiov1beta1.HTTPMatchRequest{
					{
						Uri: uri,
					},
				},
				Route:   routeDestinations,
				Rewrite: rewrite,
			},
			{
				Route: routeDestinations,
			},
		}
	}
}

// copyTopologySpreadConstraints copies the topology spread constraints using the service builder's as a template
func copyRouteDestinations(src []*istiov1beta1.HTTPRouteDestination) ([]*istiov1beta1.HTTPRouteDestination, error) {
	destRaw, err := copystructure.Copy(src)
	if err != nil {
		return nil, fmt.Errorf("error copying []*HTTPRouteDestination: %w", err)
	}

	dest, ok := destRaw.([]*istiov1beta1.HTTPRouteDestination)
	if !ok {
		return nil, fmt.Errorf("error in type assertion of copied []*HTTPRouteDestination interface: %w", err)
	}

	return dest, nil
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

	forceEnabled := true

	return c.istioClient.
		VirtualServices(vsSpec.Namespace).
		Patch(ctx, vsCfg.Name, types.ApplyPatchType, vsJSON, metav1.PatchOptions{FieldManager: "application/apply-patch", Force: &forceEnabled})
}

func (c *controller) deleteVirtualService(ctx context.Context, name, namespace string) error {
	return c.istioClient.VirtualServices(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
