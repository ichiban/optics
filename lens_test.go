package optics

import (
	"context"
	"testing"
)

func LensLaws[C, S, A any](t *testing.T, l Lens[S, A], c C, s S, a [3]A, eqS func(S, S) bool, eqA func(A, A) bool) {
	t.Helper()

	t.Run("View/Update", func(t *testing.T) {
		ctx := t.Context()
		a, err := l.View(ctx, s)
		if err != nil {
			t.Fatal(err)
		}
		got, err := l.Update(ctx, s, a)
		if err != nil {
			t.Fatal(err)
		}
		if !eqS(got, s) {
			t.Fatalf("Lens View/Update failed: got %+v want %+v", got, s)
		}
	})

	t.Run("Update/View", func(t *testing.T) {
		ctx := t.Context()
		s, err := l.Update(ctx, s, a[0])
		if err != nil {
			t.Fatal(err)
		}
		g, err := l.View(ctx, s)
		if err != nil {
			t.Fatal(err)
		}
		if !eqA(g, a[0]) {
			t.Fatalf("Lens Update/View failed: got %v want %v", g, a[0])
		}
	})

	t.Run("Update/Update", func(t *testing.T) {
		ctx := t.Context()
		s, err := l.Update(ctx, s, a[1])
		if err != nil {
			t.Fatal(err)
		}
		p3, err := l.Update(ctx, s, a[2])
		if err != nil {
			t.Fatal(err)
		}
		p4, err := l.Update(ctx, s, a[2])
		if err != nil {
			t.Fatal(err)
		}
		if !eqS(p3, p4) {
			t.Fatalf("Lens Update/Update failed: got %+v want %+v", p3, p4)
		}
	})
}

func TestLensLawsOnPair(t *testing.T) {
	type P struct{ X, Y int }
	l := Lens[P, int]{
		View: func(_ context.Context, p P) (int, error) {
			return p.X, nil
		},
		Update: func(_ context.Context, p P, x int) (P, error) {
			p.X = x
			return p, nil
		},
	}

	p := P{X: 1, Y: 2}

	LensLaws(t, l, struct{}{}, p, [3]int{42, 10, 20}, func(p P, q P) bool {
		return p == q
	}, func(i int, j int) bool {
		return i == j
	})
}
