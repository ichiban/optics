package optics

import "errors"

var ErrNoMatch = errors.New("optics: no match")

// Prism captures a partial focus (sum/variant) with a builder (Build).
// It takes a contextual value C and may return an error.
type Prism[C, S, A any] struct {
	// Match returns (a, nil) when S carries A, otherwise (zero, ErrNoMatch).
	Match func(C, S) (A, error)
	// Build builds an S from an A (i.e., the constructor for the variant).
	Build func(C, A) (S, error)
}

// Preview reads A if present.
func (p Prism[C, S, A]) Preview(c C, s S) (A, error) {
	return p.Match(c, s)
}

// Over maps A inside S if present.
func (p Prism[C, S, A]) Over(c C, s S, f func(A) (A, error)) (S, error) {
	if a, err := p.Match(c, s); err != nil {
		a, err = f(a)
		if err != nil {
			return s, err
		}
		return p.Build(c, a)
	}
	return s, nil
}

func (p Prism[C, S, A]) Traversal() Traversal[C, S, A] {
	return Traversal[C, S, A]{
		Modify: func(c C, s S, f func(C, A) (A, error)) (S, error) {
			a, err := p.Match(c, s)
			if err != nil {
				if errors.Is(err, ErrNoMatch) {
					return s, nil
				}
				return s, err
			}
			a, err = f(c, a)
			if err != nil {
				return s, err
			}
			return p.Build(c, a)
		},
	}
}

// Modify runs f on the focus if present; returns (updated, hitCount 0|1).
func (p Prism[C, S, A]) Modify(c C, s S, f func(C, A) (A, error)) (S, error) {
	a, err := p.Match(c, s)
	if err != nil {
		if errors.Is(err, ErrNoMatch) {
			return s, nil
		}
		return s, err
	}
	a, err = f(c, a)
	if err != nil {
		return s, err
	}
	return p.Build(c, a)
}

// ComposePrismLens composes a Prism and Lens and returns a Prism.
func ComposePrismLens[C, S, A, B any](p Prism[C, S, A], l Lens[C, A, B]) Prism[C, S, B] {
	return Prism[C, S, B]{
		Match: func(c C, s S) (B, error) {
			a, err := p.Match(c, s)
			if err != nil {
				var z B
				return z, err
			}
			return l.View(c, a)
		},
		Build: func(c C, b B) (S, error) {
			var a A
			a, err := l.Update(c, a, b)
			if err != nil {
				var s S
				return s, err
			}
			return p.Build(c, a)
		},
	}
}

// ComposePrismPrism composes two Prisms and returns a Prism.
func ComposePrismPrism[C, S, A, B any](p Prism[C, S, A], q Prism[C, A, B]) Prism[C, S, B] {
	return Prism[C, S, B]{
		Match: func(c C, s S) (B, error) {
			a, err := p.Match(c, s)
			if err != nil {
				var b B
				return b, err
			}
			return q.Match(c, a)
		},
		Build: func(c C, b B) (S, error) {
			a, err := q.Build(c, b)
			if err != nil {
				var s S
				return s, err
			}
			return p.Build(c, a)
		},
	}
}

// ComposePrismTraversal composes a Prism and Traversal and returns a Traversal.
func ComposePrismTraversal[C, S, A, B any](p Prism[C, S, A], t Traversal[C, A, B]) Traversal[C, S, B] {
	return Traversal[C, S, B]{
		Modify: func(c C, s S, f func(C, B) (B, error)) (S, error) {
			a, err := p.Match(c, s)
			if err != nil {
				var s S
				return s, err
			}
			a, err = t.Modify(c, a, f)
			if err != nil {
				var s S
				return s, err
			}
			return p.Build(c, a)
		},
	}
}

func Optional[C, T any]() Prism[C, *T, T] {
	return Prism[C, *T, T]{
		Match: func(_ C, t *T) (T, error) {
			if t == nil {
				var zero T
				return zero, ErrNoMatch
			}
			return *t, nil
		},
		Build: func(_ C, t T) (*T, error) {
			return &t, nil
		},
	}
}
