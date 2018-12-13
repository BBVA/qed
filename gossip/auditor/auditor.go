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

package auditor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
)

type Config struct {
	QEDUrls               []string
	PubUrls               []string
	APIKey                string
	TaskExecutionInterval time.Duration
	MaxInFlightTasks      int
}

func DefaultConfig() *Config {
	return &Config{
		TaskExecutionInterval: 200 * time.Millisecond,
		MaxInFlightTasks:      10,
	}
}

type Auditor struct {
	qed  *client.HttpClient
	conf *Config

	taskCh          chan Task
	quitCh          chan bool
	executionTicker *time.Ticker
}

func NewAuditor(conf *Config) (*Auditor, error) {
	auditor := &Auditor{
		qed:    client.NewHttpClient(conf.QEDUrls[0], conf.APIKey),
		conf:   conf,
		taskCh: make(chan Task, 100),
		quitCh: make(chan bool),
	}

	go auditor.runTaskDispatcher()

	return auditor, nil
}

type Task interface {
	Do()
}

func (t *MembershipTask) getSnapshot(version uint64) (*protocol.SignedSnapshot, error) {
	resp, err := http.Get(fmt.Sprintf("%s/snapshot?v=%d", t.pubUrl, version))
	if err != nil {
		return nil, fmt.Errorf("Error getting snapshot from the store: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error getting snapshot from the store. Status: %d", resp.StatusCode)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	var s protocol.SignedSnapshot
	err = s.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("Error decoding signed snapshot %d codec", t.s.Snapshot.Version)
	}
	return &s, nil
}

type MembershipTask struct {
	qed    *client.HttpClient
	pubUrl string
	taskCh chan Task
	s      *protocol.SignedSnapshot
}

func (t *MembershipTask) Do() {
	proof, err := t.qed.MembershipDigest(t.s.Snapshot.EventDigest, t.s.Snapshot.Version)
	if err != nil {
		// retry
		log.Errorf("Error executing membership query: %v", err)
	}

	snap, err := t.getSnapshot(proof.CurrentVersion)
	if err != nil {
		log.Infof("Unable to get snapshot from storage, try later: %v", err)
		t.taskCh <- t
		return
	}
	checkSnap := &protocol.Snapshot{
		HistoryDigest: t.s.Snapshot.HistoryDigest,
		HyperDigest:   snap.Snapshot.HyperDigest,
		Version:       t.s.Snapshot.Version,
		EventDigest:   t.s.Snapshot.EventDigest,
	}
	ok := t.qed.Verify(proof, checkSnap, hashing.NewSha256Hasher)
	if !ok {
		t.sendAlert(fmt.Sprintf("Unable to verify snapshot %v", t.s.Snapshot))
		log.Infof("Unable to verify snapshot %v", t.s.Snapshot)
	}
	log.Infof("MembershipTask.Do(): Snapshot %v has been verified by QED", t.s.Snapshot)
}

func (t *MembershipTask) sendAlert(msg string) {

	go func() {
		resp, err := http.Post(fmt.Sprintf("%s/alert", t.pubUrl), "application/json", bytes.NewBufferString(msg))
		if err != nil {
			log.Infof("Error saving batch in alertStore: %v", err)
		}
		defer resp.Body.Close()
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Infof("Error getting response from alertStore saving a batch: %v", err)
		}
	}()
}

func (a Auditor) Process(b *protocol.BatchSnapshots) {

	task := &MembershipTask{
		qed:    a.qed,
		pubUrl: a.conf.PubUrls[0],
		taskCh: a.taskCh,
		s:      b.Snapshots[0],
	}

	a.taskCh <- task
}

func (a *Auditor) runTaskDispatcher() {
	a.executionTicker = time.NewTicker(a.conf.TaskExecutionInterval)
	for {
		select {
		case <-a.executionTicker.C:
			log.Debug("Dispatching tasks...")
			a.dispatchTasks()
		case <-a.quitCh:
			return
		}
	}
}

func (a *Auditor) Shutdown() {
	a.executionTicker.Stop()
	a.quitCh <- true
	close(a.quitCh)
	close(a.taskCh)
}

func (a *Auditor) dispatchTasks() {
	count := 0
	var task Task
	defer log.Debugf("%d tasks dispatched", count)
	for {
		select {
		case task = <-a.taskCh:
			go task.Do()
			count++
		default:
			return
		}
		if count >= a.conf.MaxInFlightTasks {
			return
		}
	}
}
