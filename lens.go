package optics

// Lens focuses a single substructure of S of type A.
// It takes a context value C and may return an error.
type Lens[C, S, A any] struct {
	View   func(C, S) (A, error)
	Update func(C, S, A) (S, error)
}

// Over modifies the focused value with f.
func (l Lens[C, S, A]) Over(c C, s S, f func(C, A) (A, error)) (S, error) {
	a, err := l.View(c, s)
	if err != nil {
		return s, err
	}
	a, err = f(c, a)
	if err != nil {
		return s, err
	}
	return l.Update(c, s, a)
}

// ComposeLensLens composes two lenses.
func ComposeLensLens[C, S, A, B any](l1 Lens[C, S, A], l2 Lens[C, A, B]) Lens[C, S, B] {
	return Lens[C, S, B]{
		View: func(c C, s S) (B, error) {
			a, err := l1.View(c, s)
			if err != nil {
				var b B
				return b, err
			}
			return l2.View(c, a)
		},
		Update: func(c C, s S, b B) (S, error) {
			a, err := l1.View(c, s)
			if err != nil {
				return s, err
			}
			a, err = l2.Update(c, a, b)
			if err != nil {
				return s, err
			}
			return l1.Update(c, s, a)
		},
	}
}

// ComposeLensPrism composes a lens and prism.
func ComposeLensPrism[C, S, A, B any](l Lens[C, S, A], p Prism[C, A, B]) Prism[C, S, B] {
	return Prism[C, S, B]{
		Match: func(c C, s S) (B, error) {
			a, err := l.View(c, s)
			if err != nil {
				var b B
				return b, err
			}
			return p.Match(c, a)
		},
		Build: func(c C, b B) (S, error) {
			var s S
			a, err := p.Build(c, b)
			if err != nil {
				return s, err
			}
			return l.Update(c, s, a)
		},
	}
}
