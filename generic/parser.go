package generic

import (
	"io"
	"slices"
	"strings"
)

type Parser struct {
	tokens []*Token
	offset int
}

func NewParser(tokens []*Token) *Parser {
	result := &Parser{
		tokens: tokens,
		offset: 0,
	}

	return result
}

func (p *Parser) Available() int {
	return len(p.tokens) - p.offset
}

func (p *Parser) HasAvailable(n int) bool {
	return p.Available() >= n
}

func (p *Parser) HasAvailableEOF(n int) error {
	if !p.HasAvailable(n) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (p *Parser) Peak() (*Token, error) {
	err := p.HasAvailableEOF(1)
	if err != nil {
		return nil, err
	}
	return p.Get(), nil
}

func (p *Parser) PeakN(n int) (Tokens, error) {
	err := p.HasAvailableEOF(n)
	if err != nil {
		return nil, err
	}
	return p.tokens[p.offset:n], nil
}

func (p *Parser) ExpectType(t string) (bool, error) {
	tok, err := p.Peak()
	if err != nil {
		return false, err
	}
	return tok.Type == t, nil
}

func (p *Parser) ExpectTypeAny(ts []string) (bool, error) {
	tok, err := p.Peak()
	if err != nil {
		return false, err
	}
	return slices.Contains(ts, tok.Type), nil
}

func (p *Parser) ExpectCase(t string, content string, ignoreCase bool) (bool, error) {
	tok, err := p.Peak()
	if err != nil {
		return false, err
	}
	a := tok.Content
	b := content
	if ignoreCase {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
	}
	return tok.Type == t && a == b, nil
}

func (p *Parser) Expect(t string, content string) (bool, error) {
	return p.ExpectCase(t, content, true)
}

func (p *Parser) AdvanceN(n int) {
	p.offset += n
}

func (p *Parser) Advance() {
	p.offset++
}

func (p *Parser) Get() *Token {
	return p.tokens[p.offset]
}

func (p *Parser) Ignore(t string) bool {
	ok, err := p.ExpectType(t)
	if err == nil && ok {
		p.Advance()
		return true
	}
	return false
}

func (p *Parser) IgnoreRepeated(t string) int {
	counter := 0
	tlen := len(p.tokens)
	for i := 0; i < tlen; i++ {
		if !p.Ignore(t) {
			break
		}
		counter++
	}
	return counter
}
