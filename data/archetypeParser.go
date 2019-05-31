package data

import (
	"log"

	"github.com/eczarny/lexer"
)

type archetypeParser struct {
	stringsMap   *StringsMap
	lexer        *lexer.Lexer
	currentToken lexer.Token
}

func (p *archetypeParser) parse() map[StringID]Archetype {
	archetypes := make(map[StringID]Archetype)
Loop:
	for {
		switch p.nextToken().Type {
		case TokenVariable:
			archID := p.stringsMap.Acquire(p.tokenValue())
			archetypes[archID] = p.parseArchetype(archID)
		case TokenEOF:
			break Loop
		default:
			panic("Did not find initial Archetype declaration!")
		}
	}
	return archetypes
}

func (p *archetypeParser) parseArchetype(id StringID) Archetype {
	archetype := NewArchetype()
	archetype.ArchID = id
	p.expectToken(TokenContainerBegin, "Expected '{' after Archetype declaration.")
	p.nextToken()
Loop:
	for {
		switch p.currentToken.Type {
		case TokenVariable:
			p.parseArchetypeVariable(&archetype, p.tokenValue())
		case TokenContainerEnd:
			log.Print("leaving Archetype parse")
			break Loop
		case TokenEOF:
			log.Print("End of Archetype without closing '}'!")
			break Loop
		default:
			p.nextToken()
		}
	}
	return archetype
}

func (p *archetypeParser) parseArchetypeVariable(archetype *Archetype, name string) {
	switch name {
	case "Anim":
		p.expectToken(TokenValue, "Expected string after Anim.")
		archetype.AnimID = p.stringsMap.Acquire(p.tokenValue())
		p.nextToken()
	case "Name":
		p.expectToken(TokenValue, "Expected string after Name.")
		archetype.setStructProperty(name, p.tokenValue())
		p.nextToken()
	case "Type":
		p.expectToken(TokenValue, "Expected string after Type.")
		archetype.setStructProperty(name, p.tokenValue())
		p.nextToken()
	case "Description":
		p.expectToken(TokenValue, "Expected string after Type.")
		archetype.setStructProperty(name, p.tokenValue())
		p.nextToken()
	default:
		p.nextToken()
		if p.currentToken.Type == TokenValue {
			archetype.setStructProperty(name, p.tokenValue())
			p.nextToken()
		} else if p.currentToken.Type == TokenVariable {
			archetype.addProperty(name, "true")
		}
	}
}

//

func (p *archetypeParser) tokenValue() string {
	return p.currentToken.Value.(string)
}

func (p *archetypeParser) nextToken() lexer.Token {
	p.currentToken = p.lexer.NextToken()
	return p.currentToken
}

func (p *archetypeParser) acceptToken(tokenType lexer.TokenType) bool {
	return p.nextToken().Type == tokenType
}

func (p *archetypeParser) expectToken(tokenType lexer.TokenType, v interface{}) string {
	if !p.acceptToken(tokenType) {
		panic(v)
	}
	return p.tokenValue()
}
