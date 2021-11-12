package services

type Service interface {
	Add(a, b int) int
	Subtract(a, b int) int
	Multiply(a, b int) int

	Divide(a, b int) (int, error)
}

type ServiceMiddleware func(Service) Service
