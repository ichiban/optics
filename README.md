# Optics

An experimental optics library in Go.

To make it work in Go, we made some alterations:

- Functions take an additional contextual value `C` which doesn't have to be `context.Context`.
- Functions may return an `error`.

## Example

```go
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

func some[T any](t T) *T {
	return &t
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

func main() {
	us := []user{
		{
			name: "alice",
			age:  some(30),
		},
		{
			name: "bob",
			age:  some(40),
		},
		{
			name: "charlie",
			age:  nil,
		},
	}
	users := optics.Each[context.Context, []user]()
	usersAge := optics.ComposeTraversalLens(users, age)
	someInt := optics.Optional[context.Context, int]()
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
```

```
{name: alice, age: 31}
{name: bob, age: 41}
{name: charlie}
```