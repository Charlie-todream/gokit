package services

import (
	"errors"
	"time"
)

type Service interface {
	Add(a, b int) int
	Subtract(a, b int) int
	Multiply(a, b int) int

	Divide(a, b int) (int, error)
	HealthCheck() bool
}

type ArithmeticService struct {
}

func (s ArithmeticService) Add(a, b int) int {
	return a + b
}

func (s ArithmeticService) Subtract(a, b int) int {
	return a + b
}
func (s ArithmeticService) Multiply(a, b int) int {
	return a * b
}

func (s ArithmeticService) Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("the divided can not be zero!")
	}
	return a / b, nil
}

// 用于检测服务的健康状态
func (s ArithmeticService) HealthCheck() bool {
	return true
}

type ServiceMiddleware func(Service) Service

// loggingMiddleware实现HealthCheck
func (mw loggingMiddleware) HealthCheck() (result bool) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "HealthChcek",
			"result", result,
			"took", time.Since(begin),
		)
	}(time.Now())
	result = mw.Service.HealthCheck()
	return
}

// metricMiddleware实现HealthCheck
func (mw metricMiddleware) HealthCheck() (result bool) {

	defer func(begin time.Time) {
		lvs := []string{"method", "HealthCheck"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	result = mw.Service.HealthCheck()
	return
}
