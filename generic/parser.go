package generic

import (
	"fmt"
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
	return p.tokens[p.offset : p.offset+n], nil
}

func (p *Parser) PeakMaxN(n int) (Tokens, error) {
	_max := min(p.Available(), n)
	return p.PeakN(_max)
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

func (p *Parser) AdvanceLen(t Tokens) {
	p.offset += len(t)
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
	tlen := len(p.tokens[p.offset:])
	for i := 0; i < tlen; i++ {
		if !p.Ignore(t) {
			break
		}
		counter++
	}
	// log.Println("ignored", counter, t)
	return counter
}

func (p *Parser) ConsumeUntil(t string, content string, _max int) (Tokens, error) {
	_max = min(p.Available(), _max)
	searchSpace := p.tokens[p.offset : p.offset+_max]

	idx := slices.IndexFunc(searchSpace, func(tok *Token) bool {
		return tok.Type == t && tok.Content == content
	})

	if idx == -1 {
		return nil, fmt.Errorf("exhausted search space without finding %s %s", t, content)
	}
	return searchSpace[:idx], nil
}

func (p *Parser) IsTokenIndexValid(i int) bool {
	max := len(p.tokens) - 1
	min := 0
	return i >= min && i <= max
}

func (p *Parser) DebugLocation() string {
	var startStr string
	var contentStr string
	ts, err := p.PeakMaxN(8)
	tslen := len(ts)

	if err != nil {
		startStr = fmt.Sprintf("[failed to debug location due to: %s]", err.Error())
	} else {
		if tslen < 1 {
			startStr = "no available tokens for debug location info, past max"
		} else {
			start := ts[0]
			startStr = fmt.Sprintf("line %d", start.LineStart+1)
			contentStr = JoinTokensContent(ts, " ")
		}
	}

	return fmt.Sprintf("tokens at %d aka %s %q [..]", p.offset, startStr, contentStr)
}
