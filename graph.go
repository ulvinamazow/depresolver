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

func (graph *Graph) AddPackage(name PackageName) error {
	// paket adini databaseda axtar, eyer yoxdursa yarat
	if name == "" {
		return &InvalidGraphError{Reason: "Package name cannot be empty"}
	}
	if _, exists := graph.packages[name]; exists {
		return nil
	}

	graph.packages[name] = NewPackage(name)
	return nil
}

func (graph *Graph) AddDependency(from, to PackageName) error {
	// asililiq elave etmek
	if _, exists := graph.packages[from]; !exists {
		return &MissingPackageError{
			Name:    from,
			Context: "AddDependecy: 'from' package",
		}
	}

	// eger asili olmagini istediyimiz paket yoxdursa error qaytar
	if _, exists := graph.packages[to]; !exists {
		return &MissingPackageError{
			Name:    to,
			Context: "AddDependency: 'to' (dependency) package",
		}
	}
	// her sey normaldirsa yarat
	graph.packages[from].AddDependency(to)

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
	// varsa *Package tipindeki deyisenleri
	// pkg-ye yazib gonderirik
	pkg, exists := graph.packages[name]
	if !exists {
		return nil, &MissingPackageError{Name: name}
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
	// paket asililiga yazilibsa start-a yazirig
	// yoxdursa err dondururuk
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
	// if bloku sadece yoxlanmis bir instance ucun axtarisi dayandirir
	// mes: A [B, C], B-nin yoxlandigini qeyd edirik
	// B ucun axtarisi dayandirib C-ye kecir
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

	err = graph.detectCycleDFS(start, state, parent, &order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (graph *Graph) Validate() error {
	for name, pkg := range graph.packages {
		for _, dep := range pkg.Dependencies {
			if _, exists := graph.packages[dep]; !exists {
				return &InvalidGraphError{
					Reason: fmt.Sprintf(
						"package %q, dependency %q is missing from graph", name, dep,
					),
				}
			}
		}
	}

	return graph.detectAnyCycle()
}

func (graph *Graph) detectAnyCycle() error {
	state := make(map[PackageName]color)
	parent := make(map[PackageName]PackageName)
	var dummy []PackageName

	for name := range graph.packages {
		if state[name] == black {
			continue
		}

		pkg := graph.packages[name]
		if err := graph.detectCycleDFS(pkg, state, parent, &dummy); err != nil {
			return err
		}
	}
	return nil
}

func (graph *Graph) detectCycleDFS(
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

			if err := graph.detectCycleDFS(depPkg, state, parent, order); err != nil {
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

func (graph *Graph) ResolveKhan(name PackageName) ([]PackageName, error) {
	// Baslangic paketin movcudlugunu yoxlayiriq
	if _, err := graph.GetPackage(name); err != nil {
		return nil, err
	}

	// `A → B` o deməkdir ki, A, B-dən asılıdır.
	//  Deməli B-nin in-degree-i 1 artır (A ondan asılıdır)
	inDegree := make(map[PackageName]int)
	reachable := make(map[PackageName]bool)

	var collect func(pkg *Package)

	collect = func(pkg *Package) {
		// paket mapda yeni yoxlanmislarda varsa
		// o paket ucun yoxlamani dayandiririq
		if reachable[pkg.Name] {
			return
		}
		// yoxdursa elave edirik, sonra yoxlayiriq
		reachable[pkg.Name] = true
		// paketin asililiqlarini yoxlayiriq
		for _, dep := range pkg.Dependencies {
			depPkg, _ := graph.GetPackage(dep)
			collect(depPkg)
		}
	}
	// paketin adinin olub oladigini yoxlayiriq
	start, _ := graph.GetPackage(name)
	// sonra funksiyani burda cagiririq
	collect(start)

	// reachabledeki adlari aliriq
	// inDegreede hemin adlarin value-sini 0 edirik
	for pkgName := range reachable {
		inDegree[pkgName] = 0
	}

	// reachabledeki adlari aliriq, pkgNameya yaziriq
	for pkgName := range reachable {
		// funksiya ile asililiqlara elave edilib edilmediyini yoxlayiriq
		pkg, _ := graph.GetPackage(pkgName)
		// onlarin adi asililiqda varsa, inDegree-de asililiq sayini 1 qaldiririq
		for _, dep := range pkg.Dependencies {
			inDegree[dep]++
		}
	}
	// queue, yeni novbe yaradiriq
	var queue []PackageName
	// reachabledeki adlari aliriq
	for pkgName := range reachable {
		// adin inDegreedeki deyeri 0-dirsa
		// demeli o hec bir paketden asili deyil, onu inDegree-den cixara bilerik
		// cixarib cixarib novbeye elave edirik
		if inDegree[pkgName] == 0 {
			// cixarib cixarib novbeye elave edirik
			queue = append(queue, pkgName)
		}
	}

	var result []PackageName

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		result = append(result, current)

		// Cari node-un dependency-lərinin in-degree-ini azaldırıq
		pkg, _ := graph.GetPackage(current)
		for _, dep := range pkg.Dependencies {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	// Əgər nəticə bütün reachable paketləri əhatə etmirsə,
	// deməli bəzi paketlər cycle-da ilişib qalıb və
	// heç vaxt in-degree = 0 olmayıb.

	if len(result) != len(reachable) {
		// Cycle var. Lakin Kahn's algorithm cycle-ın yolunu bilmir.
		// Ona görə CycleError-u DFS ilə quraq.
		return nil, graph.findAnyCycle(start)
	}

	// Kahn algoritmi bize topological sort verir
	// bize reverse topological sort lazimdir
	// ona gore tersine ceviririk
	reverse(result)
	return result, nil
}

func (graph *Graph) findAnyCycle(start *Package) error {
	state := make(map[PackageName]color)
	parent := make(map[PackageName]PackageName)
	var dummy []PackageName

	return graph.detectCycleDFS(start, state, parent, &dummy)
}

func reverse(s []PackageName) {
	// ResolveKhan funksiyansinin cavabini ters cevirmek ucun
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func (graph *Graph) Reverse() *Graph {
	rev := NewGraph()

	// Butun paketleri elave edirik
	for name := range graph.packages {
		rev.AddPackage(name)
	}

	// asililiqlari ters ceviririk
	for name, pkg := range graph.packages {
		for _, dep := range pkg.Dependencies {
			rev.AddDependency(dep, name)
		}
	}

	return rev
}

// UnusedPackages graphda olub hec bir yerde istifade
// edilemeyen paketleri tapir
// hansi ki bu paketler ne bir paketden asilidir
// nede hansisa paket ondan asilidir
func (graph *Graph) UnusedPackages() []PackageName {
	inDegree := make(map[PackageName]int, len(graph.packages))
	// burda butun adlari inDegree-nin icine qoyuruq
	// ve mapdaki int deyerini 0 edirem
	for name := range graph.packages {
		inDegree[name] = 0
	}
	// graphdaki paketleri pkg-ye yaziram
	for _, pkg := range graph.packages {
		// pkg-nin dependencies-ni dep-e yaziram
		for _, dep := range pkg.Dependencies {
			// dependency adini inDegree-ye qoyub deyerini 1 artiriram
			// cunki dependencies icinde varsa demeli
			// o nedense ve kimse ondan asilidir
			inDegree[dep]++
		}
	}

	var unused []PackageName
	// key value olaraq paketleri aliriq
	for name, pkg := range graph.packages {
		// inDegredeki ve dependenciesdeki uzunlugu 0-dirsa
		// onun adini unusedin icine yaziriq, mes: A
		if inDegree[name] == 0 && len(pkg.Dependencies) == 0 {
			unused = append(unused, name)
		}
	}

	sort.Slice(unused, func(i, j int) bool {
		return unused[i] < unused[j]
	})
	return unused
}

type GraphStats struct {
	Nodes        int           // paket sayi
	Edges        int           // edge sayi
	AvgOutDegree float64       // orta dependency sayi
	MaxOutDegree int           // en cox dependecy-e sahib paket
	MaxOutPkg    PackageName   //hansi paket en cox dependency-e sahibdir
	Roots        []PackageName // Baslangic paketler
	Leaves       []PackageName
}

// stats graph haqqinda statik melumat qaytarir

func (graph *Graph) Stats() *GraphStats {
	// paket sayini yazdiq
	stats := &GraphStats{
		Nodes: len(graph.packages),
	}
	// bos amma paket uzunlugu qeder uzunlugda map yaradiriq
	inDegree := make(map[PackageName]int, len(graph.packages))

	// paket adlarini inDegreye yazdim ve hamisini 0 etdim
	for name := range graph.packages {
		inDegree[name] = 0
	}

	for name, pkg := range graph.packages {
		// dependencyleri edgeye yazdim
		stats.Edges += len(pkg.Dependencies)

		// dependency-leri oz sayina gore artirdim
		// inDegree ni hesabladim
		for _, dep := range pkg.Dependencies {
			inDegree[dep]++
		}

		// MaxOutDegree ve MaxOutPkg nin teyini
		if len(pkg.Dependencies) > stats.MaxOutDegree {
			stats.MaxOutDegree = len(pkg.Dependencies)
			stats.MaxOutPkg = name
		}

		if len(pkg.Dependencies) == 0 {
			stats.Leaves = append(stats.Leaves, name)
		}
	}

	// root paketler (asililiq yoxdu)
	for name := range graph.packages {
		if inDegree[name] == 0 {
			stats.Roots = append(stats.Roots, name)
		}
	}

	// Orta out-degree
	if stats.Nodes > 0 {
		stats.AvgOutDegree = float64(stats.Edges) / float64(stats.Nodes)
	}

	sort.Slice(stats.Roots, func(i, j int) bool {
		return stats.Roots[i] < stats.Roots[j]
	})
	sort.Slice(stats.Leaves, func(i, j int) bool {
		return stats.Leaves[i] < stats.Leaves[j]
	})

	return stats
}

func (graphStats *GraphStats) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Graph Statistics: \n")
	fmt.Fprintf(&sb, "  Nodes:          %d\n", graphStats.Nodes)
	fmt.Fprintf(&sb, "  Edges:          %d\n", graphStats.Edges)
	fmt.Fprintf(&sb, "  Avg out-degree: %.2f\n", graphStats.AvgOutDegree)
	fmt.Fprintf(&sb, "  Max out-degree: %d (paket: %s)\n", graphStats.MaxOutDegree, graphStats.MaxOutPkg)
	fmt.Fprintf(&sb, "  Roots:          %v\n", graphStats.Roots)
	fmt.Fprintf(&sb, "  Leaves:         %v\n", graphStats.Leaves)
	return sb.String()
}

// grapin en uzun zencirinin uzunlugunu tapir
func (graph *Graph) MaxDepth() int {
	memo := make(map[PackageName]int)

	visiting := make(map[PackageName]bool)

	var maxSoFar int

	var depth func(pkg *Package) int
	depth = func(pkg *Package) int {
		// paket adi mapda varsa indexve true dondurur
		if d, ok := memo[pkg.Name]; ok {
			// varsa funksiya dayanir
			return d
		}

		if visiting[pkg.Name] {
			return -1
		}
		visiting[pkg.Name] = true

		maxChild := 0
		for _, dep := range pkg.Dependencies {
			depPkg, _ := graph.GetPackage(dep)
			d := depth(depPkg)
			if d == -1 {
				return -1
			}
			if d > maxChild {
				maxChild = d
			}

		}
		delete(visiting, pkg.Name)
		result := maxChild + 1
		memo[pkg.Name] = result

		if result > maxSoFar {
			maxSoFar = result
		}
		return result
	}

	for name := range graph.packages {
		if _, ok := memo[name]; ok {
			continue
		}

		pkg := graph.packages[name]
		if depth(pkg) == -1 {
			return -1
		}
	}
	return maxSoFar
}
