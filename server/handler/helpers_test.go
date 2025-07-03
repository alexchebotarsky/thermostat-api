package handler

import (
	"context"
	"net/http"

	chi "github.com/go-chi/chi/v5"
)

func addChiURLParams(req *http.Request, params map[string]string) *http.Request {
	chiCtx := chi.NewRouteContext()
	for k, v := range params {
		chiCtx.URLParams.Add(k, v)
	}

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
}
