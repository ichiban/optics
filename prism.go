package optics

import (
	"context"
	"errors"
)

var ErrNoMatch = errors.New("optics: no match")

// Prism captures a partial focus (sum/variant) with a builder (Build).
// It takes a contextual value C and may return an error.
type Prism[S, A any] struct {
	// Match returns (a, nil) when S carries A, otherwise (zero, ErrNoMatch).
	Match func(context.Context, S) (A, error)
	// Build builds an S from an A (i.e., the constructor for the variant).
	Build func(context.Context, A) (S, error)
}

// Preview reads A if present.
func (p Prism[S, A]) Preview(ctx context.Context, s S) (A, error) {
	return p.Match(ctx, s)
}

// Over maps A inside S if present.
func (p Prism[S, A]) Over(ctx context.Context, s S, f func(context.Context, A) (A, error)) (S, error) {
	if a, err := p.Match(ctx, s); err != nil {
		a, err = f(ctx, a)
		if err != nil {
			return s, err
		}
		return p.Build(ctx, a)
	}
	return s, nil
}

func (p Prism[S, A]) Traversal() Traversal[S, A] {
	return Traversal[S, A]{
		Modify: func(ctx context.Context, s S, f func(context.Context, A) (A, error)) (S, error) {
			a, err := p.Match(ctx, s)
			if err != nil {
				if errors.Is(err, ErrNoMatch) {
					return s, nil
				}
				return s, err
			}
			a, err = f(ctx, a)
			if err != nil {
				return s, err
			}
			return p.Build(ctx, a)
		},
	}
}

// Modify runs f on the focus if present; returns (updated, hitCount 0|1).
func (p Prism[S, A]) Modify(ctx context.Context, s S, f func(context.Context, A) (A, error)) (S, error) {
	a, err := p.Match(ctx, s)
	if err != nil {
		if errors.Is(err, ErrNoMatch) {
			return s, nil
		}
		return s, err
	}
	a, err = f(ctx, a)
	if err != nil {
		return s, err
	}
	return p.Build(ctx, a)
}

// ComposePrismLens composes a Prism and Lens and returns a Prism.
func ComposePrismLens[S, A, B any](p Prism[S, A], l Lens[A, B]) Prism[S, B] {
	return Prism[S, B]{
		Match: func(ctx context.Context, s S) (B, error) {
			a, err := p.Match(ctx, s)
			if err != nil {
				var z B
				return z, err
			}
			return l.View(ctx, a)
		},
		Build: func(ctx context.Context, b B) (S, error) {
			var a A
			a, err := l.Update(ctx, a, b)
			if err != nil {
				var s S
				return s, err
			}
			return p.Build(ctx, a)
		},
	}
}

// ComposePrismPrism composes two Prisms and returns a Prism.
func ComposePrismPrism[S, A, B any](p Prism[S, A], q Prism[A, B]) Prism[S, B] {
	return Prism[S, B]{
		Match: func(ctx context.Context, s S) (B, error) {
			a, err := p.Match(ctx, s)
			if err != nil {
				var b B
				return b, err
			}
			return q.Match(ctx, a)
		},
		Build: func(ctx context.Context, b B) (S, error) {
			a, err := q.Build(ctx, b)
			if err != nil {
				var s S
				return s, err
			}
			return p.Build(ctx, a)
		},
	}
}

// ComposePrismTraversal composes a Prism and Traversal and returns a Traversal.
func ComposePrismTraversal[S, A, B any](p Prism[S, A], t Traversal[A, B]) Traversal[S, B] {
	return Traversal[S, B]{
		Modify: func(ctx context.Context, s S, f func(context.Context, B) (B, error)) (S, error) {
			a, err := p.Match(ctx, s)
			if err != nil {
				var s S
				return s, err
			}
			a, err = t.Modify(ctx, a, f)
			if err != nil {
				var s S
				return s, err
			}
			return p.Build(ctx, a)
		},
	}
}

func Optional[S ~*T, T any]() Prism[S, T] {
	return Prism[S, T]{
		Match: func(_ context.Context, s S) (T, error) {
			if s == nil {
				var zero T
				return zero, ErrNoMatch
			}
			return *s, nil
		},
		Build: func(_ context.Context, t T) (S, error) {
			return &t, nil
		},
	}
}
