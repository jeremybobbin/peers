package peers

func indiciesToElements[T any](of []T, in []int) []T {
	//fmt.Println("indiciesToElementer", of)
	out := make([]T, len(in))
	for i := range in {
		out[i] = of[in[i]]
	}
	return out
}


func toMap[T comparable](s []T) map[T]struct{} {
	m := make(map[T]struct{}, len(s))
	for i := range s {
		m[s[i]] = struct{}{}
	}
	return m
}


func subsetOfAny[T comparable](a [][]T, b []T) bool {
	for i := range a {
		if subset(a[i], b) {
			return true
		}
	}
	return false
}

func subset[T comparable](a, b []T) bool {
	am := toMap(a)
	for i := range b {
		if _, ok := am[b[i]]; !ok {
			return false
		}
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
