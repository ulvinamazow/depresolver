package main

import (
	"fmt"
	"strings"
)

type InstallPlan struct {
	// istifadecinin sorusdugu paket
	Root PackageName
	// paketlerin qurasdirilma sirasi
	Order []PackageName
	// orderin uzunlugu
	Total int
}

// BuilsInstallPlan verilmis paket ucun tam install plani qurur
func (graph *Graph) BuildInstallPlan(name PackageName) (*InstallPlan, error) {
	order, err := graph.ResolveWithCycleDedection(name)
	if err != nil {
		return nil, err
	}

	return &InstallPlan{
		Root:  name,
		Order: order,
		Total: len(order),
	}, nil

}
func (plan *InstallPlan) String() string {
	var sb strings.Builder
	// Basliq
	fmt.Fprintf(&sb, "Installation plan for %q:\n", plan.Root)

	for i, name := range plan.Order {
		fmt.Fprintf(&sb, "  %2d. %s\n", i+1, name)
	}

	fmt.Fprintf(&sb, "\nTotal: %d paket\n", plan.Total)

	return sb.String()

}

func (graph *Graph) CanInstall(name PackageName) bool {
	_, err := graph.ResolveWithCycleDedection(name)
	return err == nil
}
