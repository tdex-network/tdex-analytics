package main

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"strconv"
)

const (
	//tlsNoVerify mod, client does not authenticate the Server
	tlsNoVerify = iota
	// tlsVerifyNoCA mod, server cert signed by 3rd party CA
	tlsVerifyNoCA
	// tlsWithCertFile mod, server cert(pem file) is trusted no need to verify it
	tlsWithCertFile
	// tlsVerifyWithCA mod, client uses Certification Authority (CA) cert file to verify server
	tlsVerifyWithCA
	// insecureNoTLS mod, server side without TLS, client uses unencrypted transport
	insecureNoTLS
)

var configCmd = &cli.Command{
	Name:   "config",
	Usage:  "Configures gate cli",
	Action: configure,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "rpcserver",
			Usage: "gated daemon address host:port",
			Value: "localhost:9000",
		},
		&cli.StringFlag{
			Name:  "tls_cert_path",
			Usage: "the path of the server TLS certificate file to use",
		},
		&cli.IntFlag{
			Name: "tls_mod",
			Usage: "tls security modes:\n" +
				"	0 -> client does not authenticate the Server(recommended for testing when server uses TLS)\n" +
				"	1 -> server cert signed by 3rd party CA\n" +
				"	2 -> server cert(pem file) is trusted no need to verify it(recommended for testing)\n" +
				"	3 -> client uses Certification Authority (CA) cert file to verify server\n" +
				"	4 -> server side without TLS, client uses unencrypted transport(recommended for testing when server doesnt use TLS)",
			Value: 4,
		},
	},
	Subcommands: []*cli.Command{
		{
			Name:   "set",
			Usage:  "set individual <key> <value> in the local state",
			Action: configSetAction,
		},
		{
			Name:   "print",
			Usage:  "Print local configuration of the tower CLI",
			Action: configList,
		},
	},
}

func configure(ctx *cli.Context) error {
	configState := make(map[string]string)

	configState["rpcserver"] = ctx.String("rpcserver")

	tlsMod := ctx.Int("tls_mod")
	configState["tls_mod"] = strconv.Itoa(tlsMod)

	return setState(configState)
}

func configSetAction(c *cli.Context) error {

	if c.NArg() < 2 {
		return errors.New("key and value are missing")
	}

	key := c.Args().Get(0)
	value := c.Args().Get(1)

	err := setState(map[string]string{key: value})
	if err != nil {
		return err
	}

	fmt.Printf("%s %s has been set\n", key, value)

	return nil
}

func configList(ctx *cli.Context) error {

	state, err := getState()
	if err != nil {
		return err
	}

	for key, value := range state {
		fmt.Println(key + ": " + value)
	}

	return nil
}
