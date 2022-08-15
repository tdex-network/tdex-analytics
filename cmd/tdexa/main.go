package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	tdexav1 "tdex-analytics/api-spec/protobuf/gen/tdexa/v1"
)

var (
	maxMsgRecvSize = grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 200)
	tdexaDataDir   = btcutil.AppDataDir("tdexa", false)
	statePath      = path.Join(tdexaDataDir, "state.json")
)

func main() {
	app := cli.NewApp()
	app.Version = "0.0.1" //TODO use goreleaser for setting version
	app.Name = "Tdex-Analytics CLI"
	app.Usage = "Command line interface for Tdex-Analytics daemon"
	app.Commands = append(
		app.Commands,
		configCmd,
		listBalancesCmd,
		listPricesCmd,
		marketsCmd,
		healthCheckCmd,
	)

	err := app.Run(os.Args)
	if err != nil {
		fatal(err)
	}
}

type invalidUsageError struct {
	ctx     *cli.Context
	command string
}

func (e *invalidUsageError) Error() string {
	return fmt.Sprintf("invalid usage of command %s", e.command)
}

func fatal(err error) {
	var e *invalidUsageError
	if errors.As(err, &e) {
		_ = cli.ShowCommandHelp(e.ctx, e.command)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "[tower] %v\n", err)
	}
	os.Exit(1)
}

func getTlsOpts(state map[string]string) ([]grpc.DialOption, error) {
	tlsMod, err := strconv.Atoi(state["tls_mod"])
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{grpc.WithDefaultCallOptions(maxMsgRecvSize)}

	switch tlsMod {
	case tlsNoVerify:
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))

		return opts, nil
	case tlsVerifyNoCA:
		config := &tls.Config{
			InsecureSkipVerify: false,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))

		return opts, nil
	case tlsWithCertFile:
		certPath, ok := state["tls_cert_path"]
		if !ok {
			return nil, errors.New("tls_cert_path not provided, please provide ca.cert file path")
		}

		b, err := ioutil.ReadFile(certPath)
		if err != nil {
			return nil, err
		}
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(b) {
			return nil, errors.New("credentials: failed to append certificates")
		}
		config := &tls.Config{
			InsecureSkipVerify: false,
			RootCAs:            cp,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))

		return opts, nil
	case tlsVerifyWithCA:
		certPath, ok := state["tls_cert_path"]
		if !ok {
			return nil, errors.New("tls_cert_path not provided, please provide pem file")
		}

		creds, err := credentials.NewClientTLSFromFile(certPath, "")
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))

		return opts, nil
	case insecureNoTLS:
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

		return opts, nil
	}

	return nil, err
}

func setState(data map[string]string) error {

	if _, err := os.Stat(tdexaDataDir); os.IsNotExist(err) {
		os.Mkdir(tdexaDataDir, os.ModeDir|0755)
	}

	file, err := os.OpenFile(statePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	currentData, err := getState()
	if err != nil {
		fmt.Println(err)
		return err
	}

	mergedData := merge(currentData, data)

	jsonString, err := json.Marshal(mergedData)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(statePath, jsonString, 0755)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}

func merge(maps ...map[string]string) map[string]string {
	merge := make(map[string]string, 0)
	for _, m := range maps {
		for k, v := range m {
			merge[k] = v
		}
	}
	return merge
}

/*
Modified from https://github.com/lightninglabs/pool/blob/master/cmd/pool/main.go
Original Copyright 2017 Oliver Gugger. All Rights Reserved.
*/
func printRespJSON(resp interface{}) {
	jsonMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true,
		OrigName:     true,
		Indent:       "\t", // Matches indentation of printJSON.
	}

	jsonStr, err := jsonMarshaler.MarshalToString(resp.(proto.Message))
	if err != nil {
		fmt.Println("unable to decode response: ", err)
		return
	}

	fmt.Println(jsonStr)
}

func getState() (map[string]string, error) {
	data := map[string]string{}

	file, err := ioutil.ReadFile(statePath)
	if err != nil {
		return nil, errors.New("get config state error: try 'config init'")
	}
	json.Unmarshal(file, &data)

	return data, nil
}

func getAnalyticsClient() (tdexav1.AnalyticsClient, func(), error) {
	creds, err := getCreds()
	if err != nil {
		return nil, nil, err
	}

	conn, err := getClientConn(creds)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }

	return tdexav1.NewAnalyticsClient(conn), cleanup, nil
}

func getHealthClient() (grpchealth.HealthClient, func(), error) {
	creds, err := getCreds()
	if err != nil {
		return nil, nil, err
	}

	conn, err := getClientConn(creds)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = conn.Close() }

	return grpchealth.NewHealthClient(conn), cleanup, nil
}

func getCreds() ([]grpc.DialOption, error) {
	state, err := getState()
	if err != nil {
		return nil, err
	}

	opts, err := getTlsOpts(state)
	if err != nil {
		return nil, err
	}

	return opts, nil
}

func getClientConn(credentials []grpc.DialOption) (*grpc.ClientConn, error) {
	state, err := getState()
	if err != nil {
		return nil, err
	}
	address, ok := state["rpcserver"]
	if !ok {
		return nil, errors.New("set rpcserver with `config set rpcserver`")
	}

	conn, err := grpc.Dial(address, credentials...)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to RPC server: %v",
			err)
	}

	return conn, nil
}
