package main

import (
	"fmt"
	"os"

	log "github.com/inconshreveable/log15"
	"google.golang.org/grpc"

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

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port), grpc.WithInsecure())
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
		logger.Debug("Starting Server")

		if err := server.Serve(c, nil, logger); err != nil {
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
