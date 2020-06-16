package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/atotto/clipboard"
	log "github.com/inconshreveable/log15"
	"github.com/pocke/go-iprange"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/jrc2139/vimonade/lemon"
	"github.com/jrc2139/vimonade/models"
	"github.com/jrc2139/vimonade/pkg/protocol/grpc"
	v1 "github.com/jrc2139/vimonade/pkg/service/v1"
)

const MSGPACK = "application/x-msgpack"

var logger log.Logger
var lineEnding string
var ra *iprange.Range
var port int
var path = "./files"

func handleCopy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Copy only support post", 404)
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer r.Body.Close()

	var m models.Message

	if err := msgpack.Unmarshal(b, &m); err != nil {
		logger.Error("error unmarshalling msgpack.", "err", err.Error())
		http.Error(w, "error unmarshalling msgpack.", http.StatusInternalServerError)

		return
	}

	text := lemon.ConvertLineEnding(m.Text, lineEnding)
	logger.Debug("Copy:", "text", text)

	if err := clipboard.WriteAll(text); err != nil {
		logger.Error("error writing to clipboard.", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}
}

func handlePaste(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Paste only support get", 404)
		return
	}

	t, err := clipboard.ReadAll()
	if err != nil {
		logger.Error("error reading from clipboard.", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("content-type", MSGPACK)

	w.WriteHeader(http.StatusOK)

	b, err := msgpack.Marshal(&models.Message{Text: t})
	if err != nil {
		logger.Error("error marshalling msgpack.", "err", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	if _, err := w.Write(b); err != nil {
		logger.Error("error writing resp", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	logger.Debug("Paste: ", "text", t)
}

func translateLoopbackIP(uri string, remoteIP string) string {
	parsed, err := url.Parse(uri)
	if err != nil {
		return uri
	}

	host, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		return err.Error()
	}

	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return uri
	}

	if len(port) == 0 {
		parsed.Host = remoteIP
	} else {
		parsed.Host = fmt.Sprintf("%s:%s", remoteIP, port)
	}

	return parsed.String()
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "Not support method.", 404)
			return
		}

		remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "RemoteAddr error.", 500)
			return
		}
		if !ra.IncludeStr(remoteIP) {
			http.Error(w, "Not allow ip.", 503)
			logger.Info("not in allow ip. from: ", remoteIP)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Serve(c *lemon.CLI, _logger log.Logger) error {
	logger = _logger
	lineEnding = c.LineEnding
	port = c.Port

	var err error
	ra, err = iprange.New(c.Allow)
	if err != nil {
		logger.Error("allowIp error")
		return err
	}
	// flag.Parse()
	// lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	// if err != nil {
	// log.Fatalf("failed to listen: %v", err)
	// }
	// grpcServer := grpc.NewServer()
	// pb.RegisterRouteGuideServer(grpcServer, &routeGuideServer{})
	// ... // determine whether to use TLS
	// grpcServer.Serve(lis)

	// http.Handle("/copy", middleware(http.HandlerFunc(handleCopy)))
	// http.Handle("/paste", middleware(http.HandlerFunc(handlePaste)))
	// err = http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil)
	// if err != nil {
	// return err
	// }

	if err := grpc.RunServer(context.Background(), v1.NewMessageServerService(c.LineEnding), fmt.Sprintf("%s:%d", c.Host, c.Port)); err != nil {
		return err
	}

	return nil
}
