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
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func tError(s, level string) {
	SetLogger("test", level)
	Error(fmt.Sprintf("%s", s))
}
func tErrorf(s, level string) {
	SetLogger("test", level)
	Errorf(fmt.Sprintf("%s %s", s, "%s"), "composed")
}

func tInfo(s, level string) {
	SetLogger("test", level)
	Info(fmt.Sprintf("%s", s))
}

func tInfof(s, level string) {
	SetLogger("test", level)
	Infof(fmt.Sprintf("%s %s", s, "%s"), "composed")
}

func tDebug(s, level string) {
	SetLogger("test", level)
	Debug(fmt.Sprintf("%s", s))
}

func tDebugf(s, level string) {
	SetLogger("test", level)
	Debugf(fmt.Sprintf("%s %s", s, "%s"), "composed")
}

var testFuncMap = map[string]func(string, string){
	"tError":  tError,
	"tErrorf": tErrorf,
	"tInfo":   tInfo,
	"tInfof":  tInfof,
	"tDebug":  tDebug,
	"tDebugf": tDebugf,
}

func assertSubprocess(t *testing.T, level, testFunc, message string, silent, exit bool) {

	cmd := exec.Command(os.Args[0], "-test.run=TestLogSuite")
	var outb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SUBPROCESSED_TEST=%s", testFunc),
		fmt.Sprintf("LOG_LEVEL=%s", level),
	)

	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); exit && (!ok || e.Success()) {
		t.Errorf("log.Error ran with err %v, want exit status 1", err)
	}

	outString := outb.String()

	if !silent {
		if !strings.Contains(outString, message) {
			t.Errorf("No message emmited %s %s", testFunc, level)
		}
		if regexp.MustCompile("f$").MatchString(testFunc) && !strings.Contains(outString, "composed") {
			t.Errorf("No composed message emmited %s %s", testFunc, level)
		}
	}

	if silent && strings.Contains(outString, message) {
		t.Errorf("Stdout emmited %s: '%s'", testFunc, level)
	}

}

func TestLogSuite(t *testing.T) {

	testString := "called"

	if cast := os.Getenv("SUBPROCESSED_TEST"); len(cast) > 0 {
		testFuncMap[cast](testString, os.Getenv("LOG_LEVEL"))
		return
	}

	assertSubprocess(t, DEBUG, "tError", testString, false, true)
	assertSubprocess(t, DEBUG, "tErrorf", testString, false, true)
	assertSubprocess(t, DEBUG, "tInfo", testString, false, false)
	assertSubprocess(t, DEBUG, "tInfof", testString, false, false)
	assertSubprocess(t, DEBUG, "tDebug", testString, false, false)
	assertSubprocess(t, DEBUG, "tDebugf", testString, false, false)

	assertSubprocess(t, INFO, "tError", testString, false, true)
	assertSubprocess(t, INFO, "tErrorf", testString, false, true)
	assertSubprocess(t, INFO, "tInfo", testString, false, false)
	assertSubprocess(t, INFO, "tInfof", testString, false, false)
	assertSubprocess(t, INFO, "tDebug", testString, true, false)
	assertSubprocess(t, INFO, "tDebugf", testString, true, false)

	assertSubprocess(t, ERROR, "tError", testString, false, true)
	assertSubprocess(t, ERROR, "tErrorf", testString, false, true)
	assertSubprocess(t, ERROR, "tInfo", testString, true, false)
	assertSubprocess(t, ERROR, "tInfof", testString, true, false)
	assertSubprocess(t, ERROR, "tDebug", testString, true, false)
	assertSubprocess(t, ERROR, "tDebugf", testString, true, false)

	assertSubprocess(t, SILENT, "tError", testString, true, false)
	assertSubprocess(t, SILENT, "tErrorf", testString, true, false)
	assertSubprocess(t, SILENT, "tInfo", testString, true, false)
	assertSubprocess(t, SILENT, "tInfof", testString, true, false)
	assertSubprocess(t, SILENT, "tDebug", testString, true, false)
	assertSubprocess(t, SILENT, "tDebugf", testString, true, false)

}
