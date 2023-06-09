package peers

import (
	"sort"
)

// K is the ID, V is the relationship(weight) that K1 & K2 share
type PeerMap[K comparable, V any] map[*Node[K, V]]V
type Node[K comparable, V any] struct {
	Key K
	Peers PeerMap[K, V]
}

// determins what their relationship is
type Derive[K comparable, V any] func(K, K) (V, V)

// Prune will disconnect nodes if they don't "belong"
// belong is a predicate that takes the weight of any 2 nodes
func Prune[K comparable, V any](nodes []*Node[K, V], belongs func(*Node[K, V], *Node[K, V]) bool) {
	for i := 0; i < len(nodes); i++ {
		for j := i+1; j < len(nodes); j++ {
			if _, ok := nodes[i].Peers[nodes[j]]; ok && !belongs(nodes[i], nodes[j]) {
				delete(nodes[i].Peers, nodes[j])
			}
			if _, ok := nodes[j].Peers[nodes[i]]; ok && !belongs(nodes[j], nodes[i]) {
				delete(nodes[j].Peers, nodes[i])
			}
		}
	}
}


// Prune will disconnect nodes if they don't "belong"
// belong is a predicate that takes the weight of any 2 nodes
func PruneByWeight[K comparable, V any](nodes []*Node[K, V], belongs func(V) bool) {
	Prune(nodes, func(a *Node[K, V], b *Node[K, V]) bool {
		return belongs(a.Peers[b])
	})
}

// Prune will disconnect nodes if they don't "belong"
// belong is a predicate that takes the weight of any 2 nodes
func PruneByKey[K comparable, V any](nodes []*Node[K, V], belongs func(K, K) bool) {
	Prune(nodes, func(a *Node[K, V], b *Node[K, V]) bool {
		return belongs(a.Key, b.Key)
	})
}


// takes a set of keys & a derive function(that determines the weight of the relationship),
// and returns a set of interconnected nodes(a weighted graph)
func Connect[K comparable, V any](src []K, derive Derive[K, V]) []*Node[K, V] {
	dst := make([]*Node[K, V], len(src))
	i := 0
	j := 1
	for i < len(src)-1 {
		if i == 0 {
			if j == 1 {
				dst[i] = &Node[K, V]{
					Key: src[0],
					Peers: make(PeerMap[K, V]),
				}
			}
			dst[j] = &Node[K, V]{
				Key: src[j],
				Peers: make(PeerMap[K, V]),
			}
		}
		dst[i].Peers[dst[j]], dst[j].Peers[dst[i]] = derive(src[i], src[j])

		j++
		if j == len(src) {
			i++
			j = i+1
		}
	}
	return dst
}

// change V to another type, only running derive on the already-connected nodes 
func Map[K comparable, V1, V2 any](src []*Node[K, V1], derive Derive[K, V2]) []*Node[K, V2] {
	dst := make([]*Node[K, V2], len(src))

	i := 0
	j := 1
	for i < len(src)-1 {
		if i == 0 {
			if j == 1 {
				dst[i] = &Node[K, V2]{
					Key: src[0].Key,
					Peers: make(PeerMap[K, V2]),
				}
			}
			dst[j] = &Node[K, V2]{
				Key: src[j].Key,
				Peers: make(PeerMap[K, V2]),
			}
		}
		_, ok1 := src[i].Peers[src[j]]
		_, ok2 := src[j].Peers[src[i]]
		if ok1 && ok2 {
			dst[i].Peers[dst[j]], dst[j].Peers[dst[i]] = derive(src[i].Key, src[j].Key)
		}

		j++
		if j == len(src) {
			i++
			j = i+1
		}
	}
	return dst
}


// Limit a node to N Peers, sorted by less
func LimitPeers[K comparable, V any](node *Node[K, V], limit int, less func(*Node[K, V], *Node[K, V]) bool) {
	peers := make([]*Node[K, V], len(node.Peers))
	i := 0
	for peer := range node.Peers {
		peers[i] = peer
		i++
	}
	sort.Slice(peers, func(i, j int) bool {
		return less(peers[i], peers[j])
	})
	for i := limit; i < len(peers); i++ {
		delete(node.Peers, peers[i])
		delete(peers[i].Peers, node)
	}
}

// Limit a node to N Peers based on relationship priority
func LimitPeersByWeight[K comparable, V any](node *Node[K, V], limit int, less func(V, V) bool) {
	LimitPeers(node, limit, func (n1, n2 *Node[K, V]) bool {
		return less(node.Peers[n1], node.Peers[n2])
	})
}

// if these nodes point at eachother
func IsRelated[K comparable, V any](n1, n2 *Node[K, V]) bool {
	_, ok1 := n1.Peers[n2]
	_, ok2 := n2.Peers[n1]
	if (!ok1 && ok2) || (ok1 && !ok2) {
		panic("nodes must have either a mutual relationship or no relationship")
	}
	return ok1 && ok2
}

