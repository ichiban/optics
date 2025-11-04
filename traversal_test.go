package optics

import (
	"slices"
	"testing"
)

func TraversalLaws[C, S, A any](t *testing.T, tr Traversal[C, S, A], c C, s S, f, g func(A) A, eqS func(S, S) bool) {
	t.Helper()

	t.Run("Identity", func(t *testing.T) {
		m, err := tr.Modify(c, s, func(c C, a A) (A, error) {
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
		s1, err := tr.Modify(c, s, func(c C, a A) (A, error) {
			return f(g(a)), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		s, err := tr.Modify(c, s, func(c C, a A) (A, error) {
			return g(a), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		s2, err := tr.Modify(c, s, func(c C, a A) (A, error) {
			return f(a), nil
		})
		if !eqS(s1, s2) {
			t.Errorf("got %v, want %v", s1, s2)
		}
	})
}

func TestTraversalLawsOnSlice(t *testing.T) {
	tr := Traversal[struct{}, []int, int]{
		Modify: func(c struct{}, s []int, f func(struct{}, int) (int, error)) ([]int, error) {
			s1 := make([]int, len(s))
			for i, a := range s {
				a, err := f(c, a)
				if err != nil {
					return nil, err
				}
				s1[i] = a
			}
			return s1, nil
		},
	}

	TraversalLaws(t, tr, struct{}{}, []int{1, 2, 3}, func(i int) int {
		return i * 2
	}, func(i int) int {
		return i * 3
	}, func(s1 []int, s2 []int) bool {
		return slices.Compare(s1, s2) == 0
	})
}
