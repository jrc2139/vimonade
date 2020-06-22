package server

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"

	log "github.com/inconshreveable/log15"

	"github.com/jrc2139/vimonade/lemon"
	"github.com/jrc2139/vimonade/protocol/grpc"
	"github.com/jrc2139/vimonade/service"
)

func Serve(c *lemon.CLI, creds credentials.TransportCredentials, logger log.Logger) int {
	// create vimonade dir if !exist
	var vimonadeDir string

	if c.VimonadeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Crit("Cannot find $HOME error", err, nil)
		}

		vimonadeDir = home + "/.vimonade/files"
	} else {
		vimonadeDir = c.VimonadeDir
	}

	logger.Debug("current vimonade dir: " + vimonadeDir)

	if _, err := os.Stat(vimonadeDir); os.IsNotExist(err) {
		err = os.MkdirAll(vimonadeDir, 0755)
		if err != nil {
			logger.Crit("Creating vimonade dir error", err, nil)
			return lemon.RPCError
		}
	}

	// Server
	store := service.NewDiskFileStore(vimonadeDir)

	if err := grpc.RunServer(context.Background(), service.NewVimonadeServerService(store, c.LineEnding, logger), logger, creds, c.Allow, fmt.Sprintf("%s:%d", c.Host, c.Port)); err != nil {
		logger.Crit("Server error", err, nil)
		fmt.Fprintln(c.Err, err.Error())

		return lemon.RPCError
	}

	return lemon.RPCError
}
