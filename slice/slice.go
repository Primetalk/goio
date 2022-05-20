package slice

func Map[A any, B any](as []A, f func(A)B)(bs []B) {
	bs = make([]B, 0, len(as))
	for _, a := range as {
		bs = append(bs, f(a))
	}
	return
}

func FlatMap[A any, B any](as []A, f func(A)[]B)(bs []B) {
	bs = make([]B, 0, len(as))
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

func Filter[A any](as []A, p func(a A) bool) (res []A){
	res = make([]A, 0, len(as))
	for _, a := range as {
		if p(a) {
			res = append(res, a)
		}
	}
	return
}

func Flatten[A any](ass [][]A)(aas[]A) {
	total := 0
	for _, as := range ass {
		total += len(as)
	}
	aas = make([]A, 0, total)
	for _, as := range ass {
		aas = append(aas, as...)
	}
	return
}

type Set[A comparable] map[A]struct{}

func ToSet[A comparable](as []A)(s Set[A]){
	s = make(map[A]struct{}, len(as))
	for _, a := range as {
		s[a] = struct{}{}
	}
	return
}
