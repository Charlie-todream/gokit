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

func MakeHttpHandler(ctx context.Context, endpoints endpoints.ArithmeticEndpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(kithttp.DefaultErrorEncoder),
	}

	r.Methods("POST").Path("/calculate/{type}/{a}/{b}").Handler(kithttp.NewServer(
		endpoints.ArithmeticEndpoint,
		decodeArithmeticRequest,
		encodeArithmeticResponse,
		options...,
	))
	r.Path("/metrics").Handler(promhttp.Handler())

	r.Methods("GET").Path("/health").Handler(kithttp.NewServer(
		endpoints.HealthCheckEndpoint,
		decodeArithmeticRequest,
		encodeArithmeticResponse,
		options...,
	))

	r.Methods("POST").Path("/login").Handler(kithttp.NewServer(
		endpoints.AuthEndpoint,
		decodeLoginRequest,
		encodeLoginResponse,
		options...,
	))

	return r
}

func decodeLoginRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var loginRequest endpoints.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		return nil, err
	}
	return loginRequest, nil
}

func encodeLoginResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
