package services

import (
	"time"

	"github.com/go-kit/kit/log"
)

type loggingMiddleware struct {
	Service
	logger log.Logger
}

func LoggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next Service) Service {
		return loggingMiddleware{next, logger}
	}
}

func (mv loggingMiddleware) Add(a, b int) (ret int) {
	defer func(begin time.Time) {
		mv.logger.Log(
			"a", a,
			"b", b,
			"result", ret,
			"took", time.Since(begin),
		)
	}(time.Now())
	ret = mv.Service.Add(a, b)
	return ret
}
