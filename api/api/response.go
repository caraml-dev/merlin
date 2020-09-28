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

type ApiResponse struct {
	code int
	data interface{}
}

type Error struct {
	Message string `json:"error"`
}

func (r *ApiResponse) WriteTo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(r.code)

	if r.data != nil {
		encoder := json.NewEncoder(w)
		encoder.Encode(r.data)
	}
}

func Ok(data interface{}) *ApiResponse {
	return &ApiResponse{
		code: http.StatusOK,
		data: data,
	}
}

func Created(data interface{}) *ApiResponse {
	return &ApiResponse{
		code: http.StatusCreated,
		data: data,
	}
}

func NoContent() *ApiResponse {
	return &ApiResponse{
		code: http.StatusNoContent,
	}
}

func NewError(code int, msg string) *ApiResponse {
	return &ApiResponse{
		code: code,
		data: Error{msg},
	}
}

func NotFound(msg string) *ApiResponse {
	return NewError(http.StatusNotFound, msg)
}

func BadRequest(msg string) *ApiResponse {
	return NewError(http.StatusBadRequest, msg)
}

func InternalServerError(msg string) *ApiResponse {
	return NewError(http.StatusInternalServerError, msg)
}

func Forbidden(msg string) *ApiResponse {
	return NewError(http.StatusForbidden, msg)
}
