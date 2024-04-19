package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/caraml-dev/merlin/log"
	"github.com/caraml-dev/merlin/models"
)

type VersionImageController struct {
	*AppContext
}

func (c *VersionImageController) GetImage(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		return NotFound(fmt.Sprintf("Model not found: %v", err))
	}

	if err := c.validateModelType(model.Type); err != nil {
		return BadRequest(fmt.Sprintf("Invalid model type: %v", err))
	}

	versionID, _ := models.ParseID(vars["version_id"])
	version, err := c.VersionsService.FindByID(ctx, modelID, versionID, c.FeatureToggleConfig.MonitoringConfig)
	if err != nil {
		return NotFound(fmt.Sprintf("Model version not found: %v", err))
	}

	image, err := c.VersionImageService.GetImage(ctx, model, version)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error getting image: %v", err))
	}

	return Ok(image)
}

func (c *VersionImageController) BuildImage(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()

	modelID, _ := models.ParseID(vars["model_id"])
	model, err := c.ModelsService.FindByID(ctx, modelID)
	if err != nil {
		return NotFound(fmt.Sprintf("Model not found: %v", err))
	}

	if err := c.validateModelType(model.Type); err != nil {
		return BadRequest(fmt.Sprintf("Invalid model type: %v", err))
	}

	versionID, _ := models.ParseID(vars["version_id"])
	version, err := c.VersionsService.FindByID(ctx, modelID, versionID, c.FeatureToggleConfig.MonitoringConfig)
	if err != nil {
		return NotFound(fmt.Sprintf("Model version not found: %v", err))
	}

	options, ok := body.(*models.BuildImageOptions)
	if !ok {
		return BadRequest("Unable to parse request body")
	}

	// Image building might require a long time to complete, so we start the image building process asynchronously in a different Go routine
	// and return immediately with an accepted status.
	// We use a new context.Background() to ensure that the process is not cancelled when the request is finished
	go func() {
		_, err := c.VersionImageService.BuildImage(context.Background(), model, version, options)
		if err != nil {
			log.Errorf("Error building image: %v", err)
		}
	}()

	return Accepted(nil)
}

func (c *VersionImageController) validateModelType(modelType string) error {
	switch modelType {
	case models.ModelTypePyFunc, models.ModelTypePyFuncV2, models.ModelTypePyFuncV3:
		return nil
	default:
		return fmt.Errorf("model type %s is not supported", modelType)
	}
}
