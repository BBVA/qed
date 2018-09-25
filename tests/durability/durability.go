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

/*

	References:
	- https://www.informit.com/articles/article.aspx?p=370047&seqNum=4
	- https://stackoverflow.com/questions/5902629/mmap-msync-and-linux-process-termination
	- https://www.realworldtech.com/forum/?threadid=113923&curpostid=114068

	This program test what happens to our storage engine when it dies without calling close.
	Badger uses mmap to write to the file system and the kernel is in charge to persist the mmap'ed data
	to the disk. That happens also when the program chases, so we simulate a crass and test
	if the data is there.

	The default badger behaviour is to call fsync() on each write which obviously degrades
	write performance severely. So to achieve our desired performance we need to disable it.

	The approach of using mmap'ed files and calling fsync sporadically instead of each insert
	seems to be the normal behaviour of popular databases. We need more references to
	back this claim, which is based more on intuition than facts.

*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
	bd "github.com/bbva/qed/storage/badger"
)

func main() {
	op := flag.String("o", "select", "add | get operation")

	flag.Parse()

	fmt.Println("Executing operation ", *op)
	switch *op {
	case "add":
		add()
	case "get":
		deleteFile("/var/tmp/dur.db/LOCK")
		get()
	default:
		log.Error("Select add or get operation with -o option")
	}
}

func add() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("_add panic recovered", r)
		}
	}()

	_add()

}

func get() {
	store, closeF := openBadgerStore()
	defer closeF()

	key := []byte("Key")
	value := []byte("Value")

	stored, err := store.Get(0, key)
	if err != nil {
		log.Error(err)
	}
	if bytes.Compare(stored.Value, value) != 0 {
		log.Error("The stored key does not match the original: expected %d, actual %d", value, stored)
	}

}

func openBadgerStore() (*bd.BadgerStore, func()) {
	store, err := bd.NewBadgerStore("/var/tmp/badger_store_test.db")
	if err != nil {
		log.Errorf("Error opening badger store: %v", err)
		os.Exit(1)
	}
	return store, func() {
		store.Close()
		deleteFile("/var/tmp/badger_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}

func _add() {
	store, closeF := openBadgerStore()
	defer closeF()

	prefix := byte(0)
	key := []byte("Key")
	value := []byte("Value")

	// err := store.Mutate()
	err := store.Mutate([]*storage.Mutation{
		{prefix, value, key},
	})
	if err != nil {
		log.Error(err)
	}

	panic("we need to die")
}
