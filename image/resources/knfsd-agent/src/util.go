package main

import (
	"errors"
	"strings"
)

// lastAfterDelimiter returns the string after the last occurrence of a delimiter
func lastAfterDelimiter(path string, seperator string) (string, error) {

	split := strings.Split(path, seperator)

	if len(split) == 0 {
		return "", errors.New("slice length was 0")
	}

	return split[len(split)-1], nil

}
