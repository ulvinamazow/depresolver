package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== Nümunə 1: Etibarli graph ===")
	graph := NewGraph()

	graph.AddPackage("A")
	graph.AddPackage("B")
	graph.AddPackage("C")
	graph.AddPackage("D")
	graph.AddPackage("X")

	graph.AddDependency("A", "B")
	graph.AddDependency("A", "C")
	graph.AddDependency("B", "D")
	graph.AddDependency("C", "D")

	fmt.Println(graph.Stats())
	fmt.Println()

	unused := graph.UnusedPackages()
	fmt.Println("Unused depth: ", unused)
	fmt.Println()

	fmt.Println("Max depth:", graph.MaxDepth())
	fmt.Println()

	rev := graph.Reverse()
	fmt.Println("Reverse graph: ")
	fmt.Println(rev)

	dependents, _ := rev.DependenciesOf("A")
	fmt.Println("Those dependent on A: ", dependents)
}
