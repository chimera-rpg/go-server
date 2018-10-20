package parser

import (
  "github.com/eczarny/lexer"
)

type StateFunc func(*Parser) StateFunc

type Parser struct {
  lexer *lexer.Lexer
  initialState StateFunc
  CurrentToken lexer.Token
  PreviousToken lexer.Token
  PreviousPreviousToken lexer.Token
  Data interface{}
  CurrentData interface{}
  StoredData map[string]string
}

func NewParser(l *lexer.Lexer, initialState StateFunc, data interface{}) *Parser {
  p := &Parser{
    lexer: l,
    initialState: initialState,
    Data: data,
    StoredData: make(map[string]string, 0),
  }
  return p
}

func (p* Parser) Parse() {
  for s := p.initialState; s != nil; {
    s = s(p)
  }
}

func (p* Parser) TokenValue() string {
  return p.CurrentToken.Value.(string)
}

func (p *Parser) NextToken() lexer.Token {
  p.PreviousPreviousToken = p.PreviousToken
  p.PreviousToken = p.CurrentToken
  p.CurrentToken = p.lexer.NextToken()
  return p.CurrentToken
}

func (p *Parser) AcceptToken(tokenType lexer.TokenType) bool {
  return p.NextToken().Type == tokenType
}

func (p *Parser) ExpectToken(tokenType lexer.TokenType, v interface{}) string {
  if !p.AcceptToken(tokenType) {
    panic(v)
  }
  return p.TokenValue()
}

func (p *Parser) StoreData(key string, value string) {
  p.StoredData[key] = value
}
