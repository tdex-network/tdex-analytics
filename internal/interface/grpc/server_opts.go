package tdexagrpc

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func defaultServerOptions(serverPort string) serverOptions {
	return serverOptions{
		serverAddress: fmt.Sprintf(":%v", serverPort),
		clientDialUrl: fmt.Sprintf("localhost:%v", serverPort),
	}
}

// A ServerOption sets options such as credentials, codec and keepalive parameters, etc.
type ServerOption interface {
	apply(*serverOptions) error
}

type serverOptions struct {
	serverAddress       string
	clientDialUrl       string
	clientDialCredsOpts []grpc.DialOption
	tlsSecure           bool
	tlsConfig           *tls.Config
	grpcServerCredsOpts []grpc.ServerOption
	keyFile             string
	certFile            string
}

// funcServerOption wraps a function that modifies serverOptions into an
// implementation of the ServerOption interface.
type funcServerOption struct {
	f func(*serverOptions) error
}

func (fdo *funcServerOption) apply(do *serverOptions) error {
	return fdo.f(do)
}

func newFuncServerOption(f func(*serverOptions) error) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

func WithInsecureGrpcGateway() ServerOption {
	return newFuncServerOption(func(s *serverOptions) error {
		s.clientDialCredsOpts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

		return nil
	})
}

func WithTls(certFile, keyFile string) ServerOption {
	return newFuncServerOption(func(s *serverOptions) error {
		tlsConfig, err := tlsConfig(certFile, keyFile)
		if err != nil {
			return err
		}

		s.grpcServerCredsOpts = []grpc.ServerOption{grpc.Creds(credentials.NewTLS(tlsConfig))}

		conf := &tls.Config{
			InsecureSkipVerify: true,
		}
		s.clientDialCredsOpts = []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(conf))}

		s.certFile = certFile
		s.keyFile = keyFile
		s.tlsConfig = tlsConfig
		s.tlsSecure = true

		return nil
	})
}

func tlsConfig(certFile, keyFile string) (*tls.Config, error) {
	//TODO add acme
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	const requiredCipher = tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	tlsConfig := &tls.Config{
		CipherSuites: []uint16{requiredCipher},
		NextProtos:   []string{"http/1.1", http2.NextProtoTLS, "h2-14"}, // h2-14 is just for compatibility. will be eventually removed.
		Certificates: []tls.Certificate{certificate},
	}
	tlsConfig.Rand = rand.Reader

	return tlsConfig, nil
}
