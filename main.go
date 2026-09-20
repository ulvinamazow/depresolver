package main

import (
	"errors"
	"fmt"
)

func main() {
	graph := NewGraph()

	graph.AddPackage("A")
	graph.AddPackage("B")
	graph.AddPackage("C")

	graph.AddDependency("A", "B")
	graph.AddDependency("B", "C")
	graph.AddDependency("C", "A")

	order, err := graph.ResolveWithCycleDedection("A")
	if err != nil {
		var cycleErr *CycleError

		if errors.As(err, &cycleErr) {
			fmt.Println("Cycle found!")
			fmt.Println("Path: ", cycleErr.Path)
		} else {
			fmt.Println("Other error: ", err)
		}
		return
	}

	fmt.Println("Install list: ", order)
}
