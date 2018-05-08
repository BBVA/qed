// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package cli

import (
	"verifiabledata/client"
)

// Context is used to create varios objects across the project and is being passed
// to every command as a constructor argument.
type Context struct {
	client *client.HttpClient
}

func NewContext() *Context {
	return &Context{}
}
