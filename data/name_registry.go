package data

import "strings"

var humanNames = map[string]string{}

// RegisterHumanName associates an address with a given human-friendly name.
func RegisterHumanName(addr string, humanName string) {
	humanNames[strings.ToLower(addr)] = humanName
}
