package env

import (
	"fmt"
	"os"
	"strconv"
)

func Must(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("invalid env variable %v with value '%v'", key, val))
	}
	return val
}

func MustInt64(key string) int64 {
	val := Must(key)
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to convert '%s' to int64", val))
	}
	return i
}
