package utils

import "os"

func GetEnvFromArgs(fallback ...string) string {
	defaultValue := ""

	if len(fallback) > 0 {
		defaultValue = fallback[0]
	}

	// os.Args[0] is program name, so if we have less than 2 args, it means no env is provided
	// and we should return the default value
	if len(os.Args) < 2 {
		return defaultValue
	}

	// IMPORTANT: Please note that the first argument we provide to the program should be the env name
	env := os.Args[1]

	if env != "" {
		return env
	}

	return defaultValue
}
