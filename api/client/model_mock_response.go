/*
 * Merlin
 *
 * API Guide for accessing Merlin's model management, deployment, and serving functionalities
 *
 * API version: 0.14.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package client

type MockResponse struct {
	Body    *ModelMap `json:"body,omitempty"`
	Headers *ModelMap `json:"headers,omitempty"`
}
