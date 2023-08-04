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
	"strings"

	"github.com/caraml-dev/merlin/utils"
)

// Response handles responses of APIs.
type Response struct {
	code    int
	data    interface{}
	headers map[string]string

	// Debug info
	stacktrace string
}

// Error represents the structure of an error response.
type Error struct {
	Message string `json:"error"`
}

// WriteTo writes the response header and body.
func (r *Response) WriteTo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	exposeHeaders := make([]string, 0, len(r.headers))
	for key, value := range r.headers {
		exposeHeaders = append(exposeHeaders, key)
		w.Header().Set(key, value)
	}

	allowHeaders := strings.Join(exposeHeaders, ",")
	w.Header().Set("Access-Control-Expose-Headers", allowHeaders)

	w.WriteHeader(r.code)

	if r.data != nil {
		encoder := json.NewEncoder(w)
		encoder.Encode(r.data) //nolint:errcheck
	}
}

// IsSuccess returns true if the code is in the 2XX range
func (r *Response) IsSuccess() bool {
	return r.code >= http.StatusOK && r.code < http.StatusMultipleChoices
}

// Ok represents the response of status code 200.
func Ok(data interface{}) *Response {
	return &Response{
		code: http.StatusOK,
		data: data,
	}
}

// OkWithHeaders represents the response of status code 200 with custom headers
func OkWithHeaders(data interface{}, headers map[string]string) *Response {
	return &Response{
		code:    http.StatusOK,
		data:    data,
		headers: headers,
	}
}

// Created represents the response of status code 201.
func Created(data interface{}) *Response {
	return &Response{
		code: http.StatusCreated,
		data: data,
	}
}

// NoContent represents the response of status code 204.
func NoContent() *Response {
	return &Response{
		code:       http.StatusNoContent,
		stacktrace: utils.StackTrace(1),
	}
}

// NewError represents the response of a custom status code.
func NewError(code int, msg string) *Response {
	return &Response{
		code:       code,
		data:       Error{msg},
		stacktrace: utils.StackTrace(1),
	}
}

// NotFound represents the response of status code 404.
func NotFound(msg string) *Response {
	return &Response{
		code:       http.StatusNotFound,
		data:       Error{msg},
		stacktrace: utils.StackTrace(1),
	}
}

// BadRequest represents the response of status code 400.
func BadRequest(msg string) *Response {
	return &Response{
		code:       http.StatusBadRequest,
		data:       Error{msg},
		stacktrace: utils.StackTrace(1),
	}
}

// InternalServerError represents the response of status code 500.
func InternalServerError(msg string) *Response {
	return &Response{
		code:       http.StatusInternalServerError,
		data:       Error{msg},
		stacktrace: utils.StackTrace(1),
	}
}

// Forbidden represents the response of status code 403.
func Forbidden(msg string) *Response {
	return &Response{
		code:       http.StatusForbidden,
		data:       Error{msg},
		stacktrace: utils.StackTrace(1),
	}
}
