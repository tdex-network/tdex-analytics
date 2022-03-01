package interceptor

import (
	"context"
	"tdex-analytics/pkg/hexerr"

	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

func (i *interceptorChain) unaryErrorHandler(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	resp, err := handler(ctx, req)
	return resp, handleError(err)
}

func (i *interceptorChain) streamErrorHandler(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	err := handler(srv, stream)
	return handleError(err)
}

func handleError(err error) error {
	var result error
	if err != nil {
		switch e := err.(type) {
		case *hexerr.HexagonalError:
			switch e.Code {
			case hexerr.EntityNotFound:
				result = status.Error(codes.NotFound, err.Error())
				logError(e)
			case hexerr.InvalidArguments:
				result = status.Error(codes.InvalidArgument, err.Error())
				logError(e)
			default:
				result = status.Error(codes.Internal, err.Error())
				logError(e)
			}
		default:
			result = status.Error(codes.Internal, err.Error())
			log.Debugln(e.Error())
		}
	}

	return result
}

func logError(err *hexerr.HexagonalError) {
	switch log.GetLevel() {
	case log.DebugLevel:
		log.Debugln(err.Details())
	case log.TraceLevel:
		log.Debugln(err.StackTrace())
	}
}
