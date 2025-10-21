package generic

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

var CHARSET_WHITESPACE string = "\t "
var CHARSET_NEWLINE string = "\n\r"
var CHARSET_INT = "0123456789"
var REGEX_FLOAT regexp.Regexp = *regexp.MustCompile(`^[0-9]+\.[0-9]+`)

/** Sorts by longer keywords first
 * Then alphabet after
	* fixes early out for cases where "T" would appear before "TABLE"
*/
func SortKeywords(keywords []string) {
	sort.Slice(keywords, func(i, j int) bool {
		if len(keywords[i]) == len(keywords[j]) {
			return keywords[i] < keywords[j] // secondary alphabetical sort
		}
		return len(keywords[i]) > len(keywords[j]) // longer first
	})
}

type Token struct {
	Content   string
	Type      string
	Start     int
	End       int
	LineStart int
	LineEnd   int
}

func (t Token) String() string {
	return fmt.Sprintf("{ %s %q }\n", t.Type, t.Content)
}

type Tokens = []*Token

func JoinTokensContent(ts []*Token, delim string) string {
	results := []string{}
	for _, t := range ts {
		results = append(results, t.Content)
	}
	return strings.Join(results, delim)
}

type Lexer struct {
	src         string
	offset      int
	LineCounter int
	tokens      Tokens
}

func NewLexer(src string) *Lexer {
	result := &Lexer{
		src:         src,
		offset:      0,
		LineCounter: 0,
	}
	return result
}

func (s *Lexer) Available() int {
	return len(s.src) - s.offset
}

func (s *Lexer) HasAvailable(n int) bool {
	return s.Available()-n >= 0
}

