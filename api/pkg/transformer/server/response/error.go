package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// NewError returns `*Error` instance.
func NewError(code int, err error) *Error {
	return &Error{
		Code:    code,
		Message: err.Error(),
	}
}

// Write calls `write` function to write error response.
func (e *Error) Write(w http.ResponseWriter) {
	write(w, e, e.Code) //nolint:errcheck
}

// Write writes final response.
func write(w http.ResponseWriter, d interface{}, c int) error {
	if c == 0 {
		return errors.New("0 is not a valid code")
	}

	if !bodyAllowed(c) {
		w.WriteHeader(c)
		return nil
	}

	b, e := json.Marshal(d)
	if e != nil {
		return e
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprint(len(b)))
	w.WriteHeader(c)

	if _, e := w.Write(b); e != nil {
		return e
	}

	return nil
}

// bodyAllowed reports whether a given response status code permits a body. See
// RFC 7230, section 3.3.
func bodyAllowed(status int) bool {
	if (status >= 100 && status <= 199) || status == 204 || status == 304 {
		return false
	}

	return true
}
