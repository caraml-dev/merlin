package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/caraml-dev/merlin/models"
	mErrors "github.com/caraml-dev/merlin/pkg/errors"
)

type ModelSchemaController struct {
	*AppContext
}

func (m *ModelSchemaController) GetAllSchemas(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()
	modelID, _ := models.ParseID(vars["model_id"])
	modelSchemas, err := m.ModelSchemaService.List(ctx, modelID)
	if err != nil {
		if errors.Is(err, mErrors.ErrNotFound) {
			return NotFound(fmt.Sprintf("Model schemas not found: %v", err))
		}
		return InternalServerError(fmt.Sprintf("Error get All schemas with model id: %d with error: %v", modelID, err))
	}
	return Ok(modelSchemas)
}

func (m *ModelSchemaController) GetSchema(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()
	modelID, _ := models.ParseID(vars["model_id"])
	modelSchemaID, _ := models.ParseID(vars["schema_id"])
	modelSchema, err := m.ModelSchemaService.FindByID(ctx, modelSchemaID, modelID)
	if err != nil {
		if errors.Is(err, mErrors.ErrNotFound) {
			return NotFound(fmt.Sprintf("Model schema with id: %d not found: %v", modelSchemaID, err))
		}
		return InternalServerError(fmt.Sprintf("Error get schema with id: %d, model id: %d and error: %v", modelSchemaID, modelID, err))
	}

	return Ok(modelSchema)
}

func (m *ModelSchemaController) CreateOrUpdateSchema(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()
	modelID, _ := models.ParseID(vars["model_id"])

	modelSchema, ok := body.(*models.ModelSchema)
	if !ok {
		return BadRequest("Unable to parse request body")
	}
	modelSchema.ModelID = modelID
	schema, err := m.ModelSchemaService.Save(ctx, modelSchema)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error save model schema: %v", err))
	}
	return Ok(schema)
}

func (m *ModelSchemaController) DeleteSchema(r *http.Request, vars map[string]string, _ interface{}) *Response {
	ctx := r.Context()
	modelID, _ := models.ParseID(vars["model_id"])
	modelSchemaID, _ := models.ParseID(vars["schema_id"])
	modelSchema := &models.ModelSchema{ID: modelSchemaID, ModelID: modelID}
	if err := m.ModelSchemaService.Delete(ctx, modelSchema); err != nil {
		return InternalServerError(fmt.Sprintf("Error delete model schema: %v", err))
	}
	return NoContent()
}
