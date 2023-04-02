package peers
import (
	"context"
	"sync"
)

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

func ring[T any](arr []T, i int) T {
	return arr[i%len(arr)]
}

func OrderedAsync[T, R any](ctx context.Context, in chan T, out chan R, jobs int, fn func(T) (R, error)) (err error) {
	fanin := make([]chan R, jobs)
	fanout := make([]chan T, jobs)

	ctx, cancel := context.WithCancel(ctx)

	var once sync.Once
	var wg sync.WaitGroup

	// 2
	for i := 0; i < jobs; i++ {
		fanout[i] = make(chan T, 2)
		fanin[i] = make(chan R, 2)
		// we can defer here becuase this is the overall "OrderedAsync" sync scope
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer close(fanin[i])
			for el := range fanout[i] {
				r, e := fn(el)
				if e != nil {
					once.Do(func() {
						err = e
					})
					cancel()
				}
				fanin[i] <- r
			}
		}(i)
	}

	// 3
	go func() {
		defer close(out)
		for i := 0; true; i++ {
			el, ok := <-ring(fanin, i)
			if !ok {
				return
			}
			out <- el
		}
	}()

	// 1
	func() {
		i := 0
		for el := range in {
			select {
			case ring(fanout, i) <- el:
				i++
			case <-ctx.Done():
				return
			}
		}
	}()

	for i := 0; i < jobs; i++ {
		close(fanout[i])
	}

	wg.Wait()

	if err == nil {
		err = ctx.Err()
	}

	return
}

func IntoChan[T any](ctx context.Context, slice []T) chan T {
	ch := make(chan T)
	go func() {
		for i, _ := range slice {
			select {
			case ch <- slice[i]:
			case <-ctx.Done():
			}
		}
		close(ch)
	}()
	return ch
}

func collect[T any](ch chan T, size int) []T {
	els := make([]T, 0, size)
	for e := range ch {
		els = append(els, e)
	}
	return els
}
