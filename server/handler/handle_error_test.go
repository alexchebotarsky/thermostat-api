package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHandleError(t *testing.T) {
	type args struct {
		err        error
		statusCode int
		req        *http.Request
	}
	tests := []struct {
		name     string
		args     args
		wantBody errorResponse
	}{
		{
			name: "should return error response 400, with error message",
			args: args{
				err:        errors.New("Explanation of the error"),
				statusCode: http.StatusBadRequest,
				req:        httptest.NewRequest(http.MethodGet, "/api/v1/test", nil),
			},
			wantBody: errorResponse{
				Error:      "Explanation of the error",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name: "should return error response 500, with generic error message",
			args: args{
				err:        errors.New("Explanation of the error"),
				statusCode: http.StatusInternalServerError,
				req:        httptest.NewRequest(http.MethodGet, "/api/v1/test", nil),
			},
			wantBody: errorResponse{
				Error:      "Internal Server Error: 500",
				StatusCode: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleError(w, tt.args.err, tt.args.statusCode, false)

			if w.Code != tt.args.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.args.statusCode, w.Code)
			}

			// Decode the response body into struct for comparison
			var resBody errorResponse
			if err := json.NewDecoder(w.Body).Decode(&resBody); err != nil {
				t.Errorf("HandleError() error json decoding response body: %v", err)
			}

			if !reflect.DeepEqual(resBody, tt.wantBody) {
				t.Errorf("HandleError() response body = %v, want %v", resBody, tt.wantBody)
			}
		})
	}
}
