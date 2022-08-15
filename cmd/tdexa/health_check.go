package main

import (
	"context"

	grpchealth "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/urfave/cli/v2"
)

var healthCheckCmd = &cli.Command{
	Name:   "health",
	Usage:  "health check",
	Action: healthCheck,
}

func healthCheck(ctx *cli.Context) error {
	req := &grpchealth.HealthCheckRequest{
		Service: "",
	}

	client, cleanup, err := getHealthClient()
	if err != nil {
		return err
	}
	defer cleanup()

	result, err := client.Check(
		context.Background(),
		req,
	)
	if err != nil {
		return err
	}

	printRespJSON(result)

	return nil
}
