package main

type PackageName string

type Package struct {
	Name         PackageName
	Dependencies []PackageName
}

func NewPackage(name PackageName) *Package {
	return &Package{
		Name:         name,
		Dependencies: []PackageName{},
	}
}

func (pack *Package) AddDependency(dependency PackageName) {
	for _, existing := range pack.Dependencies {
		if existing == dependency {
			return
		}
	}

	pack.Dependencies = append(pack.Dependencies, dependency)
}
