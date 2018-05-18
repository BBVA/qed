package e2e

import (
	"flag"
	"os"
	"os/signal"
	"testing"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/server"
)

func MainTest(m *testing.M) {
	os.Exit(RunTests(m))
}

// RunTests runs the tests in a package while gracefully handling interrupts.
func RunTests(m *testing.M) int {
	log.SetLogger("client-test", "info")
	flag.Parse()
	if !testing.Short() {
		stopServer := setupServer()
		go func() {
			// Shut down tests when interrupted (for example CTRL+C).
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt)
			<-sig
			select {
			default:
				stopServer()
			}
		}()
	}
	return m.Run()
}

func setupServer() func() {
	path := "/var/tmp/balloonE2E"
	clearPath(path)
	server := server.NewServer(":8079", path, "my-awesome-api-key", uint64(50000), "badger", false, true)

	go (func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Info(err)
		}
	})()

	return func() {
		server.Close()
	}
}

func clearPath(path string) {
	os.RemoveAll(path)
	os.MkdirAll(path, os.FileMode(0755))
}
