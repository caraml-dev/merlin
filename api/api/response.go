// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
	"net/http"
)

// APIResponse handles responses of APIs.
type APIResponse struct {
	code int
	data interface{}
}

// Error represents the structure of an error response.
type Error struct {
	Message string `json:"error"`
}

// WriteTo writes the response header and body.
func (r *APIResponse) WriteTo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(r.code)

	if r.data != nil {
		encoder := json.NewEncoder(w)
		encoder.Encode(r.data)
	}
}

// Ok represents the response of status code 200.
func Ok(data interface{}) *APIResponse {
	return &APIResponse{
		code: http.StatusOK,
		data: data,
	}
}

// Created represents the response of status code 201.
func Created(data interface{}) *APIResponse {
	return &APIResponse{
		code: http.StatusCreated,
		data: data,
	}
}

// NoContent represents the response of status code 204.
func NoContent() *APIResponse {
	return &APIResponse{
		code: http.StatusNoContent,
	}
}

// NewError represents the response of a custom status code.
func NewError(code int, msg string) *APIResponse {
	return &APIResponse{
		code: code,
		data: Error{msg},
	}
}

// NotFound represents the response of status code 404.
func NotFound(msg string) *APIResponse {
	return NewError(http.StatusNotFound, msg)
}

// BadRequest represents the response of status code 400.
func BadRequest(msg string) *APIResponse {
	return NewError(http.StatusBadRequest, msg)
}

// InternalServerError represents the response of status code 500.
func InternalServerError(msg string) *APIResponse {
	return NewError(http.StatusInternalServerError, msg)
}

// Forbidden represents the response of status code 403.
func Forbidden(msg string) *APIResponse {
	return NewError(http.StatusForbidden, msg)
}
