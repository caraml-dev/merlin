/*
 * Merlin
 *
 * API Guide for accessing Merlin's model management, deployment, and serving functionalities
 *
 * API version: 0.7.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package client

type LoggerMode string

// List of LoggerMode
const (
	AllLoggerMode      LoggerMode = "all"
	RequestLoggerMode  LoggerMode = "request"
	ResponseLoggerMode LoggerMode = "response"
)
