package fun

import "sync"

// Memoize returns a function that will remember the original function in a map.
// It's thread safe, however, not super performant.
func Memoize[A comparable, B any](f func(a A) B) func(A) B {
	m := make(map[A]B)
	mu := sync.Mutex{}
	return func(a A) B {
		mu.Lock()
		defer mu.Unlock()
		res, ok := m[a]
		if ok {
			return res
		} else {
			res = f(a)
			m[a] = res
			return res
		}
	}
}
