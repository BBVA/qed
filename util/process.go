package util

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bbva/qed/log"
)

func AwaitTermSignal(closeFn func() error) {

	signals := make(chan os.Signal, 1)
	// sigint: Ctrl-C, sigterm: kill command
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// block main and wait for a signal
	sig := <-signals
	log.Infof("Signal received: %v", sig)

	closeFn()

}
