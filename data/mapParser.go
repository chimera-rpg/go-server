package data

import (
	"log"
	"strconv"
	"strings"

	"github.com/eczarny/lexer"
)

type mapParser struct {
	stringsMap   *StringsMap
	lexer        *lexer.Lexer
	currentToken lexer.Token
}

func (p *mapParser) parse() map[string]Map {
	maps := make(map[string]Map)
Loop:
	for {
		switch p.nextToken().Type {
		case TokenVariable:
			maps[p.tokenValue()] = p.parseMap(p.tokenValue())
		case TokenEOF:
			log.Print("Finished reading map!")
			break Loop
		default:
			panic("Did not find initial Map declaration!")
		}
	}
	return maps
}

func (p *mapParser) parseMap(name string) Map {
	newMap := Map{DataName: name, Height: 1}
	newMap.Tiles = make([][][][]Archetype, 0, 100)
	p.expectToken(TokenContainerBegin, "Expected '{' after Map declaration.")
	p.nextToken()
Loop:
	for {
		switch p.currentToken.Type {
		case TokenVariable:
			p.parseMapVariable(&newMap, p.tokenValue())
		case TokenContainerEnd:
			log.Print("leaving map parse")
			p.nextToken()
			break Loop
		case TokenEOF:
			log.Print("End of Map without closing '}'!")
			break Loop
		default:
			log.Print("unrecognized token, skipping")
			p.nextToken()
		}
	}
	return newMap
}

func (p *mapParser) parseMapVariable(newMap *Map, name string) {
	switch name {
	case "Name":
		p.expectToken(TokenValue, "Expected string after Name.")
		newMap.Name = p.tokenValue()
		p.nextToken()
	case "Description":
		p.expectToken(TokenValue, "Expected string after Description.")
		newMap.Description = p.tokenValue()
		p.nextToken()
	case "Lore":
		p.expectToken(TokenValue, "Expected string after Lore.")
		newMap.Lore = p.tokenValue()
		p.nextToken()
	case "Height":
		p.expectToken(TokenValue, "Expected number after Height.")
		newMap.Height, _ = strconv.Atoi(p.tokenValue())
		newMap.Tiles = make([][][][]Archetype, newMap.Height)
		for y := range newMap.Tiles {
			newMap.Tiles[y] = make([][][]Archetype, newMap.Width)
			for x := range newMap.Tiles[y] {
				newMap.Tiles[y][x] = make([][]Archetype, newMap.Depth)
			}
		}
		p.nextToken()
	case "Width":
		p.expectToken(TokenValue, "Expected number after Width.")
		newMap.Width, _ = strconv.Atoi(p.tokenValue())
		newMap.Tiles = make([][][][]Archetype, newMap.Height)
		for y := range newMap.Tiles {
			newMap.Tiles[y] = make([][][]Archetype, newMap.Width)
			for x := range newMap.Tiles[y] {
				newMap.Tiles[y][x] = make([][]Archetype, newMap.Depth)
			}
		}
		p.nextToken()
	case "Depth":
		p.expectToken(TokenValue, "Expected number after Depth.")
		newMap.Depth, _ = strconv.Atoi(p.tokenValue())
		newMap.Tiles = make([][][][]Archetype, newMap.Height)
		for y := range newMap.Tiles {
			newMap.Tiles[y] = make([][][]Archetype, newMap.Width)
			for x := range newMap.Tiles[y] {
				newMap.Tiles[y][x] = make([][]Archetype, newMap.Depth)
			}
		}
		p.nextToken()
	case "Darkness":
		p.expectToken(TokenValue, "Expected number after Darkness.")
		newMap.Darkness, _ = strconv.Atoi(p.tokenValue())
		p.nextToken()
	case "ResetTime":
		p.expectToken(TokenValue, "Expected number after ResetTime.")
		newMap.ResetTime, _ = strconv.Atoi(p.tokenValue())
		p.nextToken()
	case "East":
		p.expectToken(TokenValue, "Expected string after East")
		newMap.EastMap = p.tokenValue()
		p.nextToken()
	case "West":
		p.expectToken(TokenValue, "Expected string after West")
		newMap.WestMap = p.tokenValue()
		p.nextToken()
	case "South":
		p.expectToken(TokenValue, "Expected string after South")
		newMap.SouthMap = p.tokenValue()
		p.nextToken()
	case "North":
		p.expectToken(TokenValue, "Expected string after North")
		newMap.NorthMap = p.tokenValue()
		p.nextToken()
	case "Tiles":
		p.expectToken(TokenContainerBegin, "Expected '{' after Tiles.")
		p.parseMapTiles(newMap)
	default:
		p.nextToken()
		log.Printf("Property '%s' in %s is unknown.\n", p.tokenValue(), name)
	}
}

func (p *mapParser) parseMapTiles(newMap *Map) {
Loop:
	for {
		switch p.currentToken.Type {
		case TokenVariable:
			coords := p.tokenValue()
			p.expectToken(TokenContainerBegin, "Expected '{' after Tile.")
			p.parseMapTile(newMap, coords)
		case TokenContainerEnd:
			break Loop
		case TokenEOF:
			log.Print("End of Map Tiles without closing '}'!")
			break Loop
		default:
			p.nextToken()
		}
	}
}

func (p *mapParser) parseMapTile(newMap *Map, coords string) {
	coordsSlice := strings.Split(coords, "x")
	x := 0
	y := 0
	z := 0
	if len(coordsSlice) == 2 {
		x, _ = strconv.Atoi(coordsSlice[0])
		z, _ = strconv.Atoi(coordsSlice[1])
	} else if len(coordsSlice) == 3 {
		y, _ = strconv.Atoi(coordsSlice[0])
		x, _ = strconv.Atoi(coordsSlice[1])
		z, _ = strconv.Atoi(coordsSlice[2])
	} else {
		log.Print("Incorrect Tile coordinates format, expected YxXxZ or XxZ")
	}
	newMap.Tiles[y][x][z] = make([]Archetype, 0, 0)
Loop:
	for {
		p.nextToken()
		switch p.currentToken.Type {
		case TokenVariable:
			p.parseMapTileVariable(&newMap.Tiles[y][x][z], p.tokenValue())
		case TokenContainerEnd:
			p.nextToken()
			break Loop
		case TokenEOF:
			log.Print("End of Map Tiles without closing '}'!")
			break Loop
		}
	}
}

func (p *mapParser) parseMapTileVariable(tileStack *[]Archetype, variable string) {
	switch variable {
	case "Arch":
		p.expectToken(TokenValue, "Expected string after Arch")
		arch := Archetype{}
		arch.ArchID = p.stringsMap.Acquire(p.tokenValue())
		*tileStack = append(*tileStack, arch)
	default:
		p.nextToken()
		log.Printf("Unrecognized Map Tile property '%s'\n", variable)
	}
}

//

func (p *mapParser) tokenValue() string {
	return p.currentToken.Value.(string)
}

func (p *mapParser) nextToken() lexer.Token {
	p.currentToken = p.lexer.NextToken()
	return p.currentToken
}

func (p *mapParser) acceptToken(tokenType lexer.TokenType) bool {
	return p.nextToken().Type == tokenType
}

func (p *mapParser) expectToken(tokenType lexer.TokenType, v interface{}) string {
	if !p.acceptToken(tokenType) {
		panic(v)
	}
	return p.tokenValue()
}
