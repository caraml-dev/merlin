package api

import (
	"fmt"
	"net/http"
)

type DeploymentController struct {
	*AppContext
}

func (c *DeploymentController) ListDeployments(r *http.Request, vars map[string]string, _ interface{}) *Response {
	deployments, err := c.DeploymentService.ListDeployments(vars["model_id"], vars["version_id"], vars["endpoint_id"])
	if err != nil {
		return InternalServerError(fmt.Sprintf("Error listing deployments: %v", err))
	}

	return Ok(deployments)
}
