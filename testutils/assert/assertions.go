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

package assert

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

// Equal asserts that the values are equal and returns true
// if the assertion was successful.
func Equal(t *testing.T, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	if areEqual(actual, expected) {
		return true
	}

	actual, expected = formatValues(actual, expected)
	return fail(t, fmt.Sprintf("Not equal: actual: %s, expected: %s", actual, expected), msgAndArgs...)
}

// NotEqual asserts that the values are not equal and returns
// true if the assertion was successful.
func NotEqual(t *testing.T, actual, expected interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	if !areEqual(actual, expected) {
		return true
	}

	actual, expected = formatValues(actual, expected)
	return fail(t, fmt.Sprintf("Equal: %s should not be %s", actual, expected), msgAndArgs...)
}

// True asserts that the value is true and returns true
// if the assertion was successful.
func True(t *testing.T, value bool, msgAndArgs ...interface{}) bool {
	t.Helper()

	if value {
		return true
	}

	return fail(t, fmt.Sprintf("True: %v should be true", value), msgAndArgs...)
}

// False asserts that the value is false and returns true
// if the assertion was successful.
func False(t *testing.T, value bool, msgAndArgs ...interface{}) bool {
	t.Helper()

	if !value {
		return true
	}

	return fail(t, fmt.Sprintf("True: %v should be false", value), msgAndArgs...)
}

func Len(t *testing.T, object interface{}, expected int, msgAndArgs ...interface{}) bool {
	t.Helper()

	ok, actual := getLen(object)

	if !ok {
		return fail(t, fmt.Sprintf("Len: \"%s\" couuld not be applied builtin len()", object), msgAndArgs...)
	}

	if actual != expected {
		return fail(t, fmt.Sprintf("Len: \"%s\" should have %d item(s), but has %d", object, expected, actual), msgAndArgs...)
	}

	return true
}

// getLen try to get length of object.
// return (false, 0) if impossible.
func getLen(x interface{}) (ok bool, length int) {
	v := reflect.ValueOf(x)
	defer func() {
		if e := recover(); e != nil {
			ok = false
		}
	}()
	return true, v.Len()
}

// NoError asserts that err is nil and returns true
// if the assertion was successful.
func NoError(t *testing.T, err error, msgAndArgs ...interface{}) bool {
	if err != nil {
		return fail(t, fmt.Sprintf("unexpected error: %+v", err), msgAndArgs...)
	}
	return true
}

func fail(t *testing.T, failureMsg string, msgAndArgs ...interface{}) bool {
	t.Helper()
	msg := formatMsgAndArgs(msgAndArgs...)
	if msg == "" {
		t.Errorf("\nError from %s", failureMsg)
	} else {
		t.Errorf("\nError from %s\nMessages: %s", failureMsg, msg)
	}
	return false
}

func formatValues(actual, expected interface{}) (string, string) {
	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		return fmt.Sprintf("%T(%#v)", actual, actual), fmt.Sprintf("%T(%#v)", expected, expected)
	}
	return fmt.Sprintf("%#v", actual), fmt.Sprintf("%#v", expected)
}

func formatMsgAndArgs(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return ""
	}
	if len(msgAndArgs) == 1 {
		return msgAndArgs[0].(string)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}

func areEqual(actual, expected interface{}) bool {
	if actual == nil || expected == nil {
		return actual == expected
	}

	if exp, ok := expected.([]byte); ok {
		act, ok := actual.([]byte)
		if !ok {
			return false
		} else if exp == nil || act == nil {
			return exp == nil && act == nil
		}
		return bytes.Equal(exp, act)
	}

	return reflect.DeepEqual(expected, actual)
}
