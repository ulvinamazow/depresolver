package main

import (
	"fmt"
	"strings"
)

type color int

const (
	white color = iota
	gray
	black
)

type CycleError struct {
	Path []PackageName
}

func (e *CycleError) Error() string {
	parts := make([]string, len(e.Path))

	for i, p := range e.Path {
		parts[i] = string(p)
	}

	return fmt.Sprintf("Dependency cycle askarlandi: %s",
		strings.Join(parts, " → "))
}
