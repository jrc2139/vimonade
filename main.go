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

	// _ "github.com/jrc2139/vimonade/certs"

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
		writeError(c, err)
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
	certBox, err := conf.FindBox("certs")
	if err != nil {
		panic(err)
	}

	serverPemBytes, err := certBox.Bytes("service.pem")
	if err != nil {
		panic(err)
	}

	// serverPem, err := parcello.Open("service.pem")
	// if err != nil {
	// logger.Crit("Failed to open service.pem", err, nil)
	// }

	// defer serverPem.Close()

	// var serverPemBytes []byte

	// if n, err := serverPem.Read(serverPemBytes); err != nil {
	// n, err := serverPem.Read(serverPemBytes)
	// if err != nil {
	// logger.Crit("Failed to read service.pem", err, nil)
	// }
	// fmt.Println(n)

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(serverPemBytes) {
		logger.Crit("Failed to append to cert", err, nil)
	}

	clientCreds := credentials.NewTLS(&tls.Config{ServerName: "", RootCAs: cp})
	// credentials.NewClientTLSFromFile()

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), grpc.WithTransportCredentials(clientCreds))
	if err != nil {
		logger.Debug(err.Error())
	}
	defer conn.Close()

	lc := client.New(c, conn, logger)

	switch c.Type {
	case lemon.COPY:
		logger.Debug("Copying text")

		if err := lc.Copy(c.DataSource); err != nil {
			logger.Crit("Failed to Copy", err, nil)
		}
	case lemon.PASTE:
		logger.Debug("Pasting text")

		var text string

		text, err := lc.Paste()
		if err != nil {
			logger.Crit("Failed to Paste", err, nil)
		}

		if _, err := c.Out.Write([]byte(text)); err != nil {
			logger.Crit("Failed to output Paste to stdin", err, nil)
		}
	case lemon.SERVER:
		// serverKey, err := parcello.Open("service.key")
		// if err != nil {
		// logger.Crit("Failed to load key file", err, nil)
		// return lemon.RPCError
		// }
		//
		// var serverKeyBytes []byte
		//
		// if _, err := serverKey.Read(serverKeyBytes); err != nil {
		// logger.Crit("Failed to read service.key", err, nil)
		// return lemon.RPCError
		// }
		//
		// defer serverKey.Close()

		serverKeyBytes, err := certBox.Bytes("service.key")
		if err != nil {
			panic(err)
		}

		cert, err := tls.X509KeyPair(serverPemBytes, serverKeyBytes)
		if err != nil {
			logger.Crit("Failed to create server key pair", nil, err)
			return lemon.RPCError
		}

		// creds, err := credentials.NewServerTLSFromFile(p.Name(), key)
		serverCreds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})
		// if err != nil {
		// logger.Crit("Failed to setup TLS", err, nil)
		// }

		logger.Debug("Starting Server")

		if err := server.Serve(c, serverCreds, logger); err != nil {
			logger.Crit("Server error", err, nil)
			return lemon.RPCError
		}
	default:
		panic("Unreachable code")
	}

	if err != nil {
		writeError(c, err)
		logger.Crit("Vimonade error", err, nil)
		return lemon.RPCError
	}

	return lemon.Success
}

func writeError(c *lemon.CLI, err error) {
	fmt.Fprintln(c.Err, err.Error())
}
