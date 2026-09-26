package main

import (
	"errors"
	"testing"
)

func TestNewGraph(t *testing.T) {
	g := NewGraph()

	if g == nil {
		t.Fatal("NewGraph() returned nil")
	}

	if g.packages == nil {
		t.Fatal("the packages map is nil")
	}

	if len(g.packages) != 0 {
		t.Errorf("new graph should be empty, there are %d packages", len(g.packages))
	}
}

func TestAddDependency(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Graph)
		from, to PackageName
		wantErr  bool
	}{
		{
			name:    "Her iki paket movcuddur",
			setup:   func(g *Graph) { g.AddPackage("A"); g.AddPackage("B") },
			from:    "A",
			to:      "B",
			wantErr: false,
		},
		{
			name:    "from movcud deyil",
			setup:   func(g *Graph) { g.AddPackage("B") },
			from:    "A",
			to:      "B",
			wantErr: true,
		},
		{
			name:    "to mövcud deyil",
			setup:   func(g *Graph) { g.AddPackage("A") },
			from:    "A",
			to:      "B",
			wantErr: true,
		},
		{
			name: "dublikat dependency",
			setup: func(g *Graph) {
				g.AddPackage("A")
				g.AddPackage("B")
				g.AddDependency("A", "B")
			},
			from:    "A",
			to:      "B",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGraph()

			tt.setup(g)

			err := g.AddDependency(tt.from, tt.to)

			if tt.wantErr {
				if err == nil {
					t.Errorf("xeta gozlenilirdi, alinmadi")
				}
			} else {
				if err != nil {
					t.Errorf("xeta gozlenilmirdi, alindi: %v", err)
				}
			}
		})
	}
}

/*
func newTestGraph(t *testing.T) *Graph {
    t.Helper() // bunu mütləq çağır!

    g := NewGraph()

    //   A → B, A → C, B → D, C → D
    for _, name := range []PackageName{"A", "B", "C", "D"} {
        if err := g.AddPackage(name); err != nil {
            t.Fatalf("AddPackage(%q) xəta: %v", name, err)
        }
    }

    edges := [][2]PackageName{
        {"A", "B"},
        {"A", "C"},
        {"B", "D"},
        {"C", "D"},
    }
    for _, e := range edges {
        if err := g.AddDependency(e[0], e[1]); err != nil {
            t.Fatalf("AddDependency(%q, %q) xəta: %v", e[0], e[1], err)
        }
    }

    return g
}
*/

func newTestGraph(t *testing.T) *Graph {
	t.Helper()

	g := NewGraph()

	for _, name := range []PackageName{"A", "B", "C", "D"} {
		if err := g.AddPackage(name); err != nil {
			t.Fatalf("AddPackage(%q) xəta: %v", name, err)
		}
	}

	edges := [][2]PackageName{
		{"A", "B"},
		{"A", "C"},
		{"B", "D"},
		{"C", "D"},
	}
	for _, e := range edges {
		if err := g.AddDependency(e[0], e[1]); err != nil {
			t.Fatalf("AddDependency(%q, %q) xəta: %v", e[0], e[1], err)
		}
	}

	return g
}

func assertEqualPackageSlices(t *testing.T, got, want []PackageName) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("uzunluq fərqli:\n  got:  %v (len=%d)\n  want: %v (len=%d)",
			got, len(got), want, len(want))
		return
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("mövqe %d fərqli:\n  got:  %v\n  want: %v",
				i, got, want)
			return
		}
	}
}

// assertNoError - xəta olmadığını yoxlayır.
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("gözlənilməyən xəta: %v", err)
	}
}

// assertError - xəta olduğunu yoxlayır.
func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("xəta gözlənilirdi, alınmadı")
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) *Graph
		target    PackageName
		wantOrder []PackageName
		wantErr   bool
	}{
		{
			name: "diamond dependency",
			setup: func(t *testing.T) *Graph {
				return newTestGraph(t)
			},
			target:    "A",
			wantOrder: []PackageName{"D", "B", "C", "A"},
			wantErr:   false,
		},
		{
			name: "tək paket",
			setup: func(t *testing.T) *Graph {
				g := NewGraph()
				g.AddPackage("A")
				return g
			},
			target:    "A",
			wantOrder: []PackageName{"A"},
			wantErr:   false,
		},
		{
			name: "mövcud olmayan paket",
			setup: func(t *testing.T) *Graph {
				return NewGraph()
			},
			target:    "X",
			wantOrder: nil,
			wantErr:   true,
		},
		{
			name: "uzun zəncir",
			setup: func(t *testing.T) *Graph {
				g := NewGraph()
				for _, n := range []PackageName{"A", "B", "C", "D", "E"} {
					g.AddPackage(n)
				}
				g.AddDependency("A", "B")
				g.AddDependency("B", "C")
				g.AddDependency("C", "D")
				g.AddDependency("D", "E")
				return g
			},
			target:    "A",
			wantOrder: []PackageName{"E", "D", "C", "B", "A"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setup(t)

			gotOrder, err := g.ResolveWithCycleDedection(tt.target)

			if tt.wantErr {
				assertError(t, err)
				return
			}
			assertNoError(t, err)
			assertEqualPackageSlices(t, gotOrder, tt.wantOrder)
		})
	}
}

