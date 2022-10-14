package main

import (
	"fmt"
	"math/big"
	"strings"
)

// Unit represents the dimensions of a value.
//
// Values are normalized into a base unit for a given dimension (e.g.,
// all lengths are meters). However, Unit also tracks a display hint for
// which specific unit to use when rendering a value.
//
// The zero value for Unit is a valid, unitless unit.
type Unit struct {
	terms   []unitTerm // Sorted, no zero terms
	display unitDisplay
}

type unitTerm struct {
	base  unitBase
	power int
}

type unitBase int

const (
	unitBaseLength unitBase = iota // Meters
)

// TODO: If I want to support angles, normalizing between degrees and
// radians in a big.Rat is not going to end well. Maybe the base is
// actually degrees and we treat pi specially. (Crazily, trig functions
// of rational multiples of pi are apparently algebraic numbers.)

var (
	LengthInches = big.NewRat(254, 10000)  // Convert inches to base length
	LengthFeet   = big.NewRat(3048, 10000) // Convert feet to base length
)

type unitDisplay int

const (
	unitDisplayDefault unitDisplay = iota
	unitDisplayMetric
	unitDisplayImperial
)

var (
	big1    = big.NewRat(1, 1)
	big12   = big.NewRat(12, 1)
	big32   = big.NewRat(32, 1)
	bigHalf = big.NewRat(1, 2)
)

func NewUnitLength(imperial bool) Unit {
	display := unitDisplayMetric
	if imperial {
		display = unitDisplayImperial
	}
	return Unit{terms: []unitTerm{{unitBaseLength, 1}}, display: display}
}

// merge combines two term sets. It multiplies each y.base by yMul.
func merge(x, y []unitTerm, yMul int) []unitTerm {
	if len(x) == 0 && len(y) == 0 {
		return nil
	}
	if len(x) == 0 && yMul == 1 {
		return y
	}
	if len(y) == 0 {
		return append([]unitTerm(nil), x...)
	}
	z := make([]unitTerm, 0, len(x)+len(y))
	for xi, yi := 0, 0; xi < len(x) || yi < len(y); {
		if xi >= len(x) || y[yi].base < x[xi].base {
			z = append(z, unitTerm{y[yi].base, y[yi].power * yMul})
			yi++
		} else if x[xi].base < y[yi].base {
			z = append(z, x[xi])
			xi++
		} else {
			pow := x[xi].power + y[yi].power*yMul
			if pow != 0 {
				z = append(z, unitTerm{x[xi].base, pow})
			}
			xi++
			yi++
		}
	}
	if len(z) == 0 {
		return nil
	}
	return z
}

func mergeDisplay(x, y unitDisplay) unitDisplay {
	if x == unitDisplayDefault {
		return y
	}
	return x
}

func termsEq(a, b []unitTerm) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 || &a[0] == &b[0] {
		return true
	}
	for i, x := range a {
		if b[i] != x {
			return false
		}
	}
	return true
}

func (a Unit) Mul(b Unit) Unit {
	return Unit{merge(a.terms, b.terms, 1), mergeDisplay(a.display, b.display)}
}

func (a Unit) Div(b Unit) Unit {
	return Unit{merge(a.terms, b.terms, -1), mergeDisplay(a.display, b.display)}
}

func (a Unit) Add(b Unit) (Unit, bool) {
	if termsEq(a.terms, b.terms) {
		return Unit{a.terms, mergeDisplay(a.display, b.display)}, true
	}
	return Unit{}, false
}

func (u Unit) String() string {
	if len(u.terms) == 0 {
		return "<none>"
	}
	return u.format(nil)
}

func (u Unit) format(scaleOut *big.Rat) string {
	if scaleOut != nil {
		scaleOut.Set(big1)
	}
	var scaleTmp big.Rat

	var b strings.Builder
	sep := ""
	for _, t := range u.terms {
		if t.power > 0 {
			b.WriteString(sep)
			s := formatTerm(t.base, t.power, u.display, &scaleTmp)
			b.WriteString(s)
			sep = "·"
			if scaleOut != nil {
				scaleOut.Mul(scaleOut, &scaleTmp)
			}
		}
	}
	sep = "/"
	for _, t := range u.terms {
		if t.power < 0 {
			b.WriteString(sep)
			s := formatTerm(t.base, -t.power, u.display, &scaleTmp)
			b.WriteString(s)
			sep = "·"
			if scaleOut != nil {
				scaleOut.Quo(scaleOut, &scaleTmp)
			}
		}
	}
	return b.String()
}

func formatTerm(base unitBase, exp int, d unitDisplay, scaleOut *big.Rat) string {
	// TODO: For scaling, we have several options. Maybe this should
	// return a format configuration with all the options.
	var s string
	switch base {
	default:
		s = "<bad base>"
	case unitBaseLength:
		if d == unitDisplayImperial {
			s = "ft"
			if scaleOut != nil {
				expRat(scaleOut, LengthFeet, exp)
				scaleOut = nil
			}
		} else {
			s = "m"
		}
	}
	if scaleOut != nil {
		scaleOut.Set(big1)
	}
	if exp == 1 {
		return s
	}
	return fmt.Sprintf("%s^%d", s, exp)
}

