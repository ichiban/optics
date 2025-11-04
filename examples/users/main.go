package main

import (
	"context"
	"fmt"

	"github.com/ichiban/optics"
)

type user struct {
	name string
	age  *int
}

func (u user) String() string {
	if u.age == nil {
		return fmt.Sprintf("{name: %s}", u.name)
	}
	return fmt.Sprintf("{name: %s, age: %d}", u.name, *u.age)
}

func newInt(n int) *int {
	return &n
}

var users = optics.Traversal[context.Context, []user, user]{
	Modify: func(ctx context.Context, users []user, f func(context.Context, user) (user, error)) ([]user, error) {
		var (
			cs  = make([]user, len(users))
			err error
		)
		for i, u := range users {
			cs[i], err = f(ctx, u)
			if err != nil {
				return nil, err
			}
		}
		return cs, nil
	},
}

var age = optics.Lens[context.Context, user, *int]{
	View: func(ctx context.Context, user user) (*int, error) {
		return user.age, nil
	},
	Update: func(ctx context.Context, user user, i *int) (user, error) {
		user.age = i
		return user, nil
	},
}

var someInt = optics.Prism[context.Context, *int, int]{
	Match: func(ctx context.Context, i *int) (int, error) {
		if i == nil {
			return 0, optics.ErrNoMatch
		}
		return *i, nil
	},
	Build: func(ctx context.Context, i int) (*int, error) {
		return &i, nil
	},
}

func main() {
	us := []user{
		{
			name: "alice",
			age:  newInt(30),
		},
		{
			name: "bob",
			age:  newInt(40),
		},
		{
			name: "charlie",
			age:  nil,
		},
	}
	usersAge := optics.ComposeTraversalLens(users, age)
	someUsersAge := optics.ComposeTraversalPrism(usersAge, someInt)
	us, err := someUsersAge.Over(context.Background(), us, func(ctx context.Context, i int) (int, error) {
		return i + 1, nil
	})
	if err != nil {
		panic(err)
	}
	for _, u := range us {
		fmt.Println(u)
	}
}
