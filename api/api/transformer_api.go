package api

import (
	"fmt"
	"net/http"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/protocol"
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

	if simulationPayload.Protocol != protocol.HttpJson && simulationPayload.Protocol != protocol.UpiV1 {
		return BadRequest(`The only supported protocol are "HTTP_JSON" and "UPI_V1"`)
	}
	transformerResult, err := c.TransformerService.SimulateTransformer(ctx, simulationPayload)
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error during simulation: %v", err))
	}

	return Ok(transformerResult)
}
