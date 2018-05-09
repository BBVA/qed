package cli

import (
	"qed/client"
)

// Context is used to create varios objects across the project and is being passed
// to every command as a constructor argument.
type Context struct {
	client *client.HttpClient
}

func NewContext() *Context {
	return &Context{}
}
