package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alexchebotarsky/thermofridge-api/metrics"
	chi "github.com/go-chi/chi/v5"
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		crw := customResponseWriter{ResponseWriter: w}

		start := time.Now()
		next.ServeHTTP(&crw, r)
		duration := time.Since(start)

		routeName := fmt.Sprintf("%s %s", r.Method, chi.RouteContext(r.Context()).RoutePattern())

		metrics.AddRequestHandled(routeName, crw.status)
		metrics.ObserveRequestDuration(duration)
	})
}

type customResponseWriter struct {
	http.ResponseWriter
	status int
}

func (crw *customResponseWriter) WriteHeader(status int) {
	crw.status = status
	crw.ResponseWriter.WriteHeader(status)
}
