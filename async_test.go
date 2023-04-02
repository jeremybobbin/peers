package peers

import (
	"context"
	"testing"
)

func TestAsyncConnect(t *testing.T) {
	nodes, _ := ConnectAsync[string, byte](context.TODO(), 64, keys, func(s1 string, s2 string) (byte, byte, error) {
		b1, b2 := derive(s1, s2)
		return b1, b2, nil
	})

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
