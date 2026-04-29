package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}
	//#nosec G204
	command := exec.CommandContext(context.Background(), cmd[0], cmd[1:]...)

	command.Env = joinEnvs(os.Environ(), env)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}

		return 1
	}

	if command.ProcessState == nil {
		return 1
	}

	return command.ProcessState.ExitCode()
}

func joinEnvs(baseEnv []string, env Environment) []string {
	resultEnv := make([]string, 0, len(baseEnv))

	for _, baseEnvStr := range baseEnv {
		envParts := strings.Split(baseEnvStr, "=")
		envName := envParts[0]

		if envElem, ok := env[envName]; ok && !envElem.NeedRemove {
			resultEnv = append(resultEnv, envName+"="+envElem.Value)
		} else {
			resultEnv = append(resultEnv, baseEnvStr)
		}
	}

	return resultEnv
}
