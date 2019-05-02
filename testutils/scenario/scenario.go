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

package scenario

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type TestF func(t *testing.T)

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

func Equal(t *testing.T, exp, got interface{}, msg string) {
	t.Helper()
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("Expecting '%v' got '%v' %s\n", exp, got, msg)
	}
}

func True(t *testing.T, cond bool, msg string) {
	t.Helper()
	if cond != true {
		t.Fatalf("Condition is not true: %s", msg)
	}
}

func False(t *testing.T, cond bool, msg string) {
	t.Helper()
	if cond != false {
		t.Fatalf("Condition is not false: %s", msg)
	}
}

func NoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("Error is not nil: %s", msg)
	}
}
