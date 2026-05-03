package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "invalid amount of args")
		os.Exit(1)
	}

	env, err := ReadDir(args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	returnCode := RunCmd(args[2:], env)

	os.Exit(returnCode)
}
