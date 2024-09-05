package utils

import (
	"strings"
)

func IfElse[T any](cond bool, A T, B T) T {
	if cond {
		return A
	}
	return B
}

func UpperFirst(name string) string {
	return strings.ToUpper(name[:1]) + name[1:]
}
