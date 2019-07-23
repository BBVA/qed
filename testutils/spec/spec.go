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

package spec

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type TestF func(t testing.TB)

type LetF func(t *testing.T, desc string, fns ...TestF)
type ReportF func() string

func New() (LetF, ReportF) {
	report := make([]string, 1)
	return func(t *testing.T, desc string, fns ...TestF) {
			t.Helper()
			var ri int
			idx := strings.Count(t.Name(), "/")
			for _, fn := range fns {
				status := t.Run(strconv.Itoa(idx), func(t *testing.T) {
					report = append(report, t.Name()+": "+desc)
					ri = len(report) - 1
					fmt.Println(desc)
					fn(t)
				})
				switch {
				case status == true:
					report[ri] += " -> ok!"
				case status == false:
					report[ri] += " -> failed!"
				}
			}
		}, func() string {
			return strings.Join(report, "\n")
		}
}

func Equal(t testing.TB, exp, got interface{}, msg string) {
	t.Helper()
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("Not equals: %s -> expecting '%v' got '%v'\n", msg, exp, got)
	}
}

func True(t testing.TB, cond bool, msg string) {
	t.Helper()
	if cond != true {
		t.Fatalf("Condition is not true: %s -> %v", msg, cond)
	}
}

func False(t testing.TB, cond bool, msg string) {
	t.Helper()
	if cond != false {
		t.Fatalf("Condition is not false: %s -> %v", msg, cond)
	}
}

func NoError(t testing.TB, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("Error is not nil: %s -> %v", msg, err)
	}
}

func Error(t testing.TB, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Error is not nil: %s -> %v", msg, err)
	}
}

func isNil(object interface{}) bool {

	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)

	if value.IsNil() {
		return true
	}

	return false
}

func NotNil(t testing.TB, object interface{}, msg string) {
	t.Helper()
	if isNil(object) {
		t.Fatalf("Object is nil: %v --> %v", msg, object)
	}
}

func Retry(t testing.TB, tries int, delay time.Duration, fn func() error) {
	t.Helper()
	var i int
	var err error
	for i = 0; i < tries; i++ {
		err = fn()
		if err == nil {
			return
		}
		time.Sleep(delay)
	}
	if err != nil {
		t.Fatalf("Error in condition:: try %v --> err %v", i, err)
	}
	if err == nil {
		t.Fatalf("Retry timed out in try %v (delay %v)", tries, delay)
	}
}

