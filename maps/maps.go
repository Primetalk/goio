package maps

// Keys returns keys of the map
func Keys[K comparable, V any](m map[K]V) (keys []K) {
	for k := range m {
		keys = append(keys, k)
	}
	return
}

// Merge combines two maps.
// Function `combine` is invoked when the same key is available in both maps.
func Merge[K comparable, V any](m1 map[K]V, m2 map[K]V, combine func(V, V) V) (m map[K]V) {
	m = make(map[K]V)
	for k, v1 := range m1 {
		m[k] = v1
	}
	for k, v2 := range m2 {
		v1, ok := m[k]
		v := v2
		if ok {
			v = combine(v1, v2)
		}
		m[k] = v
	}
	return
}
