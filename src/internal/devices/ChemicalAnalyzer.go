package devices

import "math/rand"

// Lista completa dos elementos químicos
var ChemicalElementList = []string{
	"H", "He", 
	"Li", "Be", "B", "C", "N", "O", "F", "Ne", 
	"Na", "Mg", "Al", "Si", "P", "S", "Cl", "Ar",
	"K", "Ca", "Sc", "Ti", "V", "Cr", "Mn", "Fe", "Co", "Ni", "Cu", "Zn", "Ga", "Ge", "As", "Se", "Br", "Kr",
	"Rb", "Sr", "Y", "Zr", "Nb", "Mo", "Tc", "Ru", "Rh", "Pd", "Ag", "Cd", "In", "Sn", "Sb", "Te", "I", "Xe",
	"Cs", "Ba", "La", "Ce", "Pr", "Nd", "Pm", "Sm", "Eu", "Gd", "Tb", "Dy", "Ho", "Er", "Tm", "Yb", "Lu", "Hf", "Ta", "W", "Re", "Os", "Ir", "Pt", "Au", "Hg", "Tl", "Pb", "Bi", "Po", "At", "Rn",
	"Fr", "Ra", "Ac", "Th", "Pa", "U", "Np", "Pu", "Am", "Cm", "Bk", "Cf", "Es", "Fm", "Md", "No", "Lr", "Rf", "Db", "Sg", "Bh", "Hs", "Mt", "Ds", "Rg", "Cn", "Nh", "Fl", "Mc", "Lv", "Ts", "Og",
}




type ChemicalAnalyzer interface {
	Analyze() []Component
}

type Component struct {
	Name       string
	Percentage float32
}

type MockChemicalAnalyzer struct{}

func NewMockChemicalAnalyzer() *MockChemicalAnalyzer {
	return &MockChemicalAnalyzer{}
}

func (a *MockChemicalAnalyzer) Analyze() []Component {
	num := rand.Intn(5) + 2
	comps := make([]Component, num)
	total := float32(100.0)
	for i := 0; i < num; i++ {
		name := ChemicalElementList[rand.Intn(len(ChemicalElementList))]
		perc := rand.Float32() * (total / float32(num-i))
		if i == num-1 {
			perc = total
		}
		comps[i] = Component{Name: name, Percentage: perc}
		total -= perc
	}
	return comps
}
