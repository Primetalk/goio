package maps

// Keys returns keys of the map.
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

// MapKeys converts original keys to new keys.
func MapKeys[K1 comparable, V any, K2 comparable](m1 map[K1]V, f func(K1) K2, combine func(V, V) V) (m2 map[K2]V) {
	m2 = make(map[K2]V, len(m1))
	for k1, v := range m1 {
		k2 := f(k1)
		v2, ok := m2[k2]
		if ok {
			v = combine(v2, v)
		}
		m2[k2] = v
	}
	return
}

// MapValues converts values in the map using the provided function.
func MapValues[K comparable, A any, B any](m map[K]A, f func(A) B) (res map[K]B) {
	res = make(map[K]B, len(m))
	for k, a := range m {
		res[k] = f(a)
	}
	return
}
