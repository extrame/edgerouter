package math

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// LexToken holds is a (type, value) array.
type LexToken [3]string

// EOF character
var EOF string = "+++EOF+++"

// lexerState represents the state of the scanner
// as a function that returns the next state.
type lexerState func(*lexer) lexerState

// run lexes the input by executing state functions until
// the state is nil.
func (l *lexer) Run() {
	for state := l.initialState; state != nil; {
		state = state(l)
	}
}

// Lexer creates a new scanner for the input string.
func Lexer(input string) (*lexer, []LexToken) {
	l := &lexer{
		input:  input,
		tokens: make([]LexToken, 0),
		lineno: 1,
	}
	l.initialState = initLexerState
	l.Run()
	return l, l.tokens
}

// lexer holds the state of the scanner.
type lexer struct {
	input        string     // the string being scanned.
	start        int        // start position of this item.
	pos          int        // current position in the input.
	width        int        // width of last rune read from input.
	tokens       []LexToken // scanned items.
	initialState lexerState

	lineno int
}

// next returns the next rune in the input.
func (l *lexer) next() string {
	var r rune
	if l.pos >= len(l.input) {
		l.width = 0
		return EOF
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return string(r)
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.Index(valid, l.next()) >= 0 {
	}
	l.backup()
}

// emit passes an item back to the client.
func (l *lexer) emit(t string) {
	l.tokens = append(l.tokens, LexToken{t, l.input[l.start:l.pos], strconv.Itoa(l.lineno)})
	l.start = l.pos
}

// initialState is the starting point for the
// scanner. It scans through each character and decides
// which state to create for the lexer. lexerState == nil
// is exit scanner.
func initLexerState(l *lexer) lexerState {
	for r := l.next(); r != EOF; r = l.next() {
		if r == " " || r == "\t" || r == "\r" {
			l.ignore()
		} else if r == "\n" {
			l.lineno += 1
			l.ignore()
		} else if isItem(r) {
			l.backup()
			return itemLexerState
		} else if r == "(" {
			l.emit(T_LCLOSURE_MARK)
			return initLexerState
		} else if r == ")" {
			l.emit(T_RCLOSURE_MARK)
			return initLexerState
		} else {
			l.backup()
			return operatorLexerState
		}
	}

	l.emit(T_EOF)
	return nil
}

func itemLexerState(l *lexer) lexerState {
	for r := l.next(); r != EOF; r = l.next() {
		if r == "\r" {
			l.ignore()
		} else if r == "\n" {
			l.lineno += 1
			l.ignore()
		} else if r == "$" {
			l.acceptRun(numbers)
			l.emit(T_VAR_MARK)
		} else if isNumber(r) {
			l.acceptRun(numbers + ".")
			l.emit(T_NUMBER_MARK)
		} else {
			l.backup()
			return initLexerState
		}
	}
	l.emit(T_EOF)
	return nil
}

func operatorLexerState(l *lexer) lexerState {
	for r := l.next(); r != EOF; r = l.next() {
		if r == "\r" || r == " " {
			l.ignore()
		} else if r == "\n" {
			l.lineno += 1
			l.ignore()
		} else if isOperator(r) {
			l.acceptRun(operatorValues)
			l.emit(T_OPER_MARK)
			return operatorLexerState
		} else {
			l.backup()
			return itemLexerState
		}
	}
	l.emit(T_EOF)
	return nil
}

// isName() checks if a character is an alpha
func isItem(char string) bool {
	if strings.Index(itemvalues, char) >= 0 {
		return true
	} else {
		return false
	}
}

// isOperator() checks if a character is an operator
func isOperator(char string) bool {
	if strings.Index(operatorValues, char) >= 0 {
		return true
	} else {
		return false
	}
}

// isNumber() checks if a character is a number
func isNumber(char string) bool {
	if strings.Index(numbers, char) >= 0 {
		return true
	} else {
		return false
	}
}
