package main

import (
	"fmt"
	"math/big"
)

type Val struct {
	val  big.Rat
	unit Unit
}

func (x Val) Add(y Val) (Val, error) {
	var z Val
	var ok bool
	z.unit, ok = x.unit.Add(y.unit)
	if !ok {
		return Val{}, fmt.Errorf("cannot add dimensions with different units: %s and %s", x, y)
	}
	z.val.Add(&x.val, &y.val)
	return z, nil
}

func (x Val) Mul(y Val) (Val, error) {
	var z Val
	z.val.Mul(&x.val, &y.val)
	z.unit = x.unit.Mul(y.unit)
	return z, nil
}

func (x Val) Div(y Val) (Val, error) {
	if y.val.Sign() == 0 {
		return Val{}, fmt.Errorf("division by zero")
	}
	var z Val
	z.val.Quo(&x.val, &y.val)
	z.unit = x.unit.Div(y.unit)
	return z, nil
}

func (x Val) String() string {
	return x.unit.Format(&x.val, false)
}

func (x Val) VerboseString() string {
	return x.unit.Format(&x.val, true)
}
