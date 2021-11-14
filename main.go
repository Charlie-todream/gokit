package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/openzipkin/zipkin-go"
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
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
)

func main() {

	var (
		consulHost  = flag.String("consul_host", "localhost", "consul ip address")
		consulPort  = flag.String("consul_port", "8500", "consul port")
		serviceHost = flag.String("service_host", "localhost", "service ip address")
		servicePort = flag.String("service_port", "9000", "service port")
		zipkinURL   = flag.String("zipkin.url", "http://192.168.192.146:9411/api/v2/spans", "Zipkin server url")
	)
	flag.String("hello", "asan", "姓名")
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
	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = *serviceHost + ":" + *servicePort
			serviceName   = "arithmetic-service"
			useNoopTracer = (*zipkinURL == "")
			reporter      = zipkinhttp.NewReporter(*zipkinURL)
		)
		defer reporter.Close()
		zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
		zipkinTracer, err = zipkin.NewTracer(
			reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer),
		)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		if !useNoopTracer {
			logger.Log("tracer", "Zipkin", "type", "Native", "URL", *zipkinURL)
		}
	}

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
	//添加追踪，设置span的名称为health-endpoint
	healthEndpoint = kitzipkin.TraceEndpoint(zipkinTracer, "health-endpoint")(healthEndpoint)

	//把算术运算Endpoint和健康检查Endpoint封装至ArithmeticEndpoints
	//身份认证Endpoint
	authEndpoint := endpoints.MakeAuthEndpoint(svc)
	authEndpoint = services.NewTokenBucketLimitterWithBuildIn(ratebucket)(authEndpoint)
	authEndpoint = kitzipkin.TraceEndpoint(zipkinTracer, "login-endpoint")(authEndpoint)

	endpts := endpoints.ArithmeticEndpoints{
		ArithmeticEndpoint:  endpoint,
		HealthCheckEndpoint: healthEndpoint,
		AuthEndpoint:        authEndpoint,
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

	//服务退出，取消注册
	error := <-errChan
	registar.Deregister()
	fmt.Println(error)
	fmt.Println(<-errChan)

}
