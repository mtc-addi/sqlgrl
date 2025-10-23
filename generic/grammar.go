package generic

type Grammar struct {
	ignoredTypes map[string]bool
}

type TokenExpect struct {
	Type    string
	Content string
	SaveKey string
	Next    []*TokenExpect
}

func NewGrammar() *Grammar {
	result := &Grammar{
		ignoredTypes: map[string]bool{},
	}
	return result
}

func (g *Grammar) IgnoreTypes(_types ...string) *Grammar {
	for _, _type := range _types {
		g.ignoredTypes[_type] = true
	}
	return g
}

func (g *Grammar) Parse(p *Parser, output any) {

}

func (g *Grammar) TokenSave(_type string, _content string, _key string) *Grammar {

	return g
}
func (g *Grammar) Token(_type string, _content string) *Grammar {
	return g.TokenSave(_type, _content, "")
}

func test() {
	g := NewGrammar()

	g.IgnoreTypes("whitespace", "newline", "comment")

	g.Token("keyword", "CREATE").
		Token("keyword", "TABLE")
}
