package main

import (
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"

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
	if err := c.FlagParse(args, false); err != nil {
		fmt.Fprintln(c.Err, err.Error())
		return lemon.FlagParseError
	}

	logger := logging.InitLogger(c.LogLevel)

	if c.Help {
		fmt.Fprint(c.Err, lemon.Usage)
		return lemon.Help
	}

	opts := []grpc.DialOption{grpc.WithInsecure(),
		grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: 10 * time.Millisecond})}

	switch c.Type {
	case lemon.COPY:
		logger.Debug("Copying text")
		return vc.Copy(c, logger, opts...)

	case lemon.PASTE:
		logger.Debug("Pasting text")
		return vc.Paste(c, logger, opts...)

	case lemon.SEND:
		logger.Debug("Sending file")
		return vc.Send(c, logger, opts...)

	case lemon.SYNC:
		logger.Debug("Syncing file")
		return vc.Sync(c, logger, opts...)

	case lemon.SERVER:
		logger.Debug("Starting Server")
		return vs.Serve(c, nil, logger)

	default:
		logger.Error("vimonade error")
		return lemon.RPCError
	}
}
