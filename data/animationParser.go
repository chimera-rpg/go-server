package data

import (
	"log"

	"github.com/eczarny/lexer"
)

type animationParser struct {
	stringsMap   *StringsMap
	lexer        *lexer.Lexer
	currentToken lexer.Token
}

func (p *animationParser) parse() map[FileID]Animation {
	animations := make(map[FileID]Animation)
Loop:
	for {
		switch p.nextToken().Type {
		case TokenVariable:
			animID := p.stringsMap.Acquire(p.tokenValue())
			animations[animID] = p.parseAnimation(p.tokenValue())
		case TokenEOF:
			log.Print("Finished reading animation!")
			break Loop
		default:
			panic("Did not find initial Animation declaration!")
		}
	}
	return animations
}

func (p *animationParser) parseAnimation(name string) Animation {
	newAnimation := Animation{
		AnimID: p.stringsMap.Acquire(name),
		Faces:  make(map[uint32][]AnimationFrame),
	}
	p.expectToken(TokenContainerBegin, "Expected '{' after Animation declaration.")
	p.nextToken()
Loop:
	for {
		switch p.currentToken.Type {
		case TokenVariable:
			p.parseAnimationVariable(&newAnimation, p.tokenValue())
		case TokenContainerEnd:
			p.nextToken()
			break Loop
		case TokenEOF:
			log.Print("End of Animation without closing '}'!")
			break Loop
		default:
			log.Print("unrecognized token, skipping")
			p.nextToken()
		}
	}
	return newAnimation
}

func (p *animationParser) parseAnimationVariable(newAnimation *Animation, name string) {
	switch name {
	case "Faces":
		p.expectToken(TokenContainerBegin, "Expected '{' after Faces.")
		p.parseAnimationFaces(newAnimation)
	default:
		p.nextToken()
		log.Printf("Property '%s' in %s is unknown.\n", p.tokenValue(), name)
	}
}

func (p *animationParser) parseAnimationFaces(newAnimation *Animation) {
Loop:
	for {
		switch p.currentToken.Type {
		case TokenVariable:
			faceset := p.tokenValue()
			p.expectToken(TokenContainerBegin, "Expected '{' after FaceSet.")
			p.parseAnimationFaceSet(newAnimation, faceset)
		case TokenContainerEnd:
			break Loop
		case TokenEOF:
			log.Print("End of Animation Faces without closing '}'!")
			break Loop
		default:
			p.nextToken()
		}
	}
}

func (p *animationParser) parseAnimationFaceSet(newAnimation *Animation, faceset string) {
	faceID := p.stringsMap.Acquire(faceset)
	newAnimation.Faces[faceID] = make([]AnimationFrame, 0, 0)
Loop:
	for {
		p.nextToken()
		switch p.currentToken.Type {
		case TokenVariable:
			newAnimation.Faces[faceID] = append(newAnimation.Faces[faceID], p.parseAnimationFaceSetFrame(p.tokenValue()))
		case TokenContainerEnd:
			log.Print("Done with end of faceset")
			p.nextToken()
			break Loop
		case TokenEOF:
			log.Print("End of Animation FaceSet without closing '}'!")
			break Loop
		default:
			log.Printf("Property '%s'%d in %s is unknown.\n", p.currentToken.Value, p.currentToken.Type, faceset)
		}
	}
}

func (p *animationParser) parseAnimationFaceSetFrame(framepath string) AnimationFrame {
	frametime := 100 // Default to 100ms

	frame := AnimationFrame{
		ImageID:   p.stringsMap.Acquire(framepath),
		FrameTime: frametime,
	}
	return frame
}

func (p *animationParser) tokenValue() string {
	return p.currentToken.Value.(string)
}

func (p *animationParser) nextToken() lexer.Token {
	p.currentToken = p.lexer.NextToken()
	return p.currentToken
}

func (p *animationParser) acceptToken(tokenType lexer.TokenType) bool {
	return p.nextToken().Type == tokenType
}

func (p *animationParser) expectToken(tokenType lexer.TokenType, v interface{}) string {
	if !p.acceptToken(tokenType) {
		panic(v)
	}
	return p.tokenValue()
}
