package client

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/atotto/clipboard"
	log "github.com/inconshreveable/log15"
	"github.com/jrc2139/vimonade/lemon"
	"github.com/jrc2139/vimonade/models"
	"github.com/vmihailenco/msgpack/v5"
)

type client struct {
	host               string
	port               int
	addr               string
	lineEnding         string
	noFallbackMessages bool
	logger             log.Logger
}

func New(c *lemon.CLI, logger log.Logger) *client {
	return &client{
		host:               c.Host,
		port:               c.Port,
		addr:               fmt.Sprintf("http://%s:%d", c.Host, c.Port),
		lineEnding:         c.LineEnding,
		noFallbackMessages: c.NoFallbackMessages,
		logger:             logger,
	}
}

const MSGPACK = "application/x-msgpack"

func (c *client) Copy(text string) error {
	c.logger.Debug("Sending: " + text)
	url := fmt.Sprintf("%s/copy", c.addr)

	b, err := msgpack.Marshal(&models.Message{Text: text})
	if err != nil {
		c.logger.Error("error marshalling msgpack.", "err", err.Error())
		return err
	}

	resp, err := http.Post(url, MSGPACK, bytes.NewBuffer(b))
	if err != nil {
		clipboard.WriteAll(text)
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *client) Paste() (string, error) {
	url := fmt.Sprintf("%s/paste", c.addr)

	r, err := http.Get(url)
	if err != nil {
		c.logger.Error("http error.", "err", err.Error())
		return "", err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.logger.Error("http read body error.", "err", err.Error())
		return "", err
	}

	var m models.Message

	if err := msgpack.Unmarshal(body, &m); err != nil {
		c.logger.Error("error unmarshalling msgpack.", "err", err.Error())
	}

	return lemon.ConvertLineEnding(m.Text, c.lineEnding), nil
}

func fileExists(fname string) bool {
	_, err := os.Stat(fname)
	return err == nil
}

func (c *client) postFile(name string, url string) (*http.Response, error) {
	bodyBuf := bytes.NewBufferString("")
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("uploadFile", name)
	if err != nil {
		c.logger.Error("Writing to buffer", "name", name)
		return nil, err
	}

	file, err := os.Open(name)
	if err != nil {
		c.logger.Error("cant Opening file", "name", name)
		return nil, err
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	return http.Post(url, contentType, bodyBuf)
}

func (c *client) uploadFile(name string) error {
	url := fmt.Sprintf("%s/upload?open=true", c.addr)
	_, err := c.postFile(name, url)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) Open(uri string, transLocalfile bool, transLoopback bool) error {
	if transLocalfile && fileExists(uri) {
		return c.uploadFile(uri)
	}

	url := fmt.Sprintf("%s/open?uri=%s&transLoopback=%s&base64=true", c.addr, base64.URLEncoding.EncodeToString([]byte(uri)), strconv.FormatBool(transLoopback))
	c.logger.Info("Opening: " + uri)

	_, err := http.Get(url)
	if err != nil {
		c.logger.Error("http error.", "err", err.Error())
		return err
	}
	return nil
}
