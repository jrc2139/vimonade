package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	rice "github.com/GeertJohan/go.rice"
	log "github.com/inconshreveable/log15"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/jrc2139/vimonade/client"
	"github.com/jrc2139/vimonade/lemon"
	"github.com/jrc2139/vimonade/server"
)

var logLevelMap = map[int]log.Lvl{
	0: log.LvlDebug,
	1: log.LvlInfo,
	2: log.LvlWarn,
	3: log.LvlError,
	4: log.LvlCrit,
}

func main() {
	cli := &lemon.CLI{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
	os.Exit(Do(cli, os.Args))
}

func Do(c *lemon.CLI, args []string) int {
	logger := log.New()
	logger.SetHandler(log.LvlFilterHandler(log.LvlError, log.StdoutHandler))

	if err := c.FlagParse(args, false); err != nil {
		fmt.Fprintln(c.Err, err.Error())
		return lemon.FlagParseError
	}

	logLevel := logLevelMap[c.LogLevel]
	logger.SetHandler(log.LvlFilterHandler(logLevel, log.StdoutHandler))

	if c.Help {
		fmt.Fprint(c.Err, lemon.Usage)
		return lemon.Help
	}

	conf := rice.Config{
		LocateOrder: []rice.LocateMethod{rice.LocateEmbedded},
	}

	// find a rice.Box
	certBox, err := conf.FindBox("../../certs")
	if err != nil {
		panic(err)
	}

	serverPemBytes, err := certBox.Bytes("service.pem")
	if err != nil {
		panic(err)
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(serverPemBytes) {
		logger.Crit("Failed to append to cert", err, nil)
	}

	clientCreds := credentials.NewTLS(&tls.Config{ServerName: "", RootCAs: cp})

	switch c.Type {
	case lemon.COPY:
		logger.Debug("Copying text")
		return client.Copy(c, logger, grpc.WithTransportCredentials(clientCreds), grpc.WithBlock())

	case lemon.PASTE:
		logger.Debug("Pasting text")
		return client.Paste(c, logger, grpc.WithTransportCredentials(clientCreds), grpc.WithInsecure(), grpc.WithBlock())

	case lemon.SERVER:
		serverKeyBytes, err := certBox.Bytes("service.key")
		if err != nil {
			logger.Crit("Failed to create server key pair", nil, err)
			return lemon.RPCError
		}

		cert, err := tls.X509KeyPair(serverPemBytes, serverKeyBytes)
		if err != nil {
			logger.Crit("Failed to create server key pair", nil, err)
			return lemon.RPCError
		}

		serverCreds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})

		logger.Debug("Starting Server")

		if err := server.Serve(c, serverCreds, logger); err != nil {
			logger.Crit("Server error", err, nil)
			return lemon.RPCError
		}
	default:
		logger.Crit("Vimonade error", err, nil)
		return lemon.RPCError
	}

	return lemon.Success
}
