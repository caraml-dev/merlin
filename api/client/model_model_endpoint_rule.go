/*
 * Merlin
 *
 * API Guide for accessing Merlin's model management, deployment, and serving functionalities
 *
 * API version: 0.14.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package client

type ModelEndpointRule struct {
	Destinations []ModelEndpointRuleDestination `json:"destinations,omitempty"`
	Mirror       *VersionEndpoint               `json:"mirror,omitempty"`
}
