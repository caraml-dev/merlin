package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/caraml-dev/merlin/models"
	mErrors "github.com/caraml-dev/merlin/pkg/errors"
)

// ModelSchemaController
type ModelSchemaController struct {
	*AppContext
}

// GetAllSchemas list all model schemas given model ID
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

// GetSchema get detail of a model schema given the schema id and model id
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

// CreateOrUpdateSchema upsert schema
// If ID is not defined it will create new model schema
// If ID is defined but not exist, it will create new model schema
// If ID is defined and exist, it will update the existing model schema associated with that ID
func (m *ModelSchemaController) CreateOrUpdateSchema(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()
	modelID, _ := models.ParseID(vars["model_id"])

	modelSchema, ok := body.(*models.ModelSchema)
	if !ok {
		return BadRequest("Unable to parse request body")
	}

	if modelSchema.ModelID > 0 && modelSchema.ModelID != modelID {
		return BadRequest("Mismatch model id between request path and body")
	}

	modelSchema.ModelID = modelID
	schema, err := m.ModelSchemaService.Save(ctx, modelSchema)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error save model schema: %v", err))
	}
	return Ok(schema)
}

// DeleteSchema delete model schema given schema id and model id
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
