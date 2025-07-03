package handler

import (
	"fmt"
	"net/http"

	"github.com/alexchebotarsky/thermofridge-api/openapi"
)

func OpenapiYAML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(openapi.OpenapiYAML)))
	_, err := w.Write(openapi.OpenapiYAML)
	handleWritingErr(err)
}

func SwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(openapi.SwaggerHTML)))
	_, err := w.Write(openapi.SwaggerHTML)
	handleWritingErr(err)
}
