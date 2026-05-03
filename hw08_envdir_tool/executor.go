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
	//#gosec G702
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
	allEnvVars := make(map[string]string, len(baseEnv))

	for _, baseEnvStr := range baseEnv {
		envParts := strings.SplitN(baseEnvStr, "=", 2)
		if len(envParts) != 2 {
			continue
		}
		varName := envParts[0]
		varValue := envParts[1]
		allEnvVars[varName] = varValue
	}

	for varName, varValue := range env {
		if varValue.NeedRemove {
			delete(allEnvVars, varName)
		} else {
			allEnvVars[varName] = varValue.Value
		}
	}

	resultEnv := make([]string, 0, len(allEnvVars))

	for varName, varValue := range allEnvVars {
		resultEnv = append(resultEnv, varName+"="+varValue)
	}

	return resultEnv
}
