package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func HandleError(w http.ResponseWriter, handlerErr error, statusCode int, shouldLog bool) {
	if shouldLog {
		slog.Error(fmt.Sprintf("Handler error: %v", handlerErr), "status", statusCode)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Hide server errors from the client
	returnedErr := handlerErr
	if statusCode >= 500 {
		returnedErr = fmt.Errorf("%s: %d", http.StatusText(statusCode), statusCode)
	}

	err := json.NewEncoder(w).Encode(errorResponse{
		Error:      returnedErr.Error(),
		StatusCode: statusCode,
	})
	handleWritingErr(err)
}
