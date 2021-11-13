package main

import (
	"context"
	"flag"
	"fmt"
	"learn/endpoints"
	"learn/registers"
	"learn/services"
	"learn/transports"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"golang.org/x/time/rate"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

func main() {

	var (
		consulHost  = flag.String("consul.host", "", "consul ip address")
		consulPort  = flag.String("consul.port", "", "consul port")
		serviceHost = flag.String("service.host", "", "service ip address")
		servicePort = flag.String("service.port", "", "service port")
	)
	flag.Parse()
	ctx := context.Background()
	errChan := make(chan error)
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	fieldKeys := []string{"method"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "raysonxin",
		Subsystem: "arithmetic_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)

	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "raysonxin",
		Subsystem: "arithemetic_service",
		Name:      "request_latency",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

	var svc services.Service
	svc = services.ArithmeticService{}
	svc = services.Metrics(requestCount, requestLatency)(svc)

	// 日志
	svc = services.LoggingMiddleware(logger)(svc)
	endpoint := endpoints.MakeArithmeticEndpoint(svc)
	// 限流juju 每秒内容量为3
	//ratebucket := ratelimit.NewBucket(time.Second*3, 3)
	//endpoint = services.NewTokenBucketLimitterWithJuju(ratebucket)(endpoint)
	// 使用内置的 golang.org/x/time/rate 限流中间件
	ratebucket := rate.NewLimiter(rate.Every(time.Second*4), 3)
	endpoint = services.NewTokenBucketLimitterWithBuildIn(ratebucket)(endpoint)
	// 健康检查
	//创建健康检查的Endpoint，未增加限流
	healthEndpoint := endpoints.MakeHealthCheckEndpoint(svc)

	//把算术运算Endpoint和健康检查Endpoint封装至ArithmeticEndpoints
	endpts := endpoints.ArithmeticEndpoints{
		ArithmeticEndpoint:  endpoint,
		HealthCheckEndpoint: healthEndpoint,
	}

	//创建http.Handler
	r := transports.MakeHttpHandler(ctx, endpts, logger)
	// 服务注册
	registar := registers.Register(*consulHost, *consulPort, *serviceHost, *servicePort, logger)
	go func() {
		fmt.Println("Http Server start at port:9000")
		handler := r
		errChan <- http.ListenAndServe(":9000", handler)
	}()

	go func() {
		fmt.Println("Http Server start at port:" + *servicePort)
		//启动前执行注册
		registar.Register()
		handler := r
		errChan <- http.ListenAndServe(":"+*servicePort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	fmt.Println(<-errChan)

}
