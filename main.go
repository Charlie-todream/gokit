package main

import (
	"context"
	"fmt"
	"learn/endpoints"
	"learn/services"
	"learn/transports"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"golang.org/x/time/rate"
)

func main() {
	ctx := context.Background()
	errChan := make(chan error)
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	var svc services.Service
	svc = services.ArithmeticService{}
	// 日志
	svc = services.LoggingMiddleware(logger)(svc)
	endpoint := endpoints.MakeArithmeticEndpoint(svc)
	// 限流juju 每秒内容量为3
	//ratebucket := ratelimit.NewBucket(time.Second*3, 3)
	//endpoint = services.NewTokenBucketLimitterWithJuju(ratebucket)(endpoint)
	// 使用内置的 golang.org/x/time/rate 限流中间件
	ratebucket := rate.NewLimiter(rate.Every(time.Second*4), 3)
	endpoint = services.NewTokenBucketLimitterWithBuildIn(ratebucket)(endpoint)
	r := transports.MakeHttpHandler(ctx, endpoint, logger)

	go func() {
		fmt.Println("Http Server start at port:9000")
		handler := r
		errChan <- http.ListenAndServe(":9000", handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	fmt.Println(<-errChan)

}
