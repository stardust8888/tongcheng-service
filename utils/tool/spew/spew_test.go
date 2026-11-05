package spew

import "testing"

func TestSpew(t *testing.T) {
	user := &struct {
		name string
		age  int
	}{name: "name", age: 10}
	Dump(user)
}
