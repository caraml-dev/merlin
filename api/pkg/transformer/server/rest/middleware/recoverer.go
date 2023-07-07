package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/caraml-dev/merlin/pkg/transformer/server/response"
)

func RecoveryHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				response.NewError(http.StatusInternalServerError, fmt.Errorf("panic: %v", err)).Write(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
