package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	rice "github.com/GeertJohan/go.rice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	vc "github.com/jrc2139/vimonade/client"
	"github.com/jrc2139/vimonade/lemon"
	"github.com/jrc2139/vimonade/logging"
	vs "github.com/jrc2139/vimonade/server"
)

func main() {
	cli := &lemon.CLI{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
	os.Exit(Do(cli, os.Args))
}

func Do(c *lemon.CLI, args []string) int {
	// logger := log.New()
	// logger.SetHandler(log.LvlFilterHandler(log.LvlError, log.StdoutHandler))

	if err := c.FlagParse(args, false); err != nil {
		fmt.Fprintln(c.Err, err.Error())
		return lemon.FlagParseError
	}

	logger := logging.InitLogger(c.LogLevel)

	// logLevel := logLevelMap[c.LogLevel]
	// logger.SetHandler(log.LvlFilterHandler(logLevel, log.StdoutHandler))

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
		logger.Error("Failed to find cert folder:" + err.Error())
		return lemon.RPCError
	}

	serverPemBytes, err := certBox.Bytes("service.pem")
	if err != nil {
		logger.Error("Failed to read server pem: " + err.Error())
		return lemon.RPCError
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(serverPemBytes) {
		logger.Error("Failed to append to cert: " + err.Error())
	}

	clientCreds := credentials.NewTLS(&tls.Config{ServerName: "", RootCAs: cp})

	switch c.Type {
	case lemon.COPY:
		logger.Debug("Copying text")
		return vc.Copy(c, logger, grpc.WithTransportCredentials(clientCreds), grpc.WithBlock())

	case lemon.PASTE:
		logger.Debug("Pasting text")
		return vc.Paste(c, logger, grpc.WithTransportCredentials(clientCreds), grpc.WithBlock())

	case lemon.SEND:
		logger.Debug("Sending file")
		return vc.Send(c, logger, grpc.WithTransportCredentials(clientCreds), grpc.WithBlock())

	case lemon.SERVER:
		serverKeyBytes, err := certBox.Bytes("service.key")
		if err != nil {
			logger.Error("Failed to create server key pair: " + err.Error())
			return lemon.RPCError
		}

		cert, err := tls.X509KeyPair(serverPemBytes, serverKeyBytes)
		if err != nil {
			logger.Error("Failed to create server key pair: " + err.Error())
			return lemon.RPCError
		}

		serverCreds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})

		logger.Debug("Starting Server")

		return vs.Serve(c, serverCreds, logger)

	default:
		logger.Error("Vimonade error: " + err.Error())
		return lemon.RPCError
	}
}
