package main

import (
	"fmt"
	"math/big"
	"strings"
)

type Val struct {
	val  big.Rat
	unit Unit
}

type Unit int

const (
	UnitNone Unit = iota
	UnitInch
)

var (
	bigOne    = big.NewRat(1, 1)
	bigTwelve = big.NewRat(12, 1)
)

func (x Val) Add(y Val) (Val, error) {
	if x.unit != y.unit {
		return Val{}, fmt.Errorf("cannot add dimensions with different units: %s and %s", x, y)
	}
	var z Val
	z.val.Add(&x.val, &y.val)
	z.unit = x.unit
	return z, nil
}

func (x Val) Mul(y Val) (Val, error) {
	// At least one must be dimensionless
	var z Val
	if x.unit == UnitNone && y.unit == UnitNone {
		z.unit = UnitNone
	} else if x.unit == UnitNone && y.unit != UnitNone {
		z.unit = y.unit
	} else if x.unit != UnitNone && y.unit == UnitNone {
		z.unit = x.unit
	} else {
		// TODO: Implement this
		return Val{}, fmt.Errorf("cannot multiply units: %s and %s", x, y)
	}
	z.val.Mul(&x.val, &y.val)
	return z, nil
}

func (x Val) Div(y Val) (Val, error) {
	var z Val
	if x.unit == y.unit {
		z.unit = UnitNone
	} else if x.unit == UnitNone && y.unit != UnitNone {
		// TODO: Implement this
		return Val{}, fmt.Errorf("not implemented: %s / %s", x, y)
	} else if x.unit != UnitNone && y.unit == UnitNone {
		z.unit = x.unit
	} else {
		panic("unexpected case")
	}
	z.val.Quo(&x.val, &y.val)
	return z, nil
}

func (x Val) String() string {
	var b strings.Builder

	switch x.unit {
	default:
		panic("unexpected unit")
	case UnitNone:
		return x.val.RatString()
	case UnitInch:
		var val big.Rat
		val.Set(&x.val)

		if val.Sign() == 0 {
			return "0\""
		} else if val.Sign() < 0 {
			b.WriteByte('-')
			val.Neg(&val)
		}

		// Print whole feet.
		var feet big.Rat
		feet.Quo(&val, bigTwelve)
		if feet.Cmp(bigOne) >= 0 {
			var feetInt big.Int
			feetInt.Div(feet.Num(), feet.Denom())
			fmt.Fprintf(&b, "%s'", &feetInt)
			// Compute remainder
			feet.SetInt(&feetInt)
			val.Sub(&val, feet.Mul(&feet, bigTwelve))
			if val.Sign() == 0 {
				break
			}
			b.WriteByte(' ')
		}
		// val = remainder inches.

		// Print whole inches
		if val.Cmp(bigOne) >= 0 {
			var inchInt big.Int
			inchInt.Div(val.Num(), val.Denom())
			fmt.Fprintf(&b, "%s", &inchInt)
			// Compute remainder
			var inch big.Rat
			inch.SetInt(&inchInt)
			val.Sub(&val, &inch)
			if val.Sign() == 0 {
				b.WriteByte('"')
				break
			}
			b.WriteByte(' ')
		}
		// val = remainder fractional inches

		b.WriteString(val.RatString())
		b.WriteByte('"')
	}
	return b.String()
}
