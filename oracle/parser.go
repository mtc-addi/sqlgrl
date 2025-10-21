package oracle

import (
	"fmt"
	"slices"
	"strings"
	"tsqlgrl/generic"
)

type Parser struct {
	p      *generic.Parser
	Tables *generic.TablesDef
}

func NewParser() *Parser {

	result := &Parser{
		Tables: &generic.TablesDef{
			Origin: generic.DbOrigin{
				Vendor: generic.EngineVendorInfo{
					Name: "oracle",
				},
				Engine: generic.EngineInfo{
					Name:    "oracle",
					Version: "?",
				},
				Dialect:     "oracle",
				Description: "",
			},
			Tables: map[string]*generic.TableDef{},
		},
	}
	return result
}

func (p *Parser) Parse(tokens generic.Tokens) error {
	ignoreTypes := []string{"whitespace", "comment", "newline"}

	tokens = slices.DeleteFunc(tokens, func(t *generic.Token) bool {
		return slices.Contains(ignoreTypes, t.Type)
	})

	p.p = generic.NewParser(tokens)

	tlen := len(tokens)

	for range tlen {

		avail := p.p.Available()
		if avail < 1 {
			break
		}

		ok, err := p.p.Expect("keyword", "CREATE")
		if err != nil {
			return generic.Errorf(err, "failed to expect CREATE")
		}
		if ok {
			err = p.ParseCreate()
			if err != nil {
				return generic.Errorf(err, "failed to parse CREATE statement: %s", p.p.DebugLocation())
			}
			continue
		}

		//handle 'GRANT' statement
		ok, err = p.p.Expect("keyword", "GRANT")
		if err != nil {
			return generic.Errorf(err, "failed to expect GRANT")
		}
		if ok {
			toks, err := p.p.ConsumeUntil(TT_SYMBOL, ";", 256)
			if err != nil {
				return generic.Errorf(err, "failed to parse GRANT statement: %s", p.p.DebugLocation())
			}
			p.p.AdvanceLen(toks)

			ok, err = p.p.Expect(TT_SYMBOL, ";")
			if err != nil {
				return generic.Errorf(err, "failed to parse GRANT closing semicolon")
			}
			if !ok {
				return fmt.Errorf("expected GRANT closing semicolon, got: %q", p.p.Get().Content)
			}
			p.p.Advance()

			continue
		}

		//handle 'COMMENT' statement
		ok, err = p.p.Expect("keyword", "COMMENT")
		if err != nil {
			return generic.Errorf(err, "failed to expect COMMENT")
		}
		if ok {
			toks, err := p.p.ConsumeUntil(TT_SYMBOL, ";", 256)
			if err != nil {
				return generic.Errorf(err, "failed to parse COMMENT statement: %s", p.p.DebugLocation())
			}
			p.p.AdvanceLen(toks)

			ok, err = p.p.Expect(TT_SYMBOL, ";")
			if err != nil {
				return generic.Errorf(err, "failed to parse COMMENT closing semicolon")
			}
			if !ok {
				return fmt.Errorf("expected COMMENT closing semicolon, got: %q", p.p.Get().Content)
			}
			p.p.Advance()

			continue
		}

		return fmt.Errorf("unexpected unhandled token: %s", p.p.DebugLocation())
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

	tableName, err := p.ParseTableName()
	if err != nil {
		return generic.Errorf(err, "failed to parse table name")
	}

	t := &generic.TableDef{
		Columns: generic.ColumnsDef{},
	}
	p.Tables.Tables[tableName] = t

	err = p.ParseColumns(t)
	if err != nil {
		return generic.Errorf(err, "failed to parse table params")
	}

	toks, err := p.p.ConsumeUntil(TT_SYMBOL, ";", 256)
	if err != nil {
		return generic.Errorf(err, "failed to skip table commands")
	}
	p.p.AdvanceLen(toks)

	ok, err := p.p.Expect(TT_SYMBOL, ";")
	if err != nil {
		return generic.Errorf(err, "failed to read table closing semicolon")
	}
	if !ok {
		return fmt.Errorf("expected table closing semicolon, got: %q", p.p.Get().Content)
	}
	p.p.Advance()

	return nil
}

func (p *Parser) ParseTableName() (string, error) {
	ok, err := p.p.ExpectType(TT_STRING)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("expected table name 'string' token")
	}
	name := p.p.Get().Content
	p.p.Advance()

	//optional . followed by subname
	ok, err = p.p.Expect(TT_SYMBOL, ".")
	if err != nil {
		return "", err
	}
	if !ok {
		return name, nil
	}
	p.p.Advance()

	// name split by . aka "parent"."child"
	ok, err = p.p.ExpectType(TT_STRING)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("expected table sub-name 'string' token")
	}
	name = fmt.Sprintf("%s.%s", name, p.p.Get().Content)
	p.p.Advance()
	return name, nil
}

