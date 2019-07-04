package cmd

import (
	"fmt"
	"io"
	"os/exec"
)

func formatInfo(w io.Writer, title string, body func()) {
	fmt.Fprintf(w, "#### %s\n\n```\n", title)
	body()
	fmt.Fprintf(w, "```\n")
}

func debugInfo(w io.Writer) {
	qedVersion := Ctx.Value(k("version"))
	// Get build version
	formatInfo(w, "Build Info", func() {
		fmt.Fprintf(w, "QED version %s, built in $GOPATH mode\n", qedVersion)
	})

	// Get GO environment info
	formatInfo(w, "Go Info", func() {
		gov := exec.Command("go", "version")
		env := exec.Command("go", "env")
		gov.Stdout = w
		gov.Run()
		env.Stdout = w
		env.Run()
		fmt.Fprint(w)
	})
}
