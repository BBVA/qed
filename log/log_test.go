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

package log

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"
)

var testFuncMap = map[string]func(string, string){
	"tError": func(s, level string) {
		SetLogger("test", level)
		Error(fmt.Sprintf("%s", s))
	},
	"tErrorf": func(s, level string) {
		SetLogger("test", level)
		Errorf(fmt.Sprintf("%s %s", s, "%s"), "composed")
	},
	"tInfo": func(s, level string) {
		SetLogger("test", level)
		Info(fmt.Sprintf("%s", s))
	},
	"tInfof": func(s, level string) {
		SetLogger("test", level)
		Infof(fmt.Sprintf("%s %s", s, "%s"), "composed")
	},
	"tDebug": func(s, level string) {
		SetLogger("test", level)
		Debug(fmt.Sprintf("%s", s))
	},
	"tDebugf": func(s, level string) {
		SetLogger("test", level)
		Debugf(fmt.Sprintf("%s %s", s, "%s"), "composed")
	},
}

func subprocess(level, testFunc, message string) (*exec.ExitError, bool, string) {

	cmd := exec.Command(os.Args[0], "-test.run=TestLogIntegration")
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

func assertOutput(t *testing.T, exit, silent bool, level, testFunc, message string) {

	err, ok, outString := subprocess(level, testFunc, message)

	if exit {
		assert.True(t, ok && !err.Success(), "Subprocess must exit with status 1")
	} else {
		assert.False(t, ok && !err.Success(), "Subprocess must not exit with status 1")
	}

	if !silent {
		assert.True(t, strings.Contains(outString, message), "Must show message string")

		if regexp.MustCompile("f$").MatchString(testFunc) {
			assert.True(t, strings.Contains(outString, "composed"), "Must show formatted strings")
		}

	} else {
		assert.False(t, strings.Contains(outString, message), "Must not show message string")
	}

}

func TestLogIntegration(t *testing.T) {

	message := "called"

	if cast := os.Getenv("SUBPROCESSED_TEST"); len(cast) > 0 {
		testFuncMap[cast](message, os.Getenv("LOG_LEVEL"))
		return
	}

	assertOutput(t, true, false, DEBUG, "tError", message)
	assertOutput(t, true, false, DEBUG, "tErrorf", message)
	assertOutput(t, false, false, DEBUG, "tInfo", message)
	assertOutput(t, false, false, DEBUG, "tInfof", message)
	assertOutput(t, false, false, DEBUG, "tDebug", message)
	assertOutput(t, false, false, DEBUG, "tDebugf", message)

	assertOutput(t, true, false, INFO, "tError", message)
	assertOutput(t, true, false, INFO, "tErrorf", message)
	assertOutput(t, false, false, INFO, "tInfo", message)
	assertOutput(t, false, false, INFO, "tInfof", message)
	assertOutput(t, false, true, INFO, "tDebug", message)
	assertOutput(t, false, true, INFO, "tDebugf", message)

	assertOutput(t, true, false, ERROR, "tError", message)
	assertOutput(t, true, false, ERROR, "tErrorf", message)
	assertOutput(t, false, true, ERROR, "tInfo", message)
	assertOutput(t, false, true, ERROR, "tInfof", message)
	assertOutput(t, false, true, ERROR, "tDebug", message)
	assertOutput(t, false, true, ERROR, "tDebugf", message)

	assertOutput(t, true, true, SILENT, "tError", message)
	assertOutput(t, true, true, SILENT, "tErrorf", message)
	assertOutput(t, false, true, SILENT, "tInfo", message)
	assertOutput(t, false, true, SILENT, "tInfof", message)
	assertOutput(t, false, true, SILENT, "tDebug", message)
	assertOutput(t, false, true, SILENT, "tDebugf", message)

}

func TestDebug(t *testing.T) {
	var original = osExit
	osExit = func(n int) {}
	defer func() { osExit = original }()

	var out bytes.Buffer
	tLog := newDebug(&out, "test", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)

	tLog.Error("message1")
	assert.Regexp(t, "message1", out.String())
	out.Reset()

	tLog.Errorf("message2 %s", "comp")
	assert.Regexp(t, "message2 comp", out.String())
	out.Reset()

	tLog.Info("message3")
	assert.Regexp(t, "message3", out.String())
	out.Reset()

	tLog.Infof("message4 %s", "comp")
	assert.Regexp(t, "message4 comp", out.String())
	out.Reset()

	tLog.Debug("message5")
	assert.Regexp(t, "message5", out.String())
	out.Reset()

	tLog.Debugf("message6 %s", "comp")
	assert.Regexp(t, "message6 comp", out.String())
	out.Reset()

	assert.IsType(t, &log.Logger{}, tLog.GetLogger())

}

func TestInfo(t *testing.T) {

	var original = osExit
	osExit = func(n int) {}
	defer func() { osExit = original }()

	var out bytes.Buffer
	tLog := newInfo(&out, "test", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)

	tLog.Error("message1")
	assert.Regexp(t, "message1", out.String())
	out.Reset()

	tLog.Errorf("message2 %s", "comp")
	assert.Regexp(t, "message2 comp", out.String())
	out.Reset()

	tLog.Info("message3")
	assert.Regexp(t, "message3", out.String())
	out.Reset()

	tLog.Infof("message4 %s", "comp")
	assert.Regexp(t, "message4 comp", out.String())
	out.Reset()

	tLog.Debug("message5")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	tLog.Debugf("message6 %s", "comp")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	assert.IsType(t, &log.Logger{}, tLog.GetLogger())

}

func TestError(t *testing.T) {

	var original = osExit
	osExit = func(n int) {}
	defer func() { osExit = original }()

	var out bytes.Buffer
	tLog := newError(&out, "test", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)

	tLog.Error("message1")
	assert.Regexp(t, "message1", out.String())
	out.Reset()

	tLog.Errorf("message2 %s", "comp")
	assert.Regexp(t, "message2 comp", out.String())
	out.Reset()

	tLog.Info("message3")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	tLog.Infof("message4 %s", "comp")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	tLog.Debug("message5")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	tLog.Debugf("message6 %s", "comp")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	assert.IsType(t, &log.Logger{}, tLog.GetLogger())
}

func TestSilent(t *testing.T) {

	var original = osExit
	osExit = func(n int) {}
	defer func() { osExit = original }()

	tLog := newSilent()

	var out bytes.Buffer
	tLog.Error("message1")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	tLog.Errorf("message2 %s", "comp")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	tLog.Info("message1")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	tLog.Infof("message2 %s", "comp")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	tLog.Debug("message1")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	tLog.Debugf("message2 %s", "comp")
	assert.Regexp(t, "^$", out.String())
	out.Reset()

	assert.IsType(t, &log.Logger{}, tLog.GetLogger())
}
