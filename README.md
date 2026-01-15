# Optics

An experimental optics library for Go.

Optics are composable abstractions for accessing and modifying nested data structures. They provide a functional approach to working with immutable data, eliminating the need for deep manual traversals.

## Go-Specific Adaptations

To make optics practical in Go, all operations:

- Accept `context.Context` as the first parameter
- Return `error` as a second return value

## Concepts

### Lens

A **Lens** focuses on exactly one value within a structure. It provides:

- `View`: Extract the focused value
- `Update`: Replace the focused value (returns a new structure)

```go
type Lens[S, A any] struct {
    View   func(context.Context, S) (A, error)
    Update func(context.Context, S, A) (S, error)
}
```

**Use cases**: Struct fields, map entries

```go
// Lens for a struct field
var userName = optics.Lens[User, string]{
    View: func(_ context.Context, u User) (string, error) {
        return u.Name, nil
    },
    Update: func(_ context.Context, u User, name string) (User, error) {
        u.Name = name
        return u, nil
    },
}

// Use the lens
name, _ := userName.View(ctx, user)           // get
user, _ = userName.Update(ctx, user, "Alice") // set
user, _ = userName.Over(ctx, user, func(_ context.Context, s string) (string, error) {
    return strings.ToUpper(s), nil
}) // modify
```

### Prism

A **Prism** focuses on a value that may or may not exist (sum types, variants). It provides:

- `Match`: Extract the value if present (returns `ErrNoMatch` otherwise)
- `Build`: Construct the structure from the value

```go
type Prism[S, A any] struct {
    Match func(context.Context, S) (A, error)
    Build func(context.Context, A) (S, error)
}
```

**Use cases**: Pointer dereferencing, type assertions, optional values

```go
// Prism for optional pointer values
someInt := optics.Optional[*int]()

var ptr *int = &value
n, err := someInt.Match(ctx, ptr) // n = value, err = nil

ptr = nil
n, err = someInt.Match(ctx, ptr)  // err = ErrNoMatch
```

### Traversal

A **Traversal** focuses on zero or more values within a structure:

- `Modify`: Apply a function to all focused values

```go
type Traversal[S, A any] struct {
    Modify func(context.Context, S, func(context.Context, A) (A, error)) (S, error)
}
```

**Use cases**: Slice elements, map values, recursive structures

```go
// Traversal over slice elements
users := optics.Each[[]User]()

// Double all ages
users, _ := users.Over(ctx, users, func(_ context.Context, u User) (User, error) {
    u.Age *= 2
    return u, nil
})
```

## Built-in Helpers

### Key

`Key` creates a lens focusing on a map entry:

```go
settings := map[string]int{"volume": 50, "brightness": 80}

volume := optics.Key[map[string]int, string, int]("volume")
v, _ := volume.View(ctx, settings)              // 50
settings, _ = volume.Update(ctx, settings, 75)  // {"volume": 75, "brightness": 80}
```

### Optional

`Optional` creates a prism for pointer types that matches non-nil values:

```go
someStr := optics.Optional[*string]()

s := "hello"
v, _ := someStr.Match(ctx, &s)     // "hello", nil
_, err := someStr.Match(ctx, nil)  // ErrNoMatch
```

### Each

`Each` creates a traversal over slice elements:

```go
numbers := optics.Each[[]int]()

nums := []int{1, 2, 3}
nums, _ = numbers.Over(ctx, nums, func(_ context.Context, n int) (int, error) {
    return n * 2, nil
}) // [2, 4, 6]
```

## Composition

Optics compose to build complex accessors from simple ones:

| Composition | Result |
|-------------|--------|
| Lens + Lens | Lens |
| Lens + Prism | Prism |
| Prism + Lens | Prism |
| Prism + Prism | Prism |
| Traversal + Lens | Traversal |
| Traversal + Prism | Traversal |
| Traversal + Traversal | Traversal |

```go
// Compose: []User -> User -> *int -> int
users := optics.Each[[]User]()
userAge := optics.ComposeTraversalLens(users, ageField)
someAge := optics.ComposeTraversalPrism(userAge, optics.Optional[*int]())

// Increment all non-nil ages
users, _ = someAge.Over(ctx, users, func(_ context.Context, age int) (int, error) {
    return age + 1, nil
})
```

## Error Handling

- `ErrNoMatch`: Returned by `Prism.Match` when the value doesn't match
- When composing with traversals, `ErrNoMatch` is handled gracefully (the element is skipped)
- Other errors propagate up and stop the operation

## Code Generation

Use `opticsgen` to automatically generate lenses for struct fields.

### Usage

Add a `go:generate` directive to your source file:

```go
//go:generate go tool github.com/ichiban/optics/cmd/opticsgen -type User
```

Then run:

```bash
go generate ./...
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-type` | Comma-separated list of type names (required) | - |
| `-output` | Output file name | `<type>_optics.go` |
| `-package` | Package path to analyze | `.` |

### Generated Code

For a struct:

```go
type user struct {
    name string
    age  *int
}
```

The generator produces:

```go
var userName = optics.Lens[user, string]{...}
var userAge = optics.Lens[user, *int]{...}
```

### Multiple Types

```go
//go:generate go tool github.com/ichiban/optics/cmd/opticsgen -type User,Post,Comment -output models_optics.go
```

## Complete Example

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

func ptr[T any](t T) *T { return &t }

//go:generate go tool github.com/ichiban/optics/cmd/opticsgen -type user

func main() {
    us := []user{
        {name: "alice", age: ptr(30)},
        {name: "bob", age: ptr(40)},
        {name: "charlie", age: nil},
    }

    // Compose optics: []user -> user -> *int -> int
    users := optics.Each[[]user]()
    usersAge := optics.ComposeTraversalLens(users, userAge)
    someUsersAge := optics.ComposeTraversalPrism(usersAge, optics.Optional[*int]())

    // Increment all non-nil ages
    us, err := someUsersAge.Over(context.Background(), us, func(_ context.Context, i int) (int, error) {
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

Output:

```
{name: alice, age: 31}
{name: bob, age: 41}
{name: charlie}
```

## License

MIT
