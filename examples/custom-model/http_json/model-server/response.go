package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Response struct {
	code    int
	data    interface{}
	headers map[string]string
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
		encoder.Encode(r.data)
	}
}

// Ok represents the response of status code 200.
func Ok(data interface{}) *Response {
	return &Response{
		code: http.StatusOK,
		data: data,
	}
}

// NewError represents the response of a custom status code.
func NewError(code int, msg string) *Response {
	return &Response{
		code: code,
		data: Error{msg},
	}
}

// InternalServerError represents the response of status code 500.
func InternalServerError(msg string) *Response {
	return NewError(http.StatusInternalServerError, msg)
}

// BadRequest represents the response of status code 500.
func BadRequest(msg string) *Response {
	return NewError(http.StatusBadRequest, msg)
}
