package optics

import (
	"maps"
	"slices"
)

type Lens[A, B any] struct {
	Get func(A) (B, bool)
	Set func(A, B) (A, bool)
}

func Slice[S []T, T any](i int) Lens[S, T] {
	return Lens[S, T]{
		Get: func(s S) (T, bool) {
			if i < 0 || i >= len(s) {
				var zero T
				return zero, false
			}
			return s[i], true
		},
		Set: func(s S, t T) (S, bool) {
			if i < 0 || i >= len(s) {
				return s, false
			}
			c := slices.Clone(s)
			c[i] = t
			return c, true
		},
	}
}

func Map[M map[K]V, K comparable, V any](k K) Lens[M, V] {
	return Lens[M, V]{
		Get: func(m M) (V, bool) {
			v, ok := m[k]
			return v, ok
		},
		Set: func(m M, v V) (M, bool) {
			c := maps.Clone(m)
			c[k] = v
			return c, true
		},
	}
}

func Modify[A, B any](a A, l Lens[A, B], f func(B) B) (A, bool) {
	v, ok := l.Get(a)
	if !ok {
		return a, false
	}
	return l.Set(a, f(v))
}

func Compose[A, B, C any](l1 Lens[A, B], l2 Lens[B, C]) Lens[A, C] {
	return Lens[A, C]{
		Get: func(a A) (C, bool) {
			b, ok := l1.Get(a)
			if !ok {
				var zero C
				return zero, false
			}
			return l2.Get(b)
		},
		Set: func(a A, c C) (A, bool) {
			b, ok := l1.Get(a)
			if !ok {
				return a, false
			}
			b, ok = l2.Set(b, c)
			if !ok {
				return a, false
			}
			return l1.Set(a, b)
		},
	}
}
