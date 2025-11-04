package optics

import (
	"errors"
	"testing"
)

func PrismLaw[C, S, A any](t *testing.T, p Prism[C, S, A], c C, s S, a A, eqA func(A, A) bool, eqS func(S, S) bool) {
	t.Helper()

	t.Run("Match/Build", func(t *testing.T) {
		s, err := p.Build(c, a)
		if err != nil {
			t.Fatal(err)
		}
		m, err := p.Match(c, s)
		if err != nil {
			t.Fatal(err)
		}
		if !eqA(m, a) {
			t.Fatalf("Prism Match/Build failed, got %v", m)
		}
	})

	t.Run("Build/Match", func(t *testing.T) {
		a, err := p.Match(c, s)
		if err != nil {
			t.Fatal(err)
		}
		b, err := p.Build(c, a)
		if err != nil {
			t.Fatal(err)
		}
		if !eqS(b, s) {
			t.Fatalf("Prism Build/Match failed")
		}
	})
}

func TestPrismLawsOnPointer(t *testing.T) {
	p := Prism[struct{}, *int, int]{
		Match: func(_ struct{}, i *int) (int, error) {
			if i == nil {
				return 0, errors.New("nil pointer")
			}
			return *i, nil
		},
		Build: func(_ struct{}, i int) (*int, error) {
			return &i, nil
		},
	}
	n := 42
	PrismLaw(t, p, struct{}{}, &n, 7, func(i int, j int) bool {
		return i == j
	}, func(i *int, j *int) bool {
		switch {
		case i == nil && j == nil:
			return true
		case i == nil && j != nil:
			return false
		case i != nil && j == nil:
			return false
		}
		return *i == *j
	})
}
