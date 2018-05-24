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

package scope

import (
	"fmt"
	"testing"
)

type stats struct {
	tasks, ok, skipped int
}

type task struct {
	title string
	run   func(t *testing.T)
}

type Let func(string, func(t *testing.T))
type Scenario func(string, func())
type TestF func(t *testing.T)

func report(format string, v ...interface{}) {
	if testing.Verbose() {
		fmt.Printf(format, v...)
	}
}

func Scope(t *testing.T, before, after TestF) (Scenario, Let) {
	var tasks []task
	var tx task
	var s stats
	var scenarios int

	let := func(title string, run func(t *testing.T)) {
		tasks = append(tasks, task{title, run})
	}

	scenario := func(title string, prepare func()) {
		before(t)
		ok := t.Run(title, func(t *testing.T) {
			report("\n#%d %s\n", scenarios, title)
			tasks = make([]task, 0, 0)
			s = stats{}
			prepare()
			s.tasks = len(tasks)
			for len(tasks) > 0 {
				tx, tasks = tasks[0], tasks[1:]
				report("\t#%d.%d %s → ", scenarios, s.tasks-len(tasks), tx.title)
				tx.run(t)
				report("ok\n")
				s.ok += 1
			}
			scenarios += 1
		})

		after(t)

		if !ok {
			report("failed\n")

			for _, tx := range tasks {
				s.skipped += 1
				report("\t#%d.%d %s → skipped\n", scenarios, s.tasks+s.skipped-len(tasks), tx.title)
			}

			report("FAILED → %+v\n\n", s)
		} else {
			report("SUCCESS → %+v\n\n", s)
		}
	}
	return scenario, let
}