func (p *Parser) ParseColumns(t *generic.TableDef) error {
	ok, err := p.p.Expect(TT_SYMBOL, "(")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected table params open parenthesis token")
	}
	p.p.Advance()

	hasNext := true
	for hasNext {
		err = p.ParseColumn(t)
		if err != nil {
			return generic.Errorf(err, "failed to parse table param")
		}
		hasNext, err = p.p.Expect(TT_SYMBOL, ",")
		if err != nil {
			return generic.Errorf(err, "failed while checking for table param trailing comma")
		}
		if hasNext {
			p.p.Advance()
		}
	}

	ok, err = p.p.Expect(TT_SYMBOL, ")")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected table params close parenthesis token but got: %s", p.p.Get().Content)
	}
	p.p.Advance()

	return nil
}

func (p *Parser) ParseColumn(t *generic.TableDef) error {

	ok, err := p.p.ExpectType(TT_STRING)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected string token for column name")
	}
	columnName := p.p.Get().Content
	p.p.Advance()

	colDef := &generic.ColumnDef{}
	t.Columns[columnName] = colDef
	err = p.ParseColumnType(colDef)
	if err != nil {
		return generic.Errorf(err, "failed to parse column type")
	}

	ok, err = p.p.Expect(TT_KEYWORD, "DEFAULT")
	if err != nil {
		return generic.Errorf(err, "failed to check for DEFAULT keyword after column type")
	}
	if !ok {
		return nil
	}
	p.p.Advance()

	//todo check for other types too, not just string
	ok, err = p.p.ExpectTypeAny([]string{TT_STRING, "int", "float", TT_KEYWORD})
	// ok, err = p.p.ExpectType("string")
	if err != nil {
		return generic.Errorf(err, "failed to read DEFAULT value after column type")
	}
	if !ok {
		return fmt.Errorf("unexpected token after DEFAULT keyword")
	}
	p.p.Advance()

	return nil
}

func (p *Parser) ParseColumnType(d *generic.ColumnDef) error {
	ok, err := p.p.ExpectType(TT_KEYWORD)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected keyword token for column value type")
	}
	columnPrimitiveType := strings.ToUpper(p.p.Get().Content)
	p.p.Advance()
	d.Type = columnPrimitiveType

	//optional params in parenthesis
	ok, err = p.p.Expect(TT_SYMBOL, "(")
	if err != nil {
		return generic.Errorf(err, "failed to check for column type params open parenthesis")
	}
	if !ok {
		return nil
	}

	ts, err := p.p.ConsumeUntil(TT_SYMBOL, ")", 12)
	if err != nil {
		return generic.Errorf(err, "failed to scan column type params until close parenthesis, scanned %d", len(ts))
	}
	p.p.AdvanceLen(ts)

	ok, err = p.p.Expect(TT_SYMBOL, ")")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected closing parenthesis for column type params")
	}
	p.p.Advance()
	return nil
}
