package optics

import (
	"context"
	"errors"
)

// Traversal modifies 0..n focuses, returning the updated S and how many were hit.
// It takes a contextual value C and may return an error.
type Traversal[S, A any] struct {
	Modify func(context.Context, S, func(context.Context, A) (A, error)) (S, error)
}

// Over is a convenience alias for Modify.
func (t Traversal[S, A]) Over(ctx context.Context, s S, f func(context.Context, A) (A, error)) (S, error) {
	return t.Modify(ctx, s, f)
}

// ComposeTraversalLens composes a Traversal and Lens and returns a Traversal.
func ComposeTraversalLens[S, A, B any](t Traversal[S, A], l Lens[A, B]) Traversal[S, B] {
	return Traversal[S, B]{
		Modify: func(ctx context.Context, s S, f func(context.Context, B) (B, error)) (S, error) {
			return t.Modify(ctx, s, func(ctx context.Context, a A) (A, error) {
				b, err := l.View(ctx, a)
				if err != nil {
					return a, err
				}
				b, err = f(ctx, b)
				if err != nil {
					return a, err
				}
				return l.Update(ctx, a, b)
			})
		},
	}
}

// ComposeTraversalPrism composes a Traversal and Prism and returns a Traversal.
func ComposeTraversalPrism[S, A, B any](t Traversal[S, A], p Prism[A, B]) Traversal[S, B] {
	return Traversal[S, B]{
		Modify: func(ctx context.Context, s S, f func(context.Context, B) (B, error)) (S, error) {
			return t.Modify(ctx, s, func(ctx context.Context, a A) (A, error) {
				b, err := p.Match(ctx, a)
				if err != nil {
					if errors.Is(err, ErrNoMatch) {
						return a, nil
					}
					return a, err
				}
				b, err = f(ctx, b)
				if err != nil {
					return a, err
				}
				return p.Build(ctx, b)
			})
		},
	}
}

// ComposeTraversalTraversal composes two Traversals and returns a Traversal.
func ComposeTraversalTraversal[S, A, B any](t Traversal[S, A], u Traversal[A, B]) Traversal[S, B] {
	return Traversal[S, B]{
		Modify: func(ctx context.Context, s S, f func(context.Context, B) (B, error)) (S, error) {
			return t.Modify(ctx, s, func(ctx context.Context, a A) (A, error) {
				return u.Modify(ctx, a, f)
			})
		},
	}
}

func Each[S ~[]T, T any]() Traversal[S, T] {
	return Traversal[S, T]{
		Modify: func(ctx context.Context, ts S, f func(context.Context, T) (T, error)) (S, error) {
			if len(ts) == 0 {
				return ts, nil
			}
			cs := make(S, len(ts))
			for i, t := range ts {
				t, err := f(ctx, t)
				if err != nil {
					return nil, err
				}
				cs[i] = t
			}
			return cs, nil
		},
	}
}
