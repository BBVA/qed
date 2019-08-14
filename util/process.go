package util

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func AwaitTermSignal(closeFn func() error) {

	signals := make(chan os.Signal, 1)
	// sigint: Ctrl-C, sigterm: kill command
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// block main and wait for a signal
	sig := <-signals
	fmt.Printf("Signal received: %v\n", sig)

	closeFn()

}
