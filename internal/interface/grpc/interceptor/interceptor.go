package interceptor

import (
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
)

type Service interface {
	CreateServerOpts() []grpc.ServerOption
}

type interceptorChain struct {
}

func NewService() (Service, error) {
	return &interceptorChain{}, nil
}

func (i *interceptorChain) CreateServerOpts() []grpc.ServerOption {
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	//logger
	unaryInterceptors = append(
		unaryInterceptors, i.unaryLogger,
	)
	streamInterceptors = append(
		streamInterceptors, i.streamLogger,
	)

	//error handler
	unaryInterceptors = append(
		unaryInterceptors, i.unaryErrorHandler,
	)
	streamInterceptors = append(
		streamInterceptors, i.streamErrorHandler,
	)

	//panic handler
	unaryInterceptors = append(
		unaryInterceptors, grpc_recovery.UnaryServerInterceptor(),
	)
	streamInterceptors = append(
		streamInterceptors, grpc_recovery.StreamServerInterceptor(),
	)

	chainedUnary := middleware.WithUnaryServerChain(
		unaryInterceptors...,
	)
	chainedStream := middleware.WithStreamServerChain(
		streamInterceptors...,
	)

	serverOpts := []grpc.ServerOption{chainedUnary, chainedStream}

	return serverOpts
}
