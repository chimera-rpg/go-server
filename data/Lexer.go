package data

import (
	"unicode"

	"github.com/eczarny/lexer"
)

const (
	TokenVariable lexer.TokenType = iota
	TokenValue
	TokenNewline

	TokenComment

	TokenContainerBegin
	TokenContainerEnd

	TokenEOF
)

func getTokenName(token lexer.Token) string {
	switch token.Type {
	case TokenVariable:
		return "VAR"
	case TokenValue:
		return "VAL"
	case TokenNewline:
		return "NL"
	case TokenComment:
		return "CMT"
	case TokenContainerBegin:
		return "CONT_B"
	case TokenContainerEnd:
		return "CONT_E"
	case TokenEOF:
		return "EOF"
	}
	return "Err"
}

func NewObjectLexer(input string) *lexer.Lexer {
	return lexer.NewLexer(input, initialState)
}

// states

func initialState(l *lexer.Lexer) lexer.StateFunc {
	r := l.IgnoreUpTo(func(r rune) bool {
		return tNewline(r) || tNonWhitespace(r)
	})
	switch {
	case tComment(r):
		return commentState
	case tQuote(r):
		return quotedVariableState
	case tNewline(r):
		return newlineState
	case tBraceLeft(r):
		return braceLeftState
	case tBraceRight(r):
		return braceRightState
	case tNonWhitespace(r) && r != lexer.EOF:
		return variableState
	}
	l.Emit(TokenEOF)
	return nil
}

func commentState(l *lexer.Lexer) lexer.StateFunc {
	l.IgnoreUpTo(func(r rune) bool {
		return tNewline(r)
	})
	return initialState
}

func braceLeftState(l *lexer.Lexer) lexer.StateFunc {
	l.Ignore()
	l.Emit(TokenContainerBegin)
	return initialState
}

func braceRightState(l *lexer.Lexer) lexer.StateFunc {
	l.Ignore()
	l.Emit(TokenContainerEnd)
	return initialState
}

func valueState(l *lexer.Lexer) lexer.StateFunc {
	r := l.IgnoreUpTo(func(r rune) bool {
		return tNonWhitespace(r)
	})
	// If the next rune is NL, '{', or ';', presume empty Value
	nr := l.Peek()
	if tNewline(nr) || tBraceLeft(nr) || tComment(nr) {
		return initialState
	}
	// If it is a quote, then
	if tQuote(r) {
		l.Ignore()
		return quotedValueState
	}
	l.NextUpTo(func(r rune) bool {
		return tNewline(r) || tComment(r) || tBraceLeft(r)
	})
	l.Emit(TokenValue)
	return initialState
}

func quotedValueState(l *lexer.Lexer) lexer.StateFunc {
	l.NextUpTo(func(r rune) bool {
		return tQuote(r)
	})
	l.Emit(TokenValue)
	l.Ignore()
	return initialState
}

func variableState(l *lexer.Lexer) lexer.StateFunc {
	r := l.NextUpTo(func(r rune) bool {
		return tWhitespace(r) || tComment(r)
	})
	switch {
	case tComment(r):
		l.Emit(TokenVariable)
		return commentState
	case tNewline(r):
		l.Emit(TokenVariable)
		return initialState
	case tWhitespace(r):
		l.Emit(TokenVariable)
		return valueState
	}
	return initialState
}

func quotedVariableState(l *lexer.Lexer) lexer.StateFunc {
	l.Ignore()
	r := l.NextUpTo(func(r rune) bool {
		return tQuote(r) || tNewline(r)
	})
	switch {
	case tQuote(r):
		l.Emit(TokenVariable)
		l.Ignore()
		return valueState
	case tNewline(r):
		l.Emit(TokenVariable)
		return initialState
	}
	return initialState
}

func newlineState(l *lexer.Lexer) lexer.StateFunc {
	l.Ignore()
	return initialState
}

// type detection

func tWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

func tNonWhitespace(r rune) bool {
	return !tWhitespace(r)
}

func tAlphanumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func tNonAlphanumeric(r rune) bool {
	return !tAlphanumeric(r)
}

func tComment(r rune) bool {
	return r == ';'
}

func tNewline(r rune) bool {
	return r == '\n' || r == '\r'
}
func tQuote(r rune) bool {
	return r == '"'
}
func tBraceLeft(r rune) bool {
	return r == '{'
}
func tBraceRight(r rune) bool {
	return r == '}'
}
