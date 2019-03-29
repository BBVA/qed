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
package gossip

import (
	"time"

	"github.com/bbva/qed/log"
)

// Task are executed by the TasksManager loop
// calling to their Do method.
// All tasks must have their own context already
// when they are enqueued into the task manager.
//
// For example:
//
// func TaskFactory(ctx contex.Context) func() {
//	cts.Set("Test") = "test)
//	return func() {
//		fmt.Println(ctx.Get("Test"))
//	}
// }
type Task func() error

// TasksManager executes enqueued tasks, It is in charge
// of applying limits to task execution such as timeouts.
// It only has an API to stop and start the tasks execution
// loop.
type TasksManager struct {
	taskCh         chan Task
	quitCh         chan bool
	ticker         *time.Ticker
	enqueueTimeout *time.Ticker
	maxTasks       int
}

// NewTasksManager returns a new TasksManager and its task
// channel. The execution loop will try to execute up to maxTasks tasks
// each interval.
func NewTasksManager(interval time.Duration, maxTasks int, enqueueTimeout time.Duration) *TasksManager {
	taskCh := make(chan Task, 10)
	return &TasksManager{
		taskCh:         taskCh,
		quitCh:         make(chan bool),
		ticker:         time.NewTicker(interval),
		enqueueTimeout: time.NewTicker(enqueueTimeout),
		maxTasks:       maxTasks,
	}
}

// Start activates the task dispatcher
// to execute enqueued tasks
func (t *TasksManager) Start() {
	go func() {
		for {
			select {
			case <-t.ticker.C:
				go t.dispatchTasks()
			case <-t.quitCh:

				return
			}
		}
	}()
}

// Stop disables the task dispatcher
// It does not wait to empty the
// task queue nor closes the task channel.
func (t *TasksManager) Stop() {
	close(t.quitCh)
	t.ticker.Stop()
	t.enqueueTimeout.Stop()
}

// Add a task to the task manager queue, with the configured timed
// out. It will block the timeout duration in the worst case and can be
// called from multiple goroutines.
func (t *TasksManager) Add(task Task) error {
	for {
		select {
		case <-t.enqueueTimeout.C:
			return ChTimedOut
		case t.taskCh <- task:
			return nil
		}
	}
}

// Len returns the number of pending tasks
// enqueued in the tasks channel
func (t *TasksManager) Len() int {
	return len(t.taskCh)
}

// dispatchTasks dequeues tasks and
// execute them in different goroutines
// up to MaxInFlightTasks
func (t *TasksManager) dispatchTasks() {
	count := 0

	for {
		select {
		case task := <-t.taskCh:
			go func() {
				err := task()
				if err != nil {
					log.Infof("Agent task manager got an error from a task: %v", err)
				}
			}()
			count++
		default:
			return
		}
		if count >= t.maxTasks {
			return
		}
	}
}
