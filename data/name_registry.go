package data

import "strings"

var humanNames = map[string]string{}

// ListHumanNames lists known human-friendly names.
func ListHumanNames() []string {
	ret := []string{}
	for _, v := range humanNames {
		ret = append(ret, v)
	}
	return ret
}

// RegisterHumanName associates an address with a given human-friendly name.
func RegisterHumanName(addr string, humanName string) {
	humanNames[strings.ToLower(addr)] = humanName
}

// HumanName returns a human-friendly name for an address.
func HumanName(addr string) string {
	if h := humanNames[strings.ToLower(addr)]; h != "" {
		return h
	}
	return addr
}

// HasHumanName returns whether a human-friendly name for an address exists.
func HasHumanName(addr string) bool {
	_, ok := humanNames[strings.ToLower(addr)]
	return ok
}
