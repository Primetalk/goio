package slice

import "github.com/primetalk/goio/fun"

// Sum sums numbers.
func Sum[N fun.Number](ns []N) (sum N) {
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
