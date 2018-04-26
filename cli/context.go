package cli

import (
	"os"
	"verifiabledata/client"
	"verifiabledata/log"
)

// Context is used to create varios objects across the project and is being passed
// to every command as a constructor argument.
type Context struct {
	logger log.Logger
	client *client.HttpClient
}

func NewContext() *Context {
	return &Context{}
}

// Logger returns the CLI logger.
func (ctx *Context) Logger() log.Logger {
	if ctx.logger == nil {
		ctx.logger = log.NewError(os.Stdout, "QedClient", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	}
	return ctx.logger
}
