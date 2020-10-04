package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durMap = map[string]time.Duration{
	`d`: 24 * time.Hour,
	`h`: time.Hour,
	`m`: time.Minute,
	`s`: time.Second,
}

func ParseDuration(str string) (time.Duration, error) {
	match := regexp.MustCompile(
		strings.Join([]string{
			`([+-])?`,
			`(\d+d)?`,
			`(\d+h)?`,
			`(\d+m)?`,
			`(\d+s)?`,
		}, ""),
	).FindStringSubmatch(str)

	ret := time.Duration(0)
	for _, x := range match[2:] {
		dur, err := parseDurOrEmpty(x)
		if err != nil {
			return 0, err
		}

		ret += dur
	}

	if match[1] == "-" {
		ret = -ret
	}

	return ret, nil
}

func parseDurOrEmpty(x string) (time.Duration, error) {
	if len(x) == 0 {
		return 0, nil
	}
	numStr, unit := x[:len(x)-1], x[len(x)-1:]

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, err
	}

	mul, ok := durMap[unit]
	if !ok {
		return 0, fmt.Errorf("bad unit: %s", unit)
	}

	return time.Duration(num) * mul, nil
}
