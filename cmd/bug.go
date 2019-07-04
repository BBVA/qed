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
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/bbva/qed/log"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"
)

// Commands returns a list of commands to use to open a url.
func Commands() [][]string {
	var cmds [][]string
	if exec := os.Getenv("BROWSER"); exec != "" {
		cmds = append(cmds, []string{exec})
	}
	switch runtime.GOOS {
	case "windows":
		cmds = append(cmds, []string{"cmd", "/c", "start"})
	case "darwin":
		cmds = append(cmds, []string{"/usr/bin/open"})
	default:
		if os.Getenv("DISPLAY") != "" {
			cmds = append(cmds, []string{"xdg-open"})
		}
	}
	cmds = append(cmds,
		[]string{"chromium"},
		[]string{"chrome"},
		[]string{"google-chrome"},
		[]string{"firefox"},
	)
	return cmds
}

func cmdSuccessful(cmd *exec.Cmd, timeout time.Duration) bool {
	errc := make(chan error, 1)
	go func() {
		errc <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		return true
	case err := <-errc:
		return err == nil
	}
}

// Open tries to open url in a browser and reports its status.
func Open(url string) bool {
	for _, args := range Commands() {
		cmd := exec.Command(args[0], append(args[1:], url)...)
		if cmd.Start() == nil && cmdSuccessful(cmd, 3*time.Second) {
			return true
		}
	}
	return false
}

// bug implements the bug command.
type bug struct{}

func (b *bug) Name() string { return "bug" }

const qedBugPrefix = "QED: "
const qedBugHeader = `Please answer these questions before submitting your issue. Thanks!

#### What did you do?
If possible, provide a recipe for reproducing the error.
A complete runnable program is good.
A link on play.golang.org is better.

#### What did you expect to see?

#### What did you see instead?

`

type BugConfig struct {
	// Set non default issue description.
	Desc string `desc:"Write custom issue description"`
}

func bugDefaultConfig() *BugConfig {
	return &BugConfig{
		Desc: "",
	}
}

var bugCmd *cobra.Command = &cobra.Command{
	Use:   "bug",
	Short: "Generates new issue template in Github.",
	Long: `This command generates new issue template in Github with
useful information for further debugging.`,
	TraverseChildren: true,
	RunE:             runBug,
}

var bugCtx context.Context

func init() {
	bugCtx = bugConfig()
	Root.AddCommand(bugCmd)
}

func bugConfig() context.Context {

	conf := bugDefaultConfig()

	err := gpflag.ParseTo(conf, bugCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	return context.WithValue(Ctx, k("bug.config"), conf)
}

func runBug(cmd *cobra.Command, args []string) error {
	conf := bugCtx.Value(k("bug.config")).(*BugConfig)

	// Build issue metadata
	buf := &bytes.Buffer{}
	fmt.Fprint(buf, qedBugHeader)
	debugInfo(buf)
	body := buf.String()
	title := conf.Desc

	if !strings.HasPrefix(title, qedBugPrefix) {
		title = qedBugPrefix + title
	}

	if !Open("https://github.com/bbva/qed/issues/new?title=" + url.QueryEscape(title) + "&body=" + url.QueryEscape(body)) {
		fmt.Print("Please file a new issue at github.com/bbva/qed/issues/new using this template:\n\n")
		fmt.Print(body)
	}
	return nil
}
