package math

import "testing"

func TestParse0(t *testing.T) {
	_, tokens := Lexer(`

5*6

    `)
	ds := Parse(tokens)
	if ds.Val(nil) != 30 {
		t.Error(ds.Val(nil), "not equals 30")
	}
}

func TestParse(t *testing.T) {
	_, tokens := Lexer(`

5*6 - 3 * 2

    `)
	ds := Parse(tokens)
	if ds.Val(nil) != 24 {
		t.Error(ds.Val(nil), "not equals 24")
	}
}

func TestParse1(t *testing.T) {
	_, tokens := Lexer(`

5*6 - 3

    `)
	ds := Parse(tokens)
	if ds.Val(nil) != 27 {
		t.Error(ds.Val(nil), "not equals 27")
	}
}

func TestParseVar(t *testing.T) {
	_, tokens := Lexer(`

$1*6 - 3

    `)
	ds := Parse(tokens)
	if ds.Val([]int{1}) != 3 {
		t.Error(ds.Val([]int{1}), "not equals 3")
	}
}

func TestParseVar2(t *testing.T) {
	_, tokens := Lexer(`

$1*$2 - 3

    `)
	ds := Parse(tokens)
	if ds.Val([]int{1, 3}) != 0 {
		t.Error(ds.Val([]int{1, 3}), "not equals 0")
	}
}

func TestParseClosure(t *testing.T) {
	_, tokens := Lexer(`

$1* ($2 - 3)

    `)
	ds := Parse(tokens)
	if ds.Val([]int{2, 4}) != 2 {
		t.Error(ds.Val([]int{2, 4}), "not equals 2")
	}
}

func TestParseClosure1(t *testing.T) {
	_, tokens := Lexer(`

 ($2 - 3) * $1

    `)
	ds := Parse(tokens)
	if ds.Val([]int{2, 4}) != 2 {
		t.Error(ds.Val([]int{2, 4}), "not equals 2")
	}
}

func TestParseClosure3(t *testing.T) {
	_, tokens := Lexer(`

 ($2 - 3) * $1 + (1 +1)+1

    `)
	ds := Parse(tokens)
	if ds.Val([]int{2, 4}) != 5 {
		t.Error(ds.Val([]int{2, 4}), "not equals 5")
	}
}

func TestParseClosure0(t *testing.T) {
	_, tokens := Lexer(`

 ($2 - 3) 

    `)
	ds := Parse(tokens)
	if ds.Val([]int{2, 4}) != 1 {
		t.Error(ds.Val([]int{2, 4}), "not equals 1")
	}
}

// ($1-48)*16+($2-48)+($3-48)*4096+($4-48)*256

func TestParseClosure2(t *testing.T) {
	_, tokens := Lexer(`

 ($1-48)*16+($2-48)+($3-48)*4096+($4-48)*256

    `)
	ds := Parse(tokens)
	if ds.Val([]int{49, 48, 48, 48}) != 16 {
		t.Error(ds.Val([]int{49, 48, 48, 48}), "not equals 16")
	}
}
