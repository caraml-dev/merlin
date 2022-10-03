package api

import (
	"net/http"

	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/pkg/protocol"
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
		log.Errorf("Unable to parse request body %v", body)
		return BadRequest("Unable to parse request body")
	}

	if simulationPayload.Protocol != protocol.HttpJson && simulationPayload.Protocol != protocol.UpiV1 {
		return BadRequest(`The only supported protocol are "HTTP_JSON" and "UPI_V1"`)
	}
	transformerResult, err := c.TransformerService.SimulateTransformer(ctx, simulationPayload)
	if err != nil {
		log.Errorf("Failed performing transfomer simulation %v", err)
		return InternalServerError(err.Error())
	}

	return Ok(transformerResult)
}
