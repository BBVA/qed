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
	"context"
	"fmt"
	"time"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/prometheus/client_golang/prometheus"
)

// Task are executed by the TasksManager loop
// calling to their Do method.
// All tasks must have their own context already
// when they are enqueued into the task manager.
//
// For example:
// // complies with the Processor interface
// func TaskFactory(ctx contex.Context) Task {
//	a := ctx.Value("agent").(gossip.Agent)
//	b := ctx.Value("batch").(*protocol.BatchSnapshots)
//	return func() error {
//		fmt.Println(a.Send(b))
//		return nil
//	}
// }
type Task func() error

// A task factory builds tasks with the
// provided context information.
//
// The context contains the agent API
// and the context added by the message
// processor.
//
// The implenetor of the task factory must know
// the details of the data included by the processor
// in the context.
type TaskFactory interface {
	New(context.Context) Task
	Metrics() []prometheus.Collector
}

// TasksManager executes enqueued tasks, It is in charge
// of applying limits to task execution such as timeouts.
// It only has an API to stop and start the tasks execution
// loop.
type TasksManager interface {
	Start()
	Stop()
	Add(t Task) error
	Len() int
}

type DefaultTasksManagerConfig struct {
	QueueTimeout time.Duration
	Interval     time.Duration
	MaxTasks     int
}

func NewDefaultTasksManagerFromConfig(c *DefaultTasksManagerConfig) *DefaultTasksManager {
	return NewDefaultTasksManager(c.QueueTimeout, c.Interval, c.MaxTasks)
}

// Default implementation of a task manager used
// by the QED provided agents
type DefaultTasksManager struct {
	taskCh         chan Task
	quitCh         chan bool
	ticker         *time.Ticker
	enqueueTimeout *time.Ticker
	maxTasks       int
}

// NewTasksManager returns a new TasksManager and its task
// channel. The execution loop will try to execute up to maxTasks tasks
// each interval.
func NewDefaultTasksManager(i, t time.Duration, max int) *DefaultTasksManager {
	return &DefaultTasksManager{
		taskCh:         make(chan Task, max),
		quitCh:         make(chan bool),
		ticker:         time.NewTicker(i),
		enqueueTimeout: time.NewTicker(t),
		maxTasks:       max,
	}
}

// Start activates the task dispatcher
// to execute enqueued tasks
func (t *DefaultTasksManager) Start() {
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
func (t *DefaultTasksManager) Stop() {
	close(t.quitCh)
	t.ticker.Stop()
	t.enqueueTimeout.Stop()
}

// Add a task to the task manager queue, with the configured timed
// out. It will block the timeout duration in the worst case and can be
// called from multiple goroutines.
func (t *DefaultTasksManager) Add(task Task) error {
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
func (t *DefaultTasksManager) Len() int {
	return len(t.taskCh)
}

// dispatchTasks dequeues tasks and
// execute them in different goroutines
// up to MaxInFlightTasks
func (t *DefaultTasksManager) dispatchTasks() {
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

type PrinterFactory struct {
}

func (p PrinterFactory) Metrics() []prometheus.Collector {
	return nil
}

func (p PrinterFactory) New(ctx context.Context) Task {
	// a := ctx.Value("agent").(Agent)
	b := ctx.Value("batch").(*protocol.BatchSnapshots)
	return func() error {
		fmt.Printf("Printer Task: agent received batch: %+v\n", b)
		return nil
	}
}
