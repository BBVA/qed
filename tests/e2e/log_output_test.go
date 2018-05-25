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
	"regexp"
	"strings"
	"testing"

	"github.com/bbva/qed/log"
)

func tError(s, level string) {
	log.SetLogger("test", level)
	log.Error(fmt.Sprintf("%s", s))
}
func tErrorf(s, level string) {
	log.SetLogger("test", level)
	log.Errorf(fmt.Sprintf("%s %s", s, "%s"), "composed")
}

func tInfo(s, level string) {
	log.SetLogger("test", level)
	log.Info(fmt.Sprintf("%s", s))
}

func tInfof(s, level string) {
	log.SetLogger("test", level)
	log.Infof(fmt.Sprintf("%s %s", s, "%s"), "composed")
}

func tDebug(s, level string) {
	log.SetLogger("test", level)
	log.Debug(fmt.Sprintf("%s", s))
}

func tDebugf(s, level string) {
	log.SetLogger("test", level)
	log.Debugf(fmt.Sprintf("%s %s", s, "%s"), "composed")
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

	assertSubprocess(t, log.DEBUG, "tError", testString, false, true)
	assertSubprocess(t, log.DEBUG, "tErrorf", testString, false, true)
	assertSubprocess(t, log.DEBUG, "tInfo", testString, false, false)
	assertSubprocess(t, log.DEBUG, "tInfof", testString, false, false)
	assertSubprocess(t, log.DEBUG, "tDebug", testString, false, false)
	assertSubprocess(t, log.DEBUG, "tDebugf", testString, false, false)

	assertSubprocess(t, log.INFO, "tError", testString, false, true)
	assertSubprocess(t, log.INFO, "tErrorf", testString, false, true)
	assertSubprocess(t, log.INFO, "tInfo", testString, false, false)
	assertSubprocess(t, log.INFO, "tInfof", testString, false, false)
	assertSubprocess(t, log.INFO, "tDebug", testString, true, false)
	assertSubprocess(t, log.INFO, "tDebugf", testString, true, false)

	assertSubprocess(t, log.ERROR, "tError", testString, false, true)
	assertSubprocess(t, log.ERROR, "tErrorf", testString, false, true)
	assertSubprocess(t, log.ERROR, "tInfo", testString, true, false)
	assertSubprocess(t, log.ERROR, "tInfof", testString, true, false)
	assertSubprocess(t, log.ERROR, "tDebug", testString, true, false)
	assertSubprocess(t, log.ERROR, "tDebugf", testString, true, false)

	assertSubprocess(t, log.SILENT, "tError", testString, true, false)
	assertSubprocess(t, log.SILENT, "tErrorf", testString, true, false)
	assertSubprocess(t, log.SILENT, "tInfo", testString, true, false)
	assertSubprocess(t, log.SILENT, "tInfof", testString, true, false)
	assertSubprocess(t, log.SILENT, "tDebug", testString, true, false)
	assertSubprocess(t, log.SILENT, "tDebugf", testString, true, false)

}
