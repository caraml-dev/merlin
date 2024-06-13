package api

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/caraml-dev/merlin/config"
	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/caraml-dev/merlin/service"
	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
)

type requestValidator interface {
	validate() error
}

type funcValidate struct {
	f func() error
}

func newFuncValidate(f func() error) *funcValidate {
	return &funcValidate{
		f: f,
	}
}

func (fv *funcValidate) validate() error {
	return fv.f()
}

var supportedUPIModelTypes = map[string]bool{
	models.ModelTypePyFunc: true,
	models.ModelTypeCustom: true,
}

var supportedObservabilityModelTypes = []string{
	models.ModelTypePyFunc,
	models.ModelTypeXgboost,
}

var ErrUnsupportedObservabilityModelType = errors.New("observability cannot be enabled not for this model type")

func isModelSupportUPI(model *models.Model) bool {
	_, isSupported := supportedUPIModelTypes[model.Type]

	return isSupported
}

func validateRequest(validators ...requestValidator) error {
	for _, validator := range validators {
		if err := validator.validate(); err != nil {
			return err
		}
	}
	return nil
}

func resourceRequestValidation(endpoint *models.VersionEndpoint) requestValidator {
	return newFuncValidate(func() error {
		if endpoint.ResourceRequest == nil {
			return nil
		}

		if endpoint.ResourceRequest.MinReplica > endpoint.ResourceRequest.MaxReplica {
			return fmt.Errorf("min replica must be less or equal to max replica")
		}

		if endpoint.ResourceRequest.MaxReplica < 1 {
			return fmt.Errorf("max replica must be greater than 0")
		}

		return nil
	})
}

func customModelValidation(model *models.Model, version *models.Version) requestValidator {
	return newFuncValidate(func() error {
		if model.Type == models.ModelTypeCustom {
			if err := validateCustomPredictor(version); err != nil {
				return err
			}
		}
		return nil
	})
}

func upiModelValidation(model *models.Model, endpointProtocol protocol.Protocol) requestValidator {
	return newFuncValidate(func() error {
		if !isModelSupportUPI(model) && endpointProtocol == protocol.UpiV1 {
			return fmt.Errorf("%s model is not supported by UPI", model.Type)
		}
		return nil
	})
}

func newVersionEndpointValidation(version *models.Version, envName string) requestValidator {
	return newFuncValidate(func() error {
		endpoint, ok := version.GetEndpointByEnvironmentName(envName)
		if ok && (endpoint.IsRunning() || endpoint.IsServing()) {
			return fmt.Errorf("there is `%s` deployment for the model version", endpoint.Status)
		}
		return nil
	})
}

func deploymentQuotaValidation(ctx context.Context, model *models.Model, env *models.Environment, endpointSvc service.EndpointsService) requestValidator {
	return newFuncValidate(func() error {
		deployedModelVersionCount, err := endpointSvc.CountEndpoints(ctx, env, model)
		if err != nil {
			return fmt.Errorf("unable to count number of endpoints in env %s: %w", env.Name, err)
		}

		if deployedModelVersionCount >= config.MaxDeployedVersion {
			return fmt.Errorf("max deployed endpoint reached. Max: %d Current: %d, undeploy existing endpoint before continuing", config.MaxDeployedVersion, deployedModelVersionCount)
		}
		return nil
	})
}

func transformerValidation(
	ctx context.Context,
	endpoint *models.VersionEndpoint,
	stdTransformerCfg config.StandardTransformerConfig,
	feastCore core.CoreServiceClient,
) requestValidator {
	return newFuncValidate(func() error {
		if endpoint.Transformer != nil && endpoint.Transformer.Enabled {
			err := validateTransformer(ctx, endpoint, stdTransformerCfg, feastCore)
			if err != nil {
				return fmt.Errorf("Error validating transformer: %w", err)
			}
		}
		return nil
	})
}

func updateRequestValidation(prev *models.VersionEndpoint, new *models.VersionEndpoint) requestValidator {
	return newFuncValidate(func() error {
		if prev.EnvironmentName != new.EnvironmentName {
			return fmt.Errorf("updating environment is not allowed, previous: %s, new: %s", prev.EnvironmentName, new.EnvironmentName)
		}

		if prev.Status == models.EndpointPending {
			return fmt.Errorf("updating endpoint status to %s is not allowed when the endpoint is currently in the pending state", new.Status)
		}

		if new.Status != prev.Status {
			if prev.Status == models.EndpointServing {
				return fmt.Errorf("updating endpoint status to %s is not allowed when the endpoint is currently in the serving state", new.Status)
			}

			if new.Status != models.EndpointRunning && new.Status != models.EndpointTerminated {
				return fmt.Errorf("updating endpoint status to %s is not allowed", new.Status)
			}
		}
		return nil
	})
}

func deploymentModeValidation(prev *models.VersionEndpoint, new *models.VersionEndpoint) requestValidator {
	return newFuncValidate(func() error {
		// Should not allow changing the deployment mode of a pending/running/serving model for 2 reasons:
		// * For "serving" models it's risky as, we can't guarantee graceful re-deployment
		// * Kserve uses slightly different deployment resource naming under the hood and doesn't clean up the older deployment
		if (prev.IsRunning() || prev.IsServing()) && new.DeploymentMode != "" &&
			new.DeploymentMode != prev.DeploymentMode {
			return fmt.Errorf("changing deployment type of a %s model is not allowed, please terminate it first", prev.Status)
		}
		return nil
	})
}

func modelObservabilityValidation(endpoint *models.VersionEndpoint, model *models.Model) requestValidator {
	return newFuncValidate(func() error {
		if endpoint.EnableModelObservability && !slices.Contains(supportedObservabilityModelTypes, model.Type) {
			return fmt.Errorf("%s: %w", model.Type, ErrUnsupportedObservabilityModelType)
		}
		return nil
	})
}
