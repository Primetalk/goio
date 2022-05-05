package io

type Pair[A any, B any] struct {
	A A
	B B
}

func NewPair[A any, B any](a A, b B) Pair[A, B] { return Pair[A, B]{A:a, B: b}}
