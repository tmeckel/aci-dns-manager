package internal

import (
	"fmt"
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
		return def, fmt.Errorf("failed to convert value %w", err)
	}

	return ret, nil
}

func GetenvIntRange(key string, def, min, max int) (int, error) {
	ret, err := GetenvInt(key, def)
	if err != nil {
		return def, fmt.Errorf("failed to convert value %w", err)
	}
	if ret < min || ret > max {
		ret = def
	}

	return ret, nil
}

func GetenvInt64(key string, def int64) (int64, error) {
	val := getenvTrim(key)
	if val == "" {
		return def, nil
	}
	ret, err := strconv.ParseInt(val, 0, 32)
	if err != nil {
		return def, fmt.Errorf("failed to convert value %w", err)
	}

	return ret, nil
}

func GetenvInt64Range(key string, def, min, max int64) (int64, error) {
	ret, err := GetenvInt64(key, def)
	if err != nil {
		return def, fmt.Errorf("failed to convert value %w", err)
	}
	if ret < min || ret > max {
		ret = def
	}

	return ret, nil
}

func GetenvStr(key, def string) string {
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
		return def, fmt.Errorf("failed to convert value %w", err)
	}

	return ret, nil
}
