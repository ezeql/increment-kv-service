package handlers

import (
	"net/http"

	"github.com/ezeql/appcues-increment-simple/internal/incrementsrv"

	"github.com/go-chi/render"
)

type IncrementResource struct {
	Service *incrementsrv.Publisher
}

//CreateIncrement unmarshalls the JSON payload, validates it and publish to a messaging queue
func (h *IncrementResource) CreateIncrement(w http.ResponseWriter, r *http.Request) {

	// parse and validate user input from JSON Payload
	inc, err := incrementsrv.InputFromJSONReader(r.Body)

	if err != nil { // invalid request
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, IncrementResponse{ErrorDesc: err.Error()})
		return
	}

	// request is valid, push it into the messaging queue
	err = h.Service.Publish(*inc)

	if err != nil { //something is wrong with the messaing queue.
		// we might improve the error handling here by detecting what kind of error are we getting?
		// is it permanent or temporary? a retry mechanism would come handy here.
		w.WriteHeader(http.StatusServiceUnavailable)
		render.JSON(w, r, IncrementResponse{ErrorDesc: err.Error()})
		return
	}

	// all good.
	w.WriteHeader(http.StatusOK)
}
