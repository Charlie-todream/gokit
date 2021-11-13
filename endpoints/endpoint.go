package endpoints

import (
	"context"
	"errors"
	"learn/services"
	"strings"

	"github.com/go-kit/kit/endpoint"
)

var ErrInvalidRequestType = errors.New("错误！")

type ArithmeticRequest struct {
	RequestType string `json:"request_type"`
	A           int    `json:"a"`
	B           int    `json:"b"`
}

type ArithmeticResponse struct {
	Result int   `json:"Result"`
	Error  error `json:"error"`
}

type ArithmeticEndpoints struct {
	ArithmeticEndpoint  endpoint.Endpoint
	HealthCheckEndpoint endpoint.Endpoint
}

// HealthRequest 健康检查请求结构
type HealthRequest struct{}

// HealthResponse 健康检查响应结构
type HealthResponse struct {
	Status bool `json:"status"`
}

// MakeHealthCheckEndpoint 创建健康检查Endpoint
func MakeHealthCheckEndpoint(svc services.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return HealthResponse{status}, nil
	}
}

// MakeArithmeticEndpoint make endpoint
func MakeArithmeticEndpoint(svc services.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(ArithmeticRequest)

		var (
			res, a, b int
			calError  error
		)

		a = req.A
		b = req.B

		if strings.EqualFold(req.RequestType, "Add") {
			res = svc.Add(a, b)
		} else if strings.EqualFold(req.RequestType, "Substract") {
			res = svc.Subtract(a, b)
		} else if strings.EqualFold(req.RequestType, "Multiply") {
			res = svc.Multiply(a, b)
		} else if strings.EqualFold(req.RequestType, "Divide") {
			res, calError = svc.Divide(a, b)
		} else {
			return nil, ErrInvalidRequestType
		}

		return ArithmeticResponse{Result: res, Error: calError}, nil
	}
}
