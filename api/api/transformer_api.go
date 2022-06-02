package api

import (
	"net/http"

	"github.com/gojek/merlin/models"
)

// TransformerController
type TransformerController struct {
	*AppContext
}

// SimulateTransformer API handles simulation of standard transformer
func (c *TransformerController) SimulateTransformer(r *http.Request, vars map[string]string, body interface{}) *Response {
	ctx := r.Context()
	simulationPayload, ok := body.(*models.TransformerSimulation)
	if !ok {
		return BadRequest("Unable to parse request body")
	}
	transformerResult, err := c.TransformerService.SimulateTransformer(ctx, simulationPayload)
	if err != nil {
		return InternalServerError(err.Error())
	}
	return Ok(transformerResult)
}
