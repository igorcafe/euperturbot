package env

import (
	"fmt"
	"os"
)

func Must(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("invalid env variable %v with value '%v'", key, val))
	}
	return val
}
