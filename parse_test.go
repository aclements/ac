package main

import (
	"testing"
)

func p(t *testing.T, s string, want string) {
	t.Helper()
	v, err := Parse(s)
	if err != nil {
		t.Errorf("Parse(%q) failed: %s", s, err)
		return
	}
	got := v.unit.Format(&v.val, false)
	if got != want {
		t.Errorf("Parse(%q) = %s, want %s", s, got, want)
	}
}

func TestParseNumber(t *testing.T) {
	// Basic parsing
	p(t, "0", "0")
	p(t, "1", "1")
	p(t, "1/2", "1/2")
	p(t, "1 1/2", "3/2")

	// Test imperial lengths
	p(t, `1'`, `1'`)
	p(t, `1"`, `1"`)
	p(t, `1' 1"`, `1' 1"`)
}

func TestParseOrder(t *testing.T) {
	// Order of operations
	p(t, "1 + 2 * 3", "7")
	p(t, "2 * 3 + 4", "10")
	p(t, "2 * (3 + 4)", "14")
	p(t, "(((1)))", "1")

	// Associativity
	p(t, "5 - 1 - 1", "3")
	p(t, "12 / 4 / 3", "1")

	// Unary
	p(t, "-1", "-1")
	p(t, "--1", "1")
	p(t, "1 + -1", "0")
	p(t, "-(1)", "-1")
}

func TestParseError(t *testing.T) {
	c := func(s string, pos int, msg string) {
		t.Helper()
		_, err := Parse(s)
		if err == nil {
			t.Errorf("Parse(%q) unexpectedly succeeded", s)
			return
		}
		se, ok := err.(*SyntaxError)
		if !ok {
			t.Errorf("Parse(%q) failed with %T", s, err)
			return
		}
		if se.pos != pos || se.msg != msg {
			t.Errorf("Parse(%q) failed with %d:%s, want %d:%s", s, se.pos, se.msg, pos, msg)
		}
	}

	c("~", 0, "unexpected token")
	c("1 1", 2, "expected end")
	c("1' + 1", 3, "cannot add dimensions with different units: 1' and 1")
	c("1 / 0", 2, "division by zero")
	c("1 +", 3, "unexpected end")
	c("(1", 2, "expected `)`")
	c("1 + +", 4, "unexpected `+`")
}
