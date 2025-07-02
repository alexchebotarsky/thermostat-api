package handler

import (
	"fmt"
	"log/slog"
)

func handleWritingErr(err error) {
	if err != nil {
		slog.Error(fmt.Sprintf("Error writing to http.ResponseWriter: %v", err))
	}
}

type errorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"statusCode"`
}
