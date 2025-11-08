package optics

import (
	"context"
	"maps"
)

// Lens focuses a single substructure of S of type A.
// It takes a context and may return an error.
type Lens[S, A any] struct {
	View   func(context.Context, S) (A, error)
	Update func(context.Context, S, A) (S, error)
}

// Over modifies the focused value with f.
func (l Lens[S, A]) Over(ctx context.Context, s S, f func(context.Context, A) (A, error)) (S, error) {
	a, err := l.View(ctx, s)
	if err != nil {
		return s, err
	}
	a, err = f(ctx, a)
	if err != nil {
		return s, err
	}
	return l.Update(ctx, s, a)
}

// ComposeLensLens composes two lenses.
func ComposeLensLens[S, A, B any](l1 Lens[S, A], l2 Lens[A, B]) Lens[S, B] {
	return Lens[S, B]{
		View: func(ctx context.Context, s S) (B, error) {
			a, err := l1.View(ctx, s)
			if err != nil {
				var b B
				return b, err
			}
			return l2.View(ctx, a)
		},
		Update: func(ctx context.Context, s S, b B) (S, error) {
			a, err := l1.View(ctx, s)
			if err != nil {
				return s, err
			}
			a, err = l2.Update(ctx, a, b)
			if err != nil {
				return s, err
			}
			return l1.Update(ctx, s, a)
		},
	}
}

// ComposeLensPrism composes a lens and prism.
func ComposeLensPrism[S, A, B any](l Lens[S, A], p Prism[A, B]) Prism[S, B] {
	return Prism[S, B]{
		Match: func(ctx context.Context, s S) (B, error) {
			a, err := l.View(ctx, s)
			if err != nil {
				var b B
				return b, err
			}
			return p.Match(ctx, a)
		},
		Build: func(ctx context.Context, b B) (S, error) {
			var s S
			a, err := p.Build(ctx, b)
			if err != nil {
				return s, err
			}
			return l.Update(ctx, s, a)
		},
	}
}

func Key[M ~map[K]V, K comparable, V any](key K) Lens[M, V] {
	return Lens[M, V]{
		View: func(ctx context.Context, m M) (V, error) {
			return m[key], nil
		},
		Update: func(ctx context.Context, m M, v V) (M, error) {
			r := maps.Clone(m)
			r[key] = v
			return r, nil
		},
	}
}
