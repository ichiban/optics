package optics

import "testing"

func TestSlice(t *testing.T) {
	s := [][]string{
		{
			"foo",
			"bar",
		},
		{
			"hoge",
			"fuga",
		},
	}
	foo := Slice[[][]string](0)
	bar := Slice[[]string](1)
	s, ok := Modify(s, Compose(foo, bar), func(string) string {
		return "baz"
	})
	if !ok {
		t.Error("expected true")
	}
	if s[0][1] != "baz" {
		t.Errorf("s[0][1] was not modified")
	}
}

func TestMap(t *testing.T) {
	m := map[string]map[string]int{
		"foo": {
			"bar": 1,
		},
	}
	foo := Map[map[string]map[string]int]("foo")
	bar := Map[map[string]int]("bar")
	m, ok := Modify(m, Compose(foo, bar), func(i int) int {
		return i + 1
	})
	if !ok {
		t.Error("expected true")
	}
	if m["foo"]["bar"] != 2 {
		t.Errorf("foo.bar was not modified")
	}
}
