package optics

import "errors"

// Traversal modifies 0..n focuses, returning the updated S and how many were hit.
// It takes a contextual value C and may return an error.
type Traversal[C, S, A any] struct {
	Modify func(C, S, func(C, A) (A, error)) (S, error)
}

// Over is a convenience alias for Modify.
func (t Traversal[C, S, A]) Over(c C, s S, f func(C, A) (A, error)) (S, error) {
	return t.Modify(c, s, f)
}

// ComposeTraversalLens composes a Traversal and Lens and returns a Traversal.
func ComposeTraversalLens[C, S, A, B any](t Traversal[C, S, A], l Lens[C, A, B]) Traversal[C, S, B] {
	return Traversal[C, S, B]{
		Modify: func(c C, s S, f func(C, B) (B, error)) (S, error) {
			return t.Modify(c, s, func(c C, a A) (A, error) {
				b, err := l.View(c, a)
				if err != nil {
					return a, err
				}
				b, err = f(c, b)
				if err != nil {
					return a, err
				}
				return l.Update(c, a, b)
			})
		},
	}
}

// ComposeTraversalPrism composes a Traversal and Prism and returns a Traversal.
func ComposeTraversalPrism[C, S, A, B any](t Traversal[C, S, A], p Prism[C, A, B]) Traversal[C, S, B] {
	return Traversal[C, S, B]{
		Modify: func(c C, s S, f func(C, B) (B, error)) (S, error) {
			return t.Modify(c, s, func(c C, a A) (A, error) {
				b, err := p.Match(c, a)
				if err != nil {
					if errors.Is(err, ErrNoMatch) {
						return a, nil
					}
					return a, err
				}
				b, err = f(c, b)
				if err != nil {
					return a, err
				}
				return p.Build(c, b)
			})
		},
	}
}

// ComposeTraversalTraversal composes two Traversals and returns a Traversal.
func ComposeTraversalTraversal[C, S, A, B any](t Traversal[C, S, A], u Traversal[C, A, B]) Traversal[C, S, B] {
	return Traversal[C, S, B]{
		Modify: func(c C, s S, f func(C, B) (B, error)) (S, error) {
			return t.Modify(c, s, func(c C, a A) (A, error) {
				return u.Modify(c, a, f)
			})
		},
	}
}