func (s *Lexer) HasAvailableEOF(n int) error {
	if !s.HasAvailable(n) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (s *Lexer) PeakN(n int) (string, error) {
	err := s.HasAvailableEOF(n)
	if err != nil {
		return "", err
	}
	return s.src[s.offset : s.offset+n], nil
}

/**Returns string where max size is min of n or available*/
func (s *Lexer) PeakMaxN(n int) string {
	_min := min(n, s.Available())
	return s.src[s.offset : s.offset+_min]
}

func (s *Lexer) HasStringCase(m string, matchCase bool) (bool, error) {
	l := len(m)
	str, err := s.PeakN(l)
	if err != nil {
		return false, err
	}
	if !matchCase {
		str = strings.ToLower(str)
		m = strings.ToLower(m)
	}
	return str == m, nil
}

/**Ignores case*/
func (s *Lexer) HasString(m string) (bool, error) {
	return s.HasStringCase(m, false)
}

func CountNewLines(s string) int {
	return strings.Count(s, "\n")
}

/**Ignores case*/
func (s *Lexer) MatchString(m string, outputType string) (bool, error) {
	return s.MatchStringCase(m, outputType, false)
}

func (s *Lexer) MatchStringCase(m string, outputType string, matchCase bool) (bool, error) {
	has, err := s.HasStringCase(m, matchCase)
	if err != nil {
		return false, err
	}
	if has {
		s.TokenFromAdvance(len(m), outputType)
	}
	return has, nil
}

/* Tries to append a token that matches one char in chrs
 */
func (s *Lexer) MatchCharset(chrs string, outputType string) (bool, error) {
	sub, err := s.PeakN(1)
	if err != nil {
		return false, err
	}

	idx := strings.Index(chrs, sub)
	if idx == -1 {
		return false, nil
	}

	s.TokenFromAdvance(1, outputType)
	return true, nil
}

/* Tries to append a token that matches any string in ms
 * keywords that extend past available read space generate errors, but they are handled within this function
 */
func (s *Lexer) MatchAnyStringCase(ms []string, outputType string, matchCase bool) (bool, error) {
	for _, m := range ms {
		match, err := s.MatchStringCase(m, outputType, matchCase)
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				continue
			} else {
				return false, err
			}
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

/* Tries to append a token that matches any string in ms
 * Ignores case
 */
func (s *Lexer) MatchAnyString(ms []string, outputType string) (bool, error) {
	return s.MatchAnyStringCase(ms, outputType, false)
}

/* Creates a token from next n chars
 * does not append to lexer.tokens but returns instead
	* does not advance read offset or line counter
*/
func (s *Lexer) TokenFrom(n int, outputType string) *Token {
	Start := s.offset
	Count := n
	End := Start + Count

	Content := s.src[Start:End]

	LineStart := s.LineCounter
	LineCount := CountNewLines(Content)
	LineEnd := LineStart + LineCount

	result := &Token{
		Content:   Content,
		Start:     Start,
		End:       End,
		LineStart: LineStart,
		LineEnd:   LineEnd,
		Type:      outputType,
	}
	return result
}

/* Creates a token given n chars to read
 * automatically appends to lexer.tokens and advances read offset, line counter
 */
func (s *Lexer) TokenFromAdvance(n int, outputType string) *Token {
	t := s.TokenFrom(n, outputType)
	s.tokens = append(s.tokens, t)
	s.offset = t.End
	s.LineCounter = t.LineEnd
	return t
}

/* Tries to append a token that ends with a char from suffixCharset
 * originally implemented to fulfill --comment\n\r but didn't include prefix step
 */
func (s *Lexer) MatchUntilCharset(suffixCharset string, outputType string) (bool, error) {
	sub := s.src[s.offset:]
	subOffset := strings.IndexAny(sub, suffixCharset)
	if subOffset == -1 {
		return false, Errorf(io.ErrUnexpectedEOF, "index == -1 aka EOF")
	}
	// srcOffset := s.offset + subOffset
	// Content = sub[:srcOffset]
	Count := subOffset
	s.TokenFromAdvance(Count, outputType)
	return true, nil
}

/*
 * Tries to append a token that has a prefix and ends with a char from suffixCharset
	* Implemented to fulfill --some comment\n\r
*/
func (s *Lexer) PeakMatchUntilCharset(prefix string, suffixCharset string, outputType string) (bool, error) {
	str, err := s.PeakN(len(prefix))
	if err != nil {
		return false, err
	}
	if str != prefix {
		return false, nil
	}
	return s.MatchUntilCharset(suffixCharset, outputType)
}

func (s *Lexer) MatchEnclosed(edge string, outputType string) (bool, error) {
	return s.MatchEnclosedIncludeEdge(edge, outputType, false)
}

/*
 * Tries to append a token that begins and ends with 'edge' string
 * Implemented to fulfill "some string 'blah blah' blah 123"
 */
func (s *Lexer) MatchEnclosedIncludeEdge(edge string, outputType string, includeEdge bool) (bool, error) {
	edgeLen := len(edge)

	str, err := s.PeakN(edgeLen)
	if err != nil {
		return false, err
	}
	if str != edge {
		return false, nil
	}
	// srcLen := len(s.src)

	Start := s.offset
	SearchStart := Start + edgeLen
	SearchSub := s.src[SearchStart:]

	SearchEnd := strings.Index(SearchSub, edge)

	if SearchEnd == -1 {
		return false, io.ErrUnexpectedEOF
	}

	End := SearchStart + SearchEnd + edgeLen

	Count := End - Start

	t := s.TokenFromAdvance(Count, outputType)

	if !includeEdge {
		t.Content = strings.TrimSuffix(strings.TrimPrefix(t.Content, edge), edge)
	}

	return true, nil
}

/* Tries to append a token that matches characters in a regex
 * It is assumed that the regex will enforce beginning with ^ operator
	* global and multiline are NOT assumed
*/
func (s *Lexer) MatchRegex(re regexp.Regexp, outputType string) (bool, error) {
	sub := s.src[s.offset:]
	subBytes := []byte(sub)
	findBytes := re.Find(subBytes)
	if findBytes == nil {
		return false, nil
	}
	find := string(findBytes)
	findLen := len(find)
	s.TokenFromAdvance(findLen, outputType)
	return true, nil
}

/*Tries to append a token that matches characters in chrs*/
func (s *Lexer) MatchCharsetMulti(chrs string, outputType string) (bool, error) {
	srcLen := len(s.src)

	count := 0

	for i := s.offset; i < srcLen; i++ {
		if !strings.ContainsRune(chrs, rune(s.src[i])) {
			break
		}
		count++
	}

	if count == 0 {
		return false, nil
	}

	s.TokenFromAdvance(count, outputType)
	return true, nil
}

/*Helper function that returns true when available < 1*/
func (s *Lexer) MatchEOF() bool {
	return s.Available() < 1
}

func (s *Lexer) Output() Tokens {
	return s.tokens
}

func (s *Lexer) DebugLocation() string {
	return fmt.Sprintf("line %d %q [..]", s.LineCounter+1, s.PeakMaxN(64))
}
