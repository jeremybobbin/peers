package peers

import (
	"sort"
	"testing"
)

var keys = []string{
	"a",
	"ab",
	"abc",
	"bce",
	"ca",
	"d",
	"e",
}

var families = [][]string{
	{
		"a",
		"ab",
		"abc",
		"ca",
	},
	{
		"ab",
		"abc",
		"bce",
		"ca",
	},
	{
		"bce",
		"e",
	},
}

func derive(s1 string, s2 string) (byte, byte) {
	// derive function is first character in common
	// "abc" & "cda" = a, c
	b1 := byte(0)
	b2 := byte(0)
	for i := range s1 {
		for j := range s2 {
			if s1[i] == s2[j] {
				b1 = s1[i]
			}
		}
	}

	for i := range s2 {
		for j := range s1 {
			if s2[i] == s1[j] {
				b2 = s1[j]
			}
		}
	}
	return b1, b2
}

func prune(b byte) bool {
	// derive function is first character in common
	// "abc" & "cda" = a, c
	return b != byte(0)
}


func TestConnect(t *testing.T) {
	nodes := Connect[string, byte](keys, derive)
	m := AsMap(nodes)
	for _, fam := range families {
		if len(fam) < 1 {
			continue
		}
		n1 := m[fam[0]]
		for _, mem := range fam[1:] {
			n2 := m[mem]
			if !IsRelated(n1, n2) {
				t.Errorf(`%s not related to %s`, fam[0], mem)
			}
		}
	}

}

func TestFamilies(t *testing.T) {
	nodes := Connect[string, byte](keys, derive)
	Prune(nodes, prune)
	got := Families(nodes)
	for i := range families {
		sort.Strings(families[i])
	}
	for i := range got {
		sort.Slice(got[i], func(j, k int) bool {
			return got[i][j].Key < got[i][k].Key
		})
	}

	// sort by first family member
	// TODO: fully deterministic sort
	sort.Slice(families, func(i, j int) bool {
		return families[i][0] < families[j][0]
	})
	sort.Slice(got, func(i, j int) bool {
		return got[i][0].Key < got[j][0].Key
	})
	for i := 0; i < len(got) && i < len(families); i++ {
		for j := 0; j < len(got[i]) && j < len(families[i]); j++  {
			if got[i][j].Key != families[i][j] {
				t.Errorf(`got %s - expected %s`, got[i][j].Key, families[i][j])
			}
		}
	}
}

func TestCoagulate(t *testing.T) {

	var coagulated = [][]string{
		{
			"a",
			"ab",
			"abc",
			"ca",
		},
		{
			"bce",
			"e",
		},
	}

	nodes := Connect[string, byte](keys, derive)
	Prune(nodes, prune)
	Coagulate(Families(nodes))
	got := Families(nodes)

	for i := range coagulated {
		sort.Strings(coagulated[i])
	}
	for i := range got {
		sort.Slice(got[i], func(j, k int) bool {
			return got[i][j].Key < got[i][k].Key
		})
	}

	// sort by first family member
	// TODO: fully deterministic sort
	sort.Slice(coagulated, func(i, j int) bool {
		return coagulated[i][0] < coagulated[j][0]
	})
	sort.Slice(got, func(i, j int) bool {
		return got[i][0].Key < got[j][0].Key
	})
	for i := 0; i < len(got) && i < len(coagulated); i++ {
		for j := 0; j < len(got[i]) && j < len(coagulated[i]); j++  {
			if got[i][j].Key != coagulated[i][j] {
				t.Errorf(`got %s - expected %s`, got[i][j].Key, coagulated[i][j])
			}
		}
	}
}
