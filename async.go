package peers

import (
	"context"
	"sync"
)

// determins what their relationship is
type DeriveAsync[K comparable, V any] func(K, K) (V, V, error)

func ConnectAsync[K comparable, V any](ctx context.Context, jobs int, src []K, derive DeriveAsync[K, V]) ([]*Node[K, V], error) {
	ctx, cancel := context.WithCancel(ctx)
	dst := make([]*Node[K, V], len(src))
	in := make(chan [2]int)
	out := make(chan [2]V)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer close(in)
		defer wg.Done()
		i := 0
		j := 1
		for i < len(src)-1 {
			if i == 0 {
				if j == 1 {
					dst[i] = &Node[K, V]{
						Key: src[i],
						Peers: make(PeerMap[K, V]),
					}
				}
				dst[j] = &Node[K, V]{
					Key: src[j],
					Peers: make(PeerMap[K, V]),
				}
			}

			select {
			case in <- [2]int{ i, j }:
			case <-ctx.Done():
				return
			}

			j++
			if j == len(src) {
				i++
				j = i+1
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		i := 0
		j := 1
		for v := range out {
			a := v[0]
			b := v[1]
			dst[i].Peers[dst[j]] = a
			dst[j].Peers[dst[i]] = b

			j++
			if j == len(src) {
				i++
				j = i+1
			}
		}
	}()

	err := OrderedAsync(ctx, in, out, jobs, func(indices [2]int) ([2]V, error) {
		i := indices[0]
		j := indices[1]
		v1, v2, err := derive(src[i], src[j])
		return [2]V{ v1, v2 }, err

	})

	if err != nil {
		cancel()
	}

	wg.Wait()

	return dst, err
}

type match[V any] struct {
	i, j int
	v1, v2 V
}

// change V to another type, only running derive on the already-connected nodes 
func MapAsync[K comparable, V1, V2 any](ctx context.Context, jobs int, src []*Node[K, V1], derive DeriveAsync[K, V2]) ([]*Node[K, V2], error) {
	ctx, cancel := context.WithCancel(ctx)
	dst := make([]*Node[K, V2], len(src))
	in := make(chan [2]int)
	out := make(chan match[V2])

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer close(in)
		defer wg.Done()
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
				select {
				case in <- [2]int{ i, j }:
				case <-ctx.Done():
					return
				}
			}

			j++
			if j == len(src) {
				i++
				j = i+1
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for m := range out {
			dst[m.i].Peers[dst[m.j]] = m.v1
			dst[m.j].Peers[dst[m.i]] = m.v2
		}
	}()

	err := OrderedAsync(ctx, in, out, jobs, func(indices [2]int) (match[V2], error) {
		i := indices[0]
		j := indices[1]
		v1, v2, err := derive(src[i].Key, src[j].Key)
		return match[V2]{
			i: i,
			j: j,
			v1: v1,
			v2: v2,
		}, err

	})

	if err != nil {
		cancel()
	}

	wg.Wait()

	return dst, err
}
