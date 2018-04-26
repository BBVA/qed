package cli

import (
	"log"
	"os"
	"verifiabledata/client"
)

// Context is used to create varios objects across the project and is being passed
// to every command as a constructor argument.
type Context struct {
	logger *log.Logger
	client *client.HttpClient
}

func NewContext() *Context {
	return &Context{}
}

// Logger returns the CLI logger.
func (ctx *Context) Logger() *log.Logger {
	if ctx.logger == nil {
		ctx.logger = log.New(os.Stdout, "QED", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	}
	return ctx.logger
}
