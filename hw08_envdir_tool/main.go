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
	}

	_ = RunCmd(args[2:], env)
}
