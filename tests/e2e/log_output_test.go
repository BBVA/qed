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

package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/testutils/scope"
	assert "github.com/stretchr/testify/require"
)

var testFuncMap = map[string]func(string, string){
	"tError": func(s, level string) {
		log.SetLogger("test", level)
		log.Error(fmt.Sprintf("%s", s))
	},
	"tErrorf": func(s, level string) {
		log.SetLogger("test", level)
		log.Errorf(fmt.Sprintf("%s %s", s, "%s"), "composed")
	},
	"tInfo": func(s, level string) {
		log.SetLogger("test", level)
		log.Info(fmt.Sprintf("%s", s))
	},
	"tInfof": func(s, level string) {
		log.SetLogger("test", level)
		log.Infof(fmt.Sprintf("%s %s", s, "%s"), "composed")
	},
	"tDebug": func(s, level string) {
		log.SetLogger("test", level)
		log.Debug(fmt.Sprintf("%s", s))
	},
	"tDebugf": func(s, level string) {
		log.SetLogger("test", level)
		log.Debugf(fmt.Sprintf("%s %s", s, "%s"), "composed")
	},
}

func subprocess(level, testFunc, message string) (*exec.ExitError, bool, string) {

	cmd := exec.Command(os.Args[0], "-test.run=TestLogSuite")
	var outb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SUBPROCESSED_TEST=%s", testFunc),
		fmt.Sprintf("LOG_LEVEL=%s", level),
	)
	command := cmd.Run()
	err, ok := command.(*exec.ExitError)
	return err, ok, outb.String()

}

func TestLogSuite(t *testing.T) {

	message := "called"

	if cast := os.Getenv("SUBPROCESSED_TEST"); len(cast) > 0 {
		testFuncMap[cast](message, os.Getenv("LOG_LEVEL"))
		return
	}

	scenario, let := scope.Scope(t, func(t *testing.T) {}, func(t *testing.T) {})

	scenario("Test output for log.DEBUG logger", func() {

		let("test Error function", func(t *testing.T) {
			err, ok, outString := subprocess(log.DEBUG, "tError", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
			assert.True(t, ok && !err.Success(), "Subprocess must exit with status 1")
		})

		let("test Errorf function", func(t *testing.T) {
			err, ok, outString := subprocess(log.DEBUG, "tErrorf", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
			assert.True(t, ok && !err.Success(), "Subprocess must exit with status 1")
			assert.True(t, strings.Contains(outString, "composed"), "If log formatted must show formatted strings")
		})

		let("test Info function", func(t *testing.T) {
			_, _, outString := subprocess(log.DEBUG, "tInfo", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
		})

		let("test Infof function", func(t *testing.T) {
			_, _, outString := subprocess(log.DEBUG, "tInfof", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
			assert.True(t, strings.Contains(outString, "composed"), "If log formatted must show formatted strings")
		})

		let("test Debug function", func(t *testing.T) {
			_, _, outString := subprocess(log.DEBUG, "tDebug", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
		})

		let("test Debugf function", func(t *testing.T) {
			_, _, outString := subprocess(log.DEBUG, "tDebugf", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
			assert.True(t, strings.Contains(outString, "composed"), "If log formatted must show formatted strings")
		})

	})

	scenario("Test output for log.INFO logger", func() {

		let("test Error function", func(t *testing.T) {
			err, ok, outString := subprocess(log.INFO, "tError", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
			assert.True(t, ok && !err.Success(), "Subprocess must exit with status 1")
		})

		let("test Errorf function", func(t *testing.T) {
			err, ok, outString := subprocess(log.INFO, "tErrorf", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
			assert.True(t, ok && !err.Success(), "Subprocess must exit with status 1")
			assert.True(t, strings.Contains(outString, "composed"), "If log formatted must show formatted strings")
		})

		let("test Info function", func(t *testing.T) {
			_, _, outString := subprocess(log.INFO, "tInfo", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
		})

		let("test Infof function", func(t *testing.T) {
			_, _, outString := subprocess(log.INFO, "tInfof", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
			assert.True(t, strings.Contains(outString, "composed"), "If log formatted must show formatted strings")
		})

		let("test Debug function", func(t *testing.T) {
			_, _, outString := subprocess(log.INFO, "tDebug", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

		let("test Debugf function", func(t *testing.T) {
			_, _, outString := subprocess(log.INFO, "tDebugf", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

	})

	scenario("Test output for log.ERROR logger", func() {

		let("test Error function", func(t *testing.T) {
			err, ok, outString := subprocess(log.ERROR, "tError", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
			assert.True(t, ok && !err.Success(), "Subprocess must exit with status 1")
		})

		let("test Errorf function", func(t *testing.T) {
			err, ok, outString := subprocess(log.ERROR, "tErrorf", message)
			assert.True(t, strings.Contains(outString, message), "Must show message string")
			assert.True(t, ok && !err.Success(), "Subprocess must exit with status 1")
			assert.True(t, strings.Contains(outString, "composed"), "If log formatted must show formatted strings")
		})

		let("test Info function", func(t *testing.T) {
			_, _, outString := subprocess(log.ERROR, "tInfo", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

		let("test Infof function", func(t *testing.T) {
			_, _, outString := subprocess(log.ERROR, "tInfof", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

		let("test Debug function", func(t *testing.T) {
			_, _, outString := subprocess(log.ERROR, "tDebug", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

		let("test Debugf function", func(t *testing.T) {
			_, _, outString := subprocess(log.ERROR, "tDebugf", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

	})

	scenario("Test output for log.SILENT logger", func() {

		let("test Error function", func(t *testing.T) {
			_, _, outString := subprocess(log.SILENT, "tError", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

		let("test Errorf function", func(t *testing.T) {
			_, _, outString := subprocess(log.SILENT, "tErrorf", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

		let("test Info function", func(t *testing.T) {
			_, _, outString := subprocess(log.SILENT, "tInfo", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

		let("test Infof function", func(t *testing.T) {
			_, _, outString := subprocess(log.SILENT, "tInfof", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

		let("test Debug function", func(t *testing.T) {
			_, _, outString := subprocess(log.SILENT, "tDebug", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

		let("test Debugf function", func(t *testing.T) {
			_, _, outString := subprocess(log.SILENT, "tDebugf", message)
			assert.False(t, strings.Contains(outString, message), "Must not show message string")
		})

	})

}