func TestCycleDetection(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) *Graph
		target    PackageName
		wantCycle []PackageName 
	}{
		{
			name: "self-loop",
			setup: func(t *testing.T) *Graph {
				g := NewGraph()
				g.AddPackage("A")
				g.AddDependency("A", "A")
				return g
			},
			target:    "A",
			wantCycle: []PackageName{"A", "A"},
		},
		{
			name: "qısa cycle",
			setup: func(t *testing.T) *Graph {
				g := NewGraph()
				g.AddPackage("A")
				g.AddPackage("B")
				g.AddDependency("A", "B")
				g.AddDependency("B", "A")
				return g
			},
			target:    "A",
			wantCycle: []PackageName{"A", "B", "A"},
		},
		{
			name: "uzun cycle",
			setup: func(t *testing.T) *Graph {
				g := NewGraph()
				for _, n := range []PackageName{"A", "B", "C", "D"} {
					g.AddPackage(n)
				}
				g.AddDependency("A", "B")
				g.AddDependency("B", "C")
				g.AddDependency("C", "D")
				g.AddDependency("D", "A")
				return g
			},
			target:    "A",
			wantCycle: []PackageName{"A", "B", "C", "D", "A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setup(t)

			_, err := g.ResolveWithCycleDedection(tt.target)
			assertError(t, err)

			// `errors.As` ilə xəta tipini yoxlayırıq.
			var cycleErr *CycleError
			if !errors.As(err, &cycleErr) {
				t.Fatalf("gözlənilən CycleError, alındı: %T", err)
			}

			assertEqualPackageSlices(t, cycleErr.Path, tt.wantCycle)
		})
	}
}

func TestDiamondDependency(t *testing.T) {
	// A → B, A → C, B → D, C → D
	//
	// Bu, cycle DEYİL.
	// Bu, "diamond" adlanır - D-yə iki yoldan çatırıq.
	// Cycle detection-ın bu halı SƏHV aşkarlamaması vacibdir.
	g := newTestGraph(t)

	order, err := g.ResolveWithCycleDedection("A")
	assertNoError(t, err)

	// D yalnız BİR DƏFƏ olmalıdır.
	countD := 0
	for _, name := range order {
		if name == "D" {
			countD++
		}
	}
	if countD != 1 {
		t.Errorf("D %d dəfə görünür, gözlənilən 1", countD)
	}

	// Sıra etibarlı olmalıdır.
	if !g.isValidInstallOrder("A", order) {
		t.Errorf("etibarsiz install sirasi: %v", order)
	}
}

func TestMissingPackage(t *testing.T) {
	g := NewGraph()
	g.AddPackage("A")

	_, err := g.GetPackage("X")
	assertError(t, err)

	var missingErr *MissingPackageError
	if !errors.As(err, &missingErr) {
		t.Fatalf("gözlənilən MissingPackageError, alındı: %T", err)
	}

	if missingErr.Name != "X" {
		t.Errorf("Name = %q, gözlənilən %q", missingErr.Name, "X")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) *Graph
		wantErr bool
	}{
		{
			name: "etibarlı DAG",
			setup: func(t *testing.T) *Graph {
				return newTestGraph(t)
			},
			wantErr: false,
		},
		{
			name: "boş graph",
			setup: func(t *testing.T) *Graph {
				return NewGraph()
			},
			wantErr: false,
		},
		{
			name: "cycle başlanğıcdan çatılan",
			setup: func(t *testing.T) *Graph {
				g := NewGraph()
				g.AddPackage("A")
				g.AddPackage("B")
				g.AddDependency("A", "B")
				g.AddDependency("B", "A")
				return g
			},
			wantErr: true,
		},
		{
			name: "cycle başlanğıcdan KƏNAR",
			setup: func(t *testing.T) *Graph {
				g := NewGraph()
				g.AddPackage("A")
				g.AddPackage("X")
				g.AddPackage("Y")
				g.AddDependency("X", "Y")
				g.AddDependency("Y", "X")
				return g
			},
			wantErr: true, // fərq: Validate bütün graph-ı yoxlayır
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setup(t)
			err := g.Validate()

			if tt.wantErr && err == nil {
				t.Error("xəta gözlənilirdi, alınmadı")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("xəta gözlənilmirdi, alındı: %v", err)
			}
		})
	}
}
