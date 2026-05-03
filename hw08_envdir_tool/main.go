package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args

	env, err := ReadDir(args[1])
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	returnCode := RunCmd(args[2:], env)

	os.Exit(returnCode)
}
