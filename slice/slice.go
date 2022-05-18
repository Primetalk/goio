package slice

func Map[A any, B any](as []A, f func(A)B)(bs []B) {
	bs = make([]B, len(as))
	for _, a := range as {
		bs = append(bs, f(a))
	}
	return
}

func FlatMap[A any, B any](as []A, f func(A)[]B)(bs []B) {
	bs = make([]B, len(as))
	for _, a := range as {
		bs = append(bs, f(a)...)
	}
	return
}

func FoldLeft[A any, B any](as []A, zero B, f func(B, A)B) (res B) {
	res = zero
	for _, a := range as {
		res = f(res, a)
	}
	return
}
