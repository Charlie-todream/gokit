# gokit
一个简单的 go kit 计算程序
Go-kit 
website : https://gokit.io/

GO-Kit 三层架构
Transport
主要负责与http, grpc, thrift等相关的逻辑

Endpoint
定义Request和Response格式，以及各种中间件

Service
业务类，接口

go mod tidy
go build 
