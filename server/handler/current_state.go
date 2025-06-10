package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alexchebotarsky/thermofridge-api/client"
	"github.com/alexchebotarsky/thermofridge-api/model/thermofridge"
)

type CurrentStateFetcher interface {
	FetchCurrentState() (*thermofridge.CurrentState, error)
}

func GetCurrentState(fetcher CurrentStateFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := fetcher.FetchCurrentState()
		if err != nil {
			switch err.(type) {
			case *client.ErrNotFound:
				HandleError(w, fmt.Errorf("current state not found: %v", err), http.StatusNotFound, true)
			default:
				HandleError(w, fmt.Errorf("error fetching current state: %v", err), http.StatusInternalServerError, true)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(state)
		handleWritingErr(err)
	}
}
