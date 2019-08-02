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

//SimpleTasksManager configuration object used to parse
//cli options and to build the SimpleNotifier instance
type SimpleTasksManagerConfig struct {
	Interval time.Duration `desc:"Interval to execute enqueued tasks"`
	MaxTasks int           `desc:"Maximum number of concurrent tasks"`
}

// Returns the default configuration for the SimpleTasksManager
func DefaultSimpleTasksManagerConfig() *SimpleTasksManagerConfig {
	return &SimpleTasksManagerConfig{
		Interval: 200 * time.Millisecond,
		MaxTasks: 10,
	}
}

func NewSimpleTasksManagerFromConfig(c *SimpleTasksManagerConfig) *SimpleTasksManager {
	return NewSimpleTasksManager(c.Interval, c.MaxTasks)
}

// Simple implementation of a task manager used
// by the QED provided agents
type SimpleTasksManager struct {
	taskCh   chan Task
	quitCh   chan bool
	ticker   *time.Ticker
	maxTasks int
}

// NewTasksManager returns a new TasksManager and its task
// channel. The execution loop will try to execute up to maxTasks tasks
// each interval. Also the channel has maxTasks capacity.
func NewSimpleTasksManager(i time.Duration, max int) *SimpleTasksManager {
	return &SimpleTasksManager{
		taskCh:   make(chan Task, max),
		quitCh:   make(chan bool),
		ticker:   time.NewTicker(i),
		maxTasks: max,
	}
}

// Start activates the task dispatcher
// to execute enqueued tasks
func (t *SimpleTasksManager) Start() {
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
func (t *SimpleTasksManager) Stop() {
	close(t.quitCh)
	t.ticker.Stop()
}

// Add a task to the task manager queue, with the configured timed
// out. It will block until the task is read if the channel is full.
func (t *SimpleTasksManager) Add(task Task) error {
	t.taskCh <- task
	return nil
}

// Len returns the number of pending tasks
// enqueued in the tasks channel
func (t *SimpleTasksManager) Len() int {
	return len(t.taskCh)
}

// dispatchTasks dequeues tasks and
// execute them in different goroutines
// up to MaxInFlightTasks
func (t *SimpleTasksManager) dispatchTasks() {
	count := 0

	for {
		select {
		case task := <-t.taskCh:
			go func() {
				err := task()
				if err != nil {
					log.Infof("Task manager got an error from a task: %v", err)
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

// PrinterFactory create tasks than print BatchSnapshots
// for testing purposes. Its intented to be used with the
// BatchSnapshot processor
type PrinterFactory struct {
}

func (p PrinterFactory) Metrics() []prometheus.Collector {
	return nil
}

func (p PrinterFactory) New(ctx context.Context) Task {
	// a := ctx.Value("agent").(Agent)
	log.Infof("PrinterFactory creating new Task!")
	b := ctx.Value("batch").(*protocol.BatchSnapshots)
	return func() error {
		log.Debugf("Printer Task: agent received batch: %+v\n", b)
		return nil
	}
}
