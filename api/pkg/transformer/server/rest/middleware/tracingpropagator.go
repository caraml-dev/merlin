package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func TracingPropagatorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerCarier := propagation.HeaderCarrier(r.Header)
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), headerCarier)
		r2 := r.WithContext(ctx)
		next.ServeHTTP(w, r2)
	})
}
