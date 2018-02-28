package util

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
)

// AwaitTermSignal waits for standard termination signals, then rungs the given
// function; it should be run as a separate goroutine.
func AwaitTermSignal(closeFn func()) {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// block main and wait for a signal
	sig := <-signals
	glog.Infof("Signal received: %v", sig)

	closeFn()
}
