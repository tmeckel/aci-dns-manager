package internal

import (
	"os"
	"strconv"
	"strings"
)

func getenvTrim(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func GetenvInt(key string, def int) (int, error) {
	val := getenvTrim(key)
	if val == "" {
		return def, nil
	}
	ret, err := strconv.Atoi(val)
	if err != nil {
		return def, err
	}
	return ret, err
}

func GetenvInt64(key string, def int64) (int64, error) {
	val := getenvTrim(key)
	if val == "" {
		return def, nil
	}
	ret, err := strconv.ParseInt(val, 0, 32)
	if err != nil {
		return def, err
	}
	return ret, err
}

func GetenvStr(key string, def string) string {
	val := getenvTrim(key)
	if val == "" {
		return def
	}
	return val
}

func GetenvBool(key string, def bool) (bool, error) {
	val := getenvTrim(key)
	if val == "" {
		return def, nil
	}
	ret, err := strconv.ParseBool(val)
	if err != nil {
		return def, err
	}
	return ret, err
}
