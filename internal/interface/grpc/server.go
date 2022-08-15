package tdexagrpc

import (
	"context"
	"net/http"
	"strings"
	tdexav1 "tdex-analytics/api-spec/protobuf/gen/tdexa/v1"
	"tdex-analytics/internal/core/application"
	grpchandler "tdex-analytics/internal/interface/grpc/handler"
	"tdex-analytics/internal/interface/grpc/interceptor"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	shutdownTimeout = 5 * time.Second
)

type Server interface {
	Start(ctx context.Context, stop context.CancelFunc) <-chan error
}

type server struct {
	serverPort       string
	marketBalanceSvc application.MarketBalanceService
	marketPriceSvc   application.MarketPriceService
	marketsLoaderSvc application.MarketsLoaderService
	marketSvc        application.MarketService
	opts             serverOptions
}

func NewServer(
	serverPort string,
	marketBalanceSvc application.MarketBalanceService,
	marketPriceSvc application.MarketPriceService,
	marketsLoaderSvc application.MarketsLoaderService,
	marketSvc application.MarketService,
	opts ...ServerOption,
) (Server, error) {
	if err := marketsLoaderSvc.StartFetchingMarketsJob(); err != nil {
		return nil, err
	}

	if err := marketPriceSvc.StartFetchingPricesJob(); err != nil {
		return nil, err
	}

	if err := marketBalanceSvc.StartFetchingBalancesJob(); err != nil {
		return nil, err
	}

	defaultOpts := defaultServerOptions(serverPort)
	for _, o := range opts {
		if err := o.apply(&defaultOpts); err != nil {
			return nil, err
		}
	}

	return &server{
		serverPort:       serverPort,
		marketBalanceSvc: marketBalanceSvc,
		marketPriceSvc:   marketPriceSvc,
		marketsLoaderSvc: marketsLoaderSvc,
		marketSvc:        marketSvc,
		opts:             defaultOpts,
	}, nil
}

func (s *server) Start(ctx context.Context, stop context.CancelFunc) <-chan error {
	errC := make(chan error, 1)
	// tdexa grpc server
	tdexaGrpcServer, err := s.tdexaGrpcServer()
	if err != nil {
		errC <- err
	}

	// grpc web server
	grpcWebServer := grpcweb.WrapServer(
		tdexaGrpcServer,
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
	)

	// grpc gateway
	tdexaGrpcGateway, err := s.tdexaGrpcGateway(ctx)
	if err != nil {
		errC <- err
	}

	httpHandler := router(tdexaGrpcServer, grpcWebServer, tdexaGrpcGateway)
	if !s.opts.tlsSecure {
		httpHandler = h2c.NewHandler(httpHandler, &http2.Server{})
	}

	httpServer := &http.Server{
		Addr:      s.opts.serverAddress,
		Handler:   httpHandler,
		TLSConfig: s.opts.tlsConfig,
	}

	// listen to signal context in order to gracefully shutdown
	go func() {
		<-ctx.Done()

		log.Info("shutdown signal received")

		ctxTimeout, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

		defer func() {
			stop()
			cancel()
			close(errC)
		}()

		httpServer.SetKeepAlivesEnabled(false)
		if err := httpServer.Shutdown(ctxTimeout); err != nil {
			errC <- err
		}

		log.Info("tdex-analytics daemon server graceful shutdown completed")
	}()

	// start http server
	go func() {
		log.Infof("tdex-analytics daemon listening and serving at: %v", s.opts.serverAddress)

		// ListenAndServe always returns a non-nil error. After Shutdown or Close, the returned error is ErrServerClosed.
		if s.opts.tlsSecure {
			if err := httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				errC <- err
			}
		} else {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errC <- err
			}
		}
	}()

	return errC
}

func (s *server) tdexaGrpcServer() (*grpc.Server, error) {
	analyticsHandler := grpchandler.NewAnalyticsHandler(
		s.marketBalanceSvc,
		s.marketPriceSvc,
		s.marketSvc,
	)

	healthHandler := grpchandler.NewHealthHandler()

	chainInterceptorSvc, err := interceptor.NewService()
	if err != nil {
		return nil, err
	}

	opts := chainInterceptorSvc.CreateServerOpts()
	if s.opts.tlsSecure {
		if err != nil {
			return nil, err
		}
		opts = append(chainInterceptorSvc.CreateServerOpts(), s.opts.grpcServerCredsOpts...)
	}

	tdexaGrpcServer := grpc.NewServer(opts...)
	tdexav1.RegisterAnalyticsServer(tdexaGrpcServer, analyticsHandler)
	grpchealth.RegisterHealthServer(tdexaGrpcServer, healthHandler)

	return tdexaGrpcServer, nil
}

func (s *server) tdexaGrpcGateway(ctx context.Context) (http.Handler, error) {
	conn, err := grpc.DialContext(context.Background(), s.opts.clientDialUrl, s.opts.clientDialCredsOpts...)
	if err != nil {
		return nil, err
	}

	grpcGatewayMux := runtime.NewServeMux(runtime.WithHealthzEndpoint(grpchealth.NewHealthClient(conn)))
	if err := tdexav1.RegisterAnalyticsHandler(ctx, grpcGatewayMux, conn); err != nil {
		return nil, err
	}

	grpcGatewayHandler := http.Handler(grpcGatewayMux)

	return grpcGatewayHandler, nil
}

func router(
	grpcServer *grpc.Server,
	grpcWebServer *grpcweb.WrappedGrpcServer,
	grpcGateway http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isHttpRequest(r) {
			grpcGateway.ServeHTTP(w, r)
			return
		}
		if isGrpcWebRequest(r) {
			grpcWebServer.ServeHTTP(w, r)
			return
		}
		grpcServer.ServeHTTP(w, r)
	})
}

func isHttpRequest(req *http.Request) bool {
	return strings.ToLower(req.Method) == "get" ||
		strings.Contains(req.Header.Get("Content-Type"), "application/json")
}

func isGrpcWebRequest(req *http.Request) bool {
	return isValidGrpcWebOptionRequest(req) || isValidGrpcWebRequest(req)
}

func isValidGrpcWebRequest(req *http.Request) bool {
	return req.Method == http.MethodPost && isValidGrpcContentTypeHeader(req.Header.Get("content-type"))
}

func isValidGrpcContentTypeHeader(contentType string) bool {
	return strings.HasPrefix(contentType, "application/grpc-web-text") ||
		strings.HasPrefix(contentType, "application/grpc-web")
}

func isValidGrpcWebOptionRequest(req *http.Request) bool {
	accessControlHeader := req.Header.Get("Access-Control-Request-Headers")
	return req.Method == http.MethodOptions &&
		strings.Contains(accessControlHeader, "x-grpc-web") &&
		strings.Contains(accessControlHeader, "content-type")
}
