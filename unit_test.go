package main

import (
	"math/big"
	"testing"
)

func TestFormatImperialExp(t *testing.T) {
	// This test covers scaling and exponents.
	sqFt := NewUnitLength(true).Mul(NewUnitLength(true))
	var area big.Rat
	area.Mul(LengthFeet, LengthFeet)
	area.Mul(&area, big.NewRat(3, 2))
	got := sqFt.Format(&area, true)
	want := "3/2 ft^2 ≈ 1.50000 ft^2"
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestFormatImperial(t *testing.T) {
	u := NewUnitLength(true)
	check := func(inches string, want, wantNeg string) {
		t.Helper()
		var r big.Rat
		r.SetString(inches)
		r.Mul(&r, LengthInches)
		got := u.Format(&r, true)
		if want != got {
			t.Errorf("for %s inches: want %q, got %q", inches, want, got)
		}
		r.Neg(&r)
		gotNeg := u.Format(&r, true)
		if wantNeg != gotNeg {
			t.Errorf("for -%s inches: want %q, got %q", inches, wantNeg, gotNeg)
		}
	}
	check("0", `0"`, `0"`)
	check("1", `1"`, `-1"`)
	check("1/2", `1/2"`, `-1/2"`)
	check("3/2", `1 1/2"`, `-1 1/2"`)
	check("12", `1' = 12"`, `-1' = -12"`)
	check("13", `1' 1" = 13"`, `-1' 1" = -13"`)
	check("1/128", `1/128" ≈ 0"`, `-1/128" ≈ 0"`)
	check("3/128", `3/128" ≈ 1/32"`, `-3/128" ≈ -1/32"`)
}