func expRat(out *big.Rat, x *big.Rat, exp int) {
	if exp == 0 {
		out.Set(big1)
	} else if exp == 1 {
		out.Set(x)
	} else if exp == -1 {
		out.Inv(x)
	} else {
		inv := false
		if exp < 0 {
			exp = -exp
			inv = true
		}
		var n, d big.Int
		bigExp := big.NewInt(int64(exp))
		n.Exp(x.Num(), bigExp, nil)
		d.Exp(x.Denom(), bigExp, nil)
		if inv {
			out.SetFrac(&d, &n)
		} else {
			out.SetFrac(&n, &d)
		}
	}
}

func (u Unit) Format(val *big.Rat, verbose bool) string {
	if len(u.terms) == 1 && u.terms[0] == (unitTerm{unitBaseLength, 1}) && u.display == unitDisplayImperial {
		return u.formatImperialLength(val, verbose)
	}

	// TODO: Do a reasonable job of scaling this, like to ft or in or mm
	// or whatever. It's really unclear what to do if there's more than
	// one term, but it's clear what to do if there's only one term.
	// Maybe we just always scale whatever is the first term.
	var scale big.Rat
	str := u.format(&scale)
	if scale.Cmp(big1) != 0 {
		var scaled big.Rat
		scaled.Quo(val, &scale)
		val = &scaled
	}

	var b strings.Builder
	// TODO: Format as whole and fraction?
	b.WriteString(val.RatString())
	if str != "" {
		b.WriteByte(' ')
		b.WriteString(str)
	}
	if verbose && !val.IsInt() {
		// TODO: Adjustable precision
		// TODO: If exact, print "="
		b.WriteString(" ≈ ")
		b.WriteString(val.FloatString(5))
		if str != "" {
			b.WriteByte(' ')
			b.WriteString(str)
		}
	}
	return b.String()
}

func (u Unit) formatImperialLength(val *big.Rat, verbose bool) string {
	var b strings.Builder

	if val.Sign() == 0 {
		return "0\""
	}

	divMod := func(a, b *big.Rat) (div *big.Int, mod *big.Rat) {
		if a.Cmp(b) < 0 {
			return nil, a
		}
		var x big.Rat
		var divInt big.Int
		x.Quo(a, b)
		div = divInt.Div(x.Num(), x.Denom()) // div = floor(a/b)
		x.SetInt(div)
		x.Mul(&x, b)
		x.Sub(a, &x)
		mod = &x // mod = a - floor(a/b) * b
		return
	}

	printWholeAndFrac := func(b *strings.Builder, v *big.Rat) {
		whole, frac := divMod(v, big1)
		if whole != nil {
			fmt.Fprintf(b, "%s", whole)
			if frac.Sign() == 0 {
				return
			}
			b.WriteByte(' ')
		}
		b.WriteString(frac.RatString())
	}

	printFeetAndInches := func(b *strings.Builder, inches *big.Rat) {
		// Print whole feet.
		feet, inchesRem := divMod(inches, big12)
		if feet != nil {
			fmt.Fprintf(b, "%s'", feet)
			if inchesRem.Sign() == 0 {
				return
			}
			b.WriteByte(' ')
		}

		// Print remaining inches
		printWholeAndFrac(b, inchesRem)
		b.WriteByte('"')
	}

	var inches big.Rat
	inches.Quo(val, LengthInches)
	neg := inches.Sign() < 0
	if neg {
		inches.Neg(&inches)
	}

	// Print feet and inches
	if neg {
		b.WriteByte('-')
	}
	printFeetAndInches(&b, &inches)

	if !verbose {
		return b.String()
	}

	// Print inches
	if inches.Cmp(big12) >= 0 {
		b.WriteString(" = ")
		if neg {
			b.WriteByte('-')
		}
		printWholeAndFrac(&b, &inches)
		b.WriteByte('"')
	}

	// Print inches rounded to 32nds
	// TODO: Make this configurable
	var tmp big.Rat
	tmp.Mul(&inches, big32)
	if !tmp.IsInt() {
		b.WriteString(" ≈ ")
		// Round tmp to an int.
		var tmpInt big.Int
		tmp.Add(&tmp, bigHalf)
		tmpInt.Div(tmp.Num(), tmp.Denom())
		tmp.SetInt(&tmpInt)
		// Divide back by 32
		tmp.Quo(&tmp, big32)
		if neg && tmp.Sign() != 0 {
			b.WriteByte('-')
		}
		printWholeAndFrac(&b, &tmp)
		b.WriteByte('"')
	}

	return b.String()
}
