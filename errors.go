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

type MissingPackageError struct {
	Name    PackageName
	Context string
}

func (e *CycleError) Error() string {
	parts := make([]string, len(e.Path))

	for i, p := range e.Path {
		parts[i] = string(p)
	}

	return fmt.Sprintf("Dependency cycle askarlandi: %s",
		strings.Join(parts, " → "))
}

func (e *MissingPackageError) Error() string {
	if e.Context == "" {
		return fmt.Sprintf("package not found: %q", e.Name)
	}
	return fmt.Sprintf("package not found: %q (%s)", e.Name, e.Context)
}

// InvalidGraphError graphin butovluyu pozulduqda qaytarilan xetadir
type InvalidGraphError struct {
	Reason string
}

func (e *InvalidGraphError) Error() string {
	return fmt.Sprintf("The graph is invalid: %s", e.Reason)
}
