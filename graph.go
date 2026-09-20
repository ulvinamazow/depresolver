package main

import (
	"fmt"
	"sort"
	"strings"
)

type Graph struct {
	packages map[PackageName]*Package
}

type dfsFrame struct {
	pkg     *Package
	visited bool
}

func NewGraph() *Graph {
	return &Graph{
		packages: map[PackageName]*Package{},
	}
}

func (graph *Graph) AddPackage(name PackageName) {
	// paket adini databaseda axtar, eyer yoxdursa yarat
	if _, exists := graph.packages[name]; exists {
		return
	}

	graph.packages[name] = NewPackage(name)
}

func (graph *Graph) AddDependency(from, to PackageName) error {
	// asililiq elave etmek
	fromPkg, exists := graph.packages[from]

	// eger paket yoxdursa error qaytar
	if !exists {
		return fmt.Errorf("Package not found: %q", from)
	}

	// eger asili olmagini istediyimiz paket yoxdursa error qaytar
	if _, exists := graph.packages[to]; !exists {
		return fmt.Errorf("Dependency package not found: %q", to)
	}
	// her sey normaldirsa yarat
	fromPkg.AddDependency(to)

	return nil
}

func (graph *Graph) HasPackage(name PackageName) bool {
	// paketin var olub olmadigini yoxlayiriq
	// varsa true, yoxdursa false
	_, exists := graph.packages[name]
	return exists
}

func (graph *Graph) GetPackage(name PackageName) (*Package, error) {
	// paketin var olub olmadigini yoxlayiriq
	// varsa *Package-ya yazib gonderirik
	pkg, exists := graph.packages[name]
	if !exists {
		return nil, fmt.Errorf("Package not found: %q", name)
	}
	return pkg, nil
}

func (graph *Graph) DependenciesOf(name PackageName) ([]PackageName, error) {
	// GetPackage ile yoxlayiriq
	pkg, err := graph.GetPackage(name)
	if err != nil {
		// yoxdursa nil
		return nil, err
	}

	// varsa string olaraq asililiqlarini qaytaririq
	return pkg.Dependencies, nil
}

func (graph *Graph) String() string {
	// bu funksiya mapi slice kimi yazib siralamaq ucundur
	var sb strings.Builder
	sb.WriteString("Graph:\n")

	// burda map slice olur
	names := make([]string, 0, len(graph.packages))

	// map-dan adlari cixarib slice-ya yaziriq
	//slice-ya yazmagimizin sebebi:
	// mapda paket adlari her isledikde ferqli sira ile cixa biler
	// ona gore sabit bir data tipine, sliceya yaziriq ki
	// testlerde xeta almayaq
	for name := range graph.packages {
		names = append(names, string(name))
	}
	// elifbaya gore sirlayiriq
	sort.Strings(names)

	for _, name := range names {
		pkg := graph.packages[PackageName(name)]
		fmt.Fprintf(&sb, " %s -> ", pkg.Name)

		if len(pkg.Dependencies) == 0 {
			sb.WriteString("[]\n")
			continue
		}
		fmt.Fprintf(&sb, "%v\n", pkg.Dependencies)
	}
	return sb.String()
}

func (graph *Graph) TraverseReachable(name PackageName) ([]PackageName, error) {
	start, err := graph.GetPackage(name)
	if err != nil {
		return nil, err
	}

	visited := make(map[PackageName]bool)

	var result []PackageName

	graph.traverseDFS(start, visited, &result)

	return result, nil
}

// Depth-First Search
// gedib sonra arxaya qayidir
func (graph *Graph) traverseDFS(pkg *Package, visited map[PackageName]bool, result *[]PackageName) {
	if visited[pkg.Name] {
		return
	}

	visited[pkg.Name] = true

	for _, dep := range pkg.Dependencies {

		depPkg, _ := graph.GetPackage(dep)

		graph.traverseDFS(depPkg, visited, result)
	}

	*result = append(*result, pkg.Name)
}

func (graph *Graph) TraverseReachableIterative(name PackageName) ([]PackageName, error) {
	start, err := graph.GetPackage(name)
	if err != nil {
		return nil, err
	}

	visited := make(map[PackageName]bool)
	var result []PackageName

	stack := []dfsFrame{{pkg: start, visited: false}}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if top.visited {
			result = append(result, top.pkg.Name)
			continue
		}

		if visited[top.pkg.Name] {
			continue
		}
		visited[top.pkg.Name] = true

		stack = append(stack, dfsFrame{pkg: top.pkg, visited: true})

		for i := len(top.pkg.Dependencies) - 1; i >= 0; i-- {
			dep := top.pkg.Dependencies[i]
			depPkg, _ := graph.GetPackage(dep)
			stack = append(stack, dfsFrame{pkg: depPkg, visited: false})
		}
	}
	return result, nil
}

func (graph *Graph) Resolve(name PackageName) ([]PackageName, error) {
	start, err := graph.GetPackage(name)
	if err != nil {
		return nil, err
	}

	visited := make(map[PackageName]bool)

	var order []PackageName

	graph.resolveDFS(start, visited, &order)

	return order, nil

}

func (graph *Graph) resolveDFS(pkg *Package, visited map[PackageName]bool, order *[]PackageName) {
	if visited[pkg.Name] {
		return
	}

	visited[pkg.Name] = true

	for _, dep := range pkg.Dependencies {
		depPkg, _ := graph.GetPackage(dep)
		graph.resolveDFS(depPkg, visited, order)
	}

	*order = append(*order, pkg.Name)

}

func (graph *Graph) isValidInstallOrder(name PackageName, order []PackageName) bool {
	// Hər paketin sıradakı mövqeyini qeyd edirik.
	position := make(map[PackageName]int, len(order))
	for i, p := range order {
		position[p] = i
	}

	// Hər paket üçün, onun bütün dependency-ləri sırada
	// DAHA ƏVVƏL olmalıdır.
	for _, pkg := range order {
		p, _ := graph.GetPackage(pkg)
		for _, dep := range p.Dependencies {
			if position[dep] >= position[pkg] {
				return false
			}
		}
	}
	return true
}

func (graph *Graph) ResolveWithCycleDedection(name PackageName) ([]PackageName, error) {
	start, err := graph.GetPackage(name)
	if err != nil {
		return nil, err
	}

	state := make(map[PackageName]color)
	parent := make(map[PackageName]PackageName)
	var order []PackageName

	err = graph.dedectCycleDFS(start, state, parent, &order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (graph *Graph) dedectCycleDFS(
	pkg *Package,
	state map[PackageName]color,
	parent map[PackageName]PackageName,
	order *[]PackageName,
) error {
	state[pkg.Name] = gray

	for _, dep := range pkg.Dependencies {
		depPkg, _ := graph.GetPackage(dep)
		depColor := state[depPkg.Name]

		switch depColor {
		case gray:
			return graph.buildCycleError(pkg.Name, depPkg.Name, parent)

		case white:
			parent[depPkg.Name] = pkg.Name

			if err := graph.dedectCycleDFS(depPkg, state, parent, order); err != nil {
				return err
			}
		case black:
			continue
		}
	}

	state[pkg.Name] = black

	*order = append(*order, pkg.Name)

	return nil
}

func (graph *Graph) buildCycleError(from, to PackageName, parent map[PackageName]PackageName) error {
	path := []PackageName{from}
	current := from

	for current != to {
		p, ok := parent[current]

		if !ok {
			break
		}
		path = append(path, p)
		current = p
	}

	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	path = append(path, path[0])

	return &CycleError{Path: path}
}
