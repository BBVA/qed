/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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

package cmd

import (
	"fmt"
	"io"
	"os/exec"
)

func formatInfo(w io.Writer, title string, body func()) {
	fmt.Fprintf(w, "#### %s\n\n```\n", title)
	body()
	fmt.Fprintf(w, "```\n")
}

func debugInfo(w io.Writer) {
	qedVersion := Ctx.Value(k("version"))
	// Get build version
	formatInfo(w, "Build Info", func() {
		fmt.Fprintf(w, "QED version %s, built in $GOPATH mode\n", qedVersion)
	})

	// Get GO environment info
	formatInfo(w, "Go Info", func() {
		gov := exec.Command("go", "version")
		env := exec.Command("go", "env")
		gov.Stdout = w
		gov.Run()
		env.Stdout = w
		env.Run()
		fmt.Fprint(w)
	})
}
