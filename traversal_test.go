package optics

import (
	"context"
	"slices"
	"testing"
)

func TraversalLaws[S, A any](t *testing.T, tr Traversal[S, A], s S, f, g func(A) A, eqS func(S, S) bool) {
	t.Helper()

	t.Run("Identity", func(t *testing.T) {
		ctx := t.Context()
		m, err := tr.Modify(ctx, s, func(ctx context.Context, a A) (A, error) {
			return a, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if !eqS(m, s) {
			t.Errorf("got %v, want %v", m, s)
		}
	})

	t.Run("Composition", func(t *testing.T) {
		ctx := t.Context()
		s1, err := tr.Modify(ctx, s, func(ctx context.Context, a A) (A, error) {
			return f(g(a)), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		s, err := tr.Modify(ctx, s, func(ctx context.Context, a A) (A, error) {
			return g(a), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		s2, err := tr.Modify(ctx, s, func(ctx context.Context, a A) (A, error) {
			return f(a), nil
		})
		if !eqS(s1, s2) {
			t.Errorf("got %v, want %v", s1, s2)
		}
	})
}

func TestTraversalLawsOnSlice(t *testing.T) {
	tr := Traversal[[]int, int]{
		Modify: func(ctx context.Context, s []int, f func(context.Context, int) (int, error)) ([]int, error) {
			s1 := make([]int, len(s))
			for i, a := range s {
				a, err := f(ctx, a)
				if err != nil {
					return nil, err
				}
				s1[i] = a
			}
			return s1, nil
		},
	}

	TraversalLaws(t, tr, []int{1, 2, 3}, func(i int) int {
		return i * 2
	}, func(i int) int {
		return i * 3
	}, func(s1 []int, s2 []int) bool {
		return slices.Compare(s1, s2) == 0
	})
}
