package main

import (
	"strings"
)

func parseDestination(args []string) (user, host string) {
	argsWithParams := map[string]bool{
		"-b": true, "-c": true, "-D": true, "-E": true, "-e": true,
		"-F": true, "-I": true, "-i": true, "-L": true, "-l": true,
		"-m": true, "-O": true, "-o": true, "-p": true, "-R": true,
		"-S": true, "-W": true, "-w": true,
	}

	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if strings.HasPrefix(arg, "-") {
			flag := arg
			if len(arg) > 2 && !strings.Contains(arg, "=") {
				flag = "-" + string(arg[len(arg)-1])
			}

			if argsWithParams[flag] {
				skipNext = true
			}
			continue
		}

		parts := strings.SplitN(arg, "@", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
		return "", parts[0] // No user specified, just host
	}
	return "", ""
}
