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

package build

import (
	"fmt"
	"runtime"
	"time"
)

// TimeFormat is the reference format for build.Time. Make sure it stays in sync
// with the string passed to the linker in the goreleaser config file
const TimeFormat = "2006/01/02 15:04:05"

var (
	// These variables are initialized via the linker -X flag in the
	// goleaser config file when compiling release binaries.
	tag         = "unknown" // Tag of this build (git describe --tags)
	utcTime     string      // Build time in UTC (year/month/day hour:min:sec)
	rev         string      // SHA-1 of this build (git rev-parse)
	cgoCompiler = cgoVersion()
	platform    = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
)

// Info stores the build information
type Info struct {
	GoVersion   string
	Tag         string
	Time        string
	Revision    string
	CgoCompiler string
	Platform    string
}

// Short returns a pretty printed build and version summary.
func (i Info) Short() string {
	return fmt.Sprintf("QED %s (%s, built %s, %s)",
		i.Tag, i.Platform, i.Time, i.GoVersion)
}

// GoTime parses the utcTime string and returns a time.Time.
func (i Info) GoTime() time.Time {
	val, err := time.Parse(TimeFormat, i.Time)
	if err != nil {
		return time.Time{}
	}
	return val
}

// Timestamp parses the utcTime string and returns the number of seconds since epoch.
func (i Info) Timestamp() (int64, error) {
	val, err := time.Parse(TimeFormat, i.Time)
	if err != nil {
		return 0, err
	}
	return val.Unix(), nil
}

// GetInfo returns an Info struct populated with the build information.
func GetInfo() Info {
	return Info{
		GoVersion:   runtime.Version(),
		Tag:         tag,
		Time:        utcTime,
		Revision:    rev,
		CgoCompiler: cgoCompiler,
		Platform:    platform,
	}
}
