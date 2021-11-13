package transports

import (
	"context"
	"encoding/json"
	"errors"
	"learn/endpoints"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

var (
	ErrorBadRequest = errors.New("invalid request parameter")
)

func decodeArithmeticRequest(_ context.Context, r *http.Request) (interface{}, error) {

	vars := mux.Vars(r)

	requestType, ok := vars["type"]
	if !ok {
		return nil, ErrorBadRequest
	}

	pa, ok := vars["a"]

	if !ok {
		return nil, ErrorBadRequest
	}

	pb, ok := vars["b"]

	if !ok {
		return nil, ErrorBadRequest
	}
	a, _ := strconv.Atoi(pa)
	b, _ := strconv.Atoi(pb)

	return endpoints.ArithmeticRequest{
		RequestType: requestType,
		A:           a,
		B:           b,
	}, nil
}

func encodeArithmeticResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func MakeHttpHandler(ctx context.Context, endpoint endpoint.Endpoint, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(kithttp.DefaultErrorEncoder),
	}

	r.Methods("POST").Path("/calculate/{type}/{a}/{b}").Handler(kithttp.NewServer(
		endpoint,
		decodeArithmeticRequest,
		encodeArithmeticResponse,
		options...,
	))
	r.Path("/metrics").Handler(promhttp.Handler())
	return r
}