// if everyone is interconnected, it's a family
func IsFamily[K comparable, V any](nodes []*Node[K, V]) bool {
	for i := 0; i < len(nodes); i++ {
		for j := i+1; j < len(nodes); j++ {
			if !IsRelated(nodes[i], nodes[j]) {
				return false
			}
		}
	}
	return true
}

// if everyone is interconnected, it's a family
func IsFamilyByIndex[K comparable, V any](indicies []int, nodes []*Node[K, V]) bool {
	for i := 0; i < len(indicies); i++ {
		for j := i+1; j < len(indicies); j++ {
			if !IsRelated(nodes[indicies[i]], nodes[indicies[j]]) {
				return false
			}
		}
	}
	return true
}

/*

Test at nodes at indicies like this:

0 1 2 3 4

0 1 2 3
0 1 2 4
0 1 3 4
0 2 3 4
1 2 3 4

0 1 2
0 1 3 a[i]++
0 1 4 a[i]++; i--
0 2 3 a[i]++; i++; a[i] = a[i-1]+1
0 2 4 a[i]++; i--
0 3 4 a[i]++; i++; a[i] = a[i-1]+1; i--; i--
1 2 3 i--; i--; a[i]++; i++; a[i] = a[i-1]+1; i++; a[i] = a[i-1]+1
1 2 4 a[i]++
1 3 4 i--; a[i-1]++
2 3 4

if 1 2 3 4 passed, then no need to test 1 2 3 or 2 3 4, since we're not returning the subsets


// returns the families the node is a part of
// - selects the largest families
// - families can intersect, meaning, this Families function
//   will return duplicated Node pointers
*/

func Families[K comparable, V any](nodes []*Node[K, V]) [][]*Node[K, V] {
	var familyIndicies [][]int

	for n := len(nodes); n >= 2; n-- {
		d := make([]int, n)
		var i int 
		for i = range d {
			d[i] = i
		}

		max := func(i int) int {
			return (len(nodes)-len(d))+i
		}

		for d[0] != max(0) {
			// max value for index
			i := len(d) - 1

			if d[i] != max(i) {
				d[i]++
			} else {
				o := i
				for i > 0 && d[i] == max(i) {
					i--
				}
				d[i]++
				for i < o {
					i++
					d[i] = d[i-1]+1
				}
			}
			if !subsetOfAny(familyIndicies, d) {
				if IsFamilyByIndex(d, nodes) {
					familyIndicies = append(familyIndicies, d)
					d = make([]int, n)
				}
			}
		}
	}
	var families [][]*Node[K, V]

	for i := range familyIndicies {
		families = append(families, indiciesToElements(nodes, familyIndicies[i]))
	}

	return families
}

func FamiliesFromGroups[K comparable, V any](groups [][]*Node[K, V]) [][]*Node[K, V] {
	var families [][]*Node[K, V]

	for i := range groups {
		sub := Families(groups[i])
		for j := range sub {
			if !subsetOfAny(families, sub[j]) {
				families = append(families, sub[j])
			}
		}
	}
	return families
}

// returns a node & all of its peers UNORDERED
func Group[K comparable, V any](node *Node[K, V]) []*Node[K, V] {
	group := []*Node[K, V]{ node }
	for peer := range node.Peers {
		group = append(group, peer)
	}
	return group
}

func Groups[K comparable, V any](nodes []*Node[K, V]) (groups [][]*Node[K, V]) {
	for i := range nodes {
		groups = append(groups, Group(nodes[i]))
	}
	return
}

func Fold[K comparable, V any](node *Node[K, V], init V, fn func(V, V) V) V {
	for _, val := range node.Peers {
		init = fn(init, val)
	}
	return init
}


/*
	cuts ties of nodes outside beyond direct family
	nodes is assumed to be returned by Families
	order matters
*/
func Coagulate[K comparable, V any](groups [][]*Node[K, V]) {
	seen := make(map[*Node[K, V]]struct{})
	Groups: for i := range groups {
		// direct family set
		family := make(map[*Node[K, V]]struct{}, len(groups[i]))
		for j := range groups[i] {
			family[groups[i][j]] = struct{}{}
			// if any of these have been seen already, invalidate the family
			if _, ok := seen[groups[i][j]]; ok {
				continue Groups
			}
		}
		for j := range groups[i] {
			// each peer may be family. check if they're in the family set
			for peer := range groups[i][j].Peers {
				_, belongs := family[peer]
				if !belongs {
					delete(peer.Peers, groups[i][j])
					delete(groups[i][j].Peers, peer)
				}
			}
			seen[groups[i][j]] = struct{}{}
		}
	}
}


func AsMap[K comparable, V any](nodes []*Node[K, V]) map[K]*Node[K, V] {
	m := make(map[K]*Node[K, V], len(nodes))
	for i := range nodes {
		m[nodes[i].Key] = nodes[i]
	}
	return m
}

func PeerWeights[K comparable, V any](node *Node[K, V]) map[K]V {
	m := make(map[K]V, len(node.Peers))
	for p, v := range node.Peers {
		m[p.Key] = v
	}
	return m
}
