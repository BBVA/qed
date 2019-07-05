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

func formatInfo(w io.Writer, title string, body func() error) error {
	fmt.Fprintf(w, "#### %s\n\n```\n", title)
	err := body()
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "```\n")
	return nil
}

func debugInfo(w io.Writer) error {
	var err error
	qedVersion := Ctx.Value(k("version"))
	// Get build version
	formatInfo(w, "Build Info", func() error {
		fmt.Fprintf(w, "QED version %s, built in $GOPATH mode\n", qedVersion)
		return nil
	})

	// Get GO environment info
	err = formatInfo(w, "Go Info", func() error {
		gov := exec.Command("go", "version")
		env := exec.Command("go", "env")
		gov.Stdout = w
		env.Stdout = w

		err = gov.Run()
		if err != nil {
			return fmt.Errorf("Getting Go version: %v", err)
		}

		err = env.Run()
		if err != nil {
			return fmt.Errorf("Getting Go environment: %v", err)
		}

		fmt.Fprint(w)
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
