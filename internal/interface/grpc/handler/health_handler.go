package grpchandler

import (
	"context"

	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
)

type healthHandler struct{}

func NewHealthHandler() grpchealth.HealthServer {
	return &healthHandler{}
}

func (h *healthHandler) Check(
	ctx context.Context,
	req *grpchealth.HealthCheckRequest,
) (*grpchealth.HealthCheckResponse, error) {

	return &grpchealth.HealthCheckResponse{
		Status: grpchealth.HealthCheckResponse_SERVING,
	}, nil
}

func (h *healthHandler) Watch(
	req *grpchealth.HealthCheckRequest,
	w grpchealth.Health_WatchServer,
) error {

	return nil
}
