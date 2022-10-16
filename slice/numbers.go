package slice

// Number is a generic number interface that covers all Go number types.
type Number interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64 |
		complex64 | complex128
}

// Sum sums numbers.
func Sum[N Number](ns []N) (sum N) {
	for _, n := range ns {
		sum = sum + n
	}
	return
}

// Range starts at `from` and progresses until `to` exclusive.
func Range(from, to int) (res []int) {
	res = make([]int, 0, to-from)
	for i := from; i < to; i++ {
		res = append(res, i)
	}
	return
}

// Nats return slice []int{1, 2, ..., n}
func Nats(n int) []int {
	return Range(1, n+1)
}
