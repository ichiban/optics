package optics

import (
	"context"
	"errors"
	"testing"
)

func PrismLaw[S, A any](t *testing.T, p Prism[S, A], s S, a A, eqA func(A, A) bool, eqS func(S, S) bool) {
	t.Helper()

	t.Run("Match/Build", func(t *testing.T) {
		ctx := t.Context()
		s, err := p.Build(ctx, a)
		if err != nil {
			t.Fatal(err)
		}
		m, err := p.Match(ctx, s)
		if err != nil {
			t.Fatal(err)
		}
		if !eqA(m, a) {
			t.Fatalf("Prism Match/Build failed, got %v", m)
		}
	})

	t.Run("Build/Match", func(t *testing.T) {
		ctx := t.Context()
		a, err := p.Match(ctx, s)
		if err != nil {
			t.Fatal(err)
		}
		b, err := p.Build(ctx, a)
		if err != nil {
			t.Fatal(err)
		}
		if !eqS(b, s) {
			t.Fatalf("Prism Build/Match failed")
		}
	})
}

func TestPrismLawsOnPointer(t *testing.T) {
	p := Prism[*int, int]{
		Match: func(_ context.Context, i *int) (int, error) {
			if i == nil {
				return 0, errors.New("nil pointer")
			}
			return *i, nil
		},
		Build: func(_ context.Context, i int) (*int, error) {
			return &i, nil
		},
	}
	n := 42
	PrismLaw(t, p, &n, 7, func(i int, j int) bool {
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
