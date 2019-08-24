package slice

import (
	"testing"
)

func Test_EqualStringSlice(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"a", "b", "c"}
	if !EqualStringSlice(a, b) {
		t.Error(a, "\nand\n", b, "\nis not the same?")
	}
	b[1] = "c"
	if EqualStringSlice(a, b) {
		t.Error(a, "\nand\n", b, "\nis the same?")
	}
	b = b[1:]
	if EqualStringSlice(a, b) {
		t.Error(a, "\nand\n", b, "\nis the same?")
	}
}

func Test_RemoveFromStringSlice(t *testing.T) {
	src := []string{"a", "b", "c"}
	dst := []string{"a", "c"}

	result := RemoveFromStringSlice("b", src)

	if !EqualStringSlice(dst, result) {
		t.Error("result of rm b from\n", src, "\nis\n", result)
	}
}
