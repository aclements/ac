package main

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

func Parse(s string) (Val, error) {
	toks, err := tokenize(s)
	if err != nil {
		return Val{}, err
	}

	p := parser{}
	x := p.expr(toks)
	if p.err != nil {
		return Val{}, p.err
	}
	if p.err2 != nil {
		return Val{}, p.err2
	}
	return x, nil
}

type SyntaxError struct {
	pos int
	msg string
}

func (e *SyntaxError) Error() string {
	return e.msg
}

type tok struct {
	pos  int
	kind byte
	val  big.Rat
}

func (t tok) KindString() string {
	switch t.kind {
	case 0:
		return "end"
	case 'n':
		return "number"
	}
	return fmt.Sprintf("`%c`", t.kind)
}

var numRe = regexp.MustCompile("^([0-9]+/[0-9]+)|([0-9.]+)")

func tokenize(s string) ([]tok, error) {
	var toks []tok
	mergeFrac := false
	sOrig := s

	for {
		for len(s) > 0 && isSpace(s[0]) {
			s = s[1:]
		}
		if len(s) == 0 {
			break
		}
		pos := len(sOrig) - len(s)

		if ('0' <= s[0] && s[0] <= '9') || s[0] == '.' {
			var val big.Rat
			num := numRe.FindString(s)
			if _, ok := val.SetString(num); !ok {
				// Should never happen since we passed numRe.
				return nil, &SyntaxError{pos: pos, msg: "malformed number"}
			}
			s = s[len(num):]

			isFrac := strings.ContainsRune(num, '/')
			if isFrac && mergeFrac {
				// Merge this fractional value into the previous token
				mergeFrac = false
				v := &toks[len(toks)-1].val
				v.Add(v, &val)
				continue
			}
			// If this was a non-fraction, allow merging a following fraction.
			mergeFrac = !isFrac

			toks = append(toks, tok{pos, 'n', val})
			continue
		}
		mergeFrac = false

		switch s[0] {
		case '(', ')', '+', '-', '*', '/', '\'', '"':
			toks = append(toks, tok{pos: pos, kind: s[0]})
			s = s[1:]
		default:
			return nil, &SyntaxError{pos: pos, msg: "unexpected token"}
		}
	}
	toks = append(toks, tok{pos: len(sOrig), kind: 0})
	return toks, nil
}

func isSpace(c byte) bool {
	const mask uint64 = 1<<'\t' | 1<<'\n' | 1<<'\v' | 1<<'\f' | 1<<'\r' | 1<<' '
	return mask&(1<<c) != 0
}

type parser struct {
	err  *SyntaxError
	err2 *SyntaxError
}

func (p *parser) error(tok tok, f string, args ...interface{}) {
	if p.err != nil && p.err.pos < tok.pos {
		return
	}
	p.err = &SyntaxError{pos: tok.pos, msg: fmt.Sprintf(f, args...)}
}

func (p *parser) mathError(tok tok, err error) {
	if p.err2 == nil {
		p.err2 = &SyntaxError{pos: tok.pos, msg: err.Error()}
	}
}

func (p *parser) expr(toks []tok) Val {
	toks, x := p.addExpr(toks)
	if toks[0].kind != 0 {
		p.error(toks[0], "expected end")
	}
	return x
}

func (p *parser) addExpr(toks []tok) ([]tok, Val) {
	toks, x := p.mulExpr(toks)

	for toks[0].kind == '+' || toks[0].kind == '-' {
		op := toks[0]
		var y Val
		toks, y = p.mulExpr(toks[1:])
		if toks == nil {
			return nil, Val{}
		}
		if op.kind == '-' {
			y.val.Neg(&y.val)
		}
		var err error
		x, err = x.Add(y)
		if err != nil {
			p.mathError(op, err)
		}
	}

	return toks, x
}

func (p *parser) mulExpr(toks []tok) ([]tok, Val) {
	toks, x := p.numExpr(toks)

	for toks[0].kind == '*' || toks[0].kind == '/' {
		op := toks[0]
		var y Val
		toks, y = p.numExpr(toks[1:])
		var err error
		if op.kind == '*' {
			x, err = x.Mul(y)
		} else {
			x, err = x.Div(y)
		}
		if err != nil {
			p.mathError(op, err)
		}
	}

	return toks, x
}

func (p *parser) numExpr(toks []tok) ([]tok, Val) {
	switch toks[0].kind {
	case '(':
		toks, x := p.addExpr(toks[1:])
		if toks[0].kind == ')' {
			return toks[1:], x
		}
		p.error(toks[0], "expected `)`")
		return toks, Val{}

	case 'n':
		return p.number(toks)

	case '-':
		toks, x := p.numExpr(toks[1:])
		x.val.Neg(&x.val)
		return toks, x
	}

	p.error(toks[0], "unexpected "+toks[0].KindString())
	return toks, Val{}
}

func (p *parser) number(toks []tok) ([]tok, Val) {
	// number :=
	//  <n>
	//  <n> ' [<n> "]
	//  <n> "

	var x Val
	var tmp big.Rat
	switch toks[1].kind {
	case '\'':
		// Feet
		x.val.Add(&x.val, tmp.Mul(&toks[0].val, LengthFeet))
		toks = toks[2:]
		if !(toks[0].kind == 'n' && toks[1].kind == '"') {
			x.unit = NewUnitLength(true)
			return toks, x
		}
		fallthrough
	case '"':
		// Inches (possibly preceded by feet)
		x.val.Add(&x.val, tmp.Mul(&toks[0].val, LengthInches))
		x.unit = NewUnitLength(true)
		toks = toks[2:]
		return toks, x
	}

	// Just a number
	x.val.Set(&toks[0].val)
	toks = toks[1:]
	return toks, x
}
