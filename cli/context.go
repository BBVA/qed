/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package cli

import (
	"github.com/bbva/qed/client"
)

// Context is used to create various objects across the project and is being passed
// to every command as a constructor argument.
type Context struct {
	client *client.HttpClient
}

// NewContext will return a new instance of Context.
func NewContext() *Context {
	return &Context{}
}
