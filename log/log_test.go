package log

import (
	"os"
	"os/exec"
	"testing"
)

func TestLog(t *testing.T) {
	SetLogger("TestDebug", DEBUG)

	Debug("print driven development")
	Info("hello")

}

func Crasher() {
	Error("killed")
}

func Crasherf() {
	Errorf("killed in the name %s", "off")
}

func TestErrorDoingOsExit(t *testing.T) {

	if os.Getenv("BE_CRASHER") == "1" {
		Crasher()
		return
	}

	if os.Getenv("BE_CRASHER") == "2" {
		Crasherf()
		return
	}

	// Testing log.Error that runs os.Exit(1) succesfully
	cmd := exec.Command(os.Args[0], "-test.run=TestErrorDoingOsExit")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		// pass
	} else {
		t.Fatalf("log.Error ran with err %v, want exit status 1", err)
	}

	// Testing log.ErrorF that runs os.Exit(1) succesfully
	cmd2 := exec.Command(os.Args[0], "-test.run=TestErrorDoingOsExit")
	cmd2.Env = append(os.Environ(), "BE_CRASHER=2")

	err2 := cmd2.Run()
	if e, ok := err2.(*exec.ExitError); ok && !e.Success() {
		// pass
	} else {
		t.Fatalf("log.Errorf ran with err %v, want exit status 1", err)
	}

}
