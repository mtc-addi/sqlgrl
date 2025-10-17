package oracle

import (
	"fmt"
	"log"
	"tsqlgrl/generic"
)

type Parser struct {
	p      *generic.Parser
	Tables *generic.TablesDef
}

func NewParser() *Parser {

	result := &Parser{
		Tables: &generic.TablesDef{},
	}
	return result
}

func (p *Parser) Parse(tokens generic.Tokens) error {
	p.p = generic.NewParser(tokens)

	tlen := len(tokens)

	for i := 0; i < tlen; i++ {
		p.p.IgnoreRepeated("whitespace")
		p.p.IgnoreRepeated("comment")
		p.p.IgnoreRepeated("newline")

		ok, err := p.p.Expect("keyword", "CREATE")
		if err != nil {
			return generic.Errorf(err, "failed to expect CREATE")
		}
		if ok {
			err = p.ParseCreate()
			if err != nil {
				return err
			}
			continue
		}

		return fmt.Errorf("unexpected unhandled token %v", p.p.Get())
	}
	return nil
}

func (p *Parser) ParseCreate() error {
	p.p.Advance() //consume 'CREATE'

	ok, err := p.p.Expect("keyword", "TABLE")
	if err != nil {
		return err
	}
	if ok {
		return p.ParseTable()
	}

	return fmt.Errorf("unexpected unhandled token for 'CREATE' substep %v", p.p.Get())
}

func (p *Parser) ParseTable() error {
	p.p.Advance() //consume 'TABLE'

	ok, err := p.p.ExpectType("string")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected table name 'string' token")
	}
	name := p.p.Get().Content
	log.Println("table name", name)

	p.p.Advance()

	return nil
}
