package client

func (c *client) Open(uri string, transLocalfile, transLoopback bool) error {
	var finished <-chan struct{}
	c.logger.Debug("Opening " + uri)
	// if transLocalfile && fileExists(uri) {
	// var err error
	// uri, finished, err = serveFile(uri)
	// if err != nil {
	// return err
	// }
	// }

	// err := c.withRPCClient(func(rc *rpc.Client) error {
	// p := &param.OpenParam{
	// URI:           uri,
	// TransLoopback: transLoopback || transLocalfile,
	// }

	// return rc.Call("URI.Open", p, dummy)
	// })
	// if err != nil {
	// return err
	// }

	if finished != nil {
		<-finished
	}
	return nil
}
