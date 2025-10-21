package oracle

import (
	"fmt"
	"io"
	"log"
	"tsqlgrl/generic"
)

const TT_SYMBOL string = "symbol"
const TT_STRING string = "string"
const TT_KEYWORD string = "keyword"

func TokenizeFile(r io.Reader) (generic.Tokens, error) {

	//ensure keywords sorted by length then alphabet
	generic.SortKeywords(KEYWORDS)

	//read the whole file at once into a string
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	src := string(buf)

	//run the lexer over the string content
	s := generic.NewLexer(src)
	srcLen := len(src)

	//loop until file is empty but not past max of srcLen (arbitrary anti-forever loop)
	for i := 0; i < srcLen; i++ {

		// only other valid option is EOF
		ok := s.MatchEOF()
		if ok {
			break
		}

		//scan for whitespace
		ok, err := s.MatchCharsetMulti(generic.CHARSET_WHITESPACE, "whitespace")
		if err != nil {
			return nil, generic.Errorf(err, "error while reading whitespace at %s", s.DebugLocation())
		}
		if ok {
			continue
		}

		//scan for whitespace
		ok, err = s.MatchCharsetMulti(generic.CHARSET_NEWLINE, "newline")
		if err != nil {
			return nil, generic.Errorf(err, "error while reading newline at %s", s.DebugLocation())
		}
		if ok {
			continue
		}

		//scan for comment lines
		ok, err = s.PeakMatchUntilCharset("--", generic.CHARSET_NEWLINE, "comment")
		if err != nil {
			return nil, generic.Errorf(err, "error while reading comment at %s", s.DebugLocation())
		}
		if ok {
			continue
		}

		//scan for string literals enclosed by double quotes
		ok, err = s.MatchEnclosed("\"", TT_STRING)
		if err != nil {
			return nil, generic.Errorf(err, "error while reading string at %s", s.DebugLocation())
		}
		if ok {
			continue
		}

		//scan for string literals enclosed by single quotes
		ok, err = s.MatchEnclosed("'", TT_STRING)
		if err != nil {
			return nil, generic.Errorf(err, "error while reading string at %s", s.DebugLocation())
		}
		if ok {
			continue
		}

		//scan for floating point (decimal separated)
		ok, err = s.MatchRegex(generic.REGEX_FLOAT, "float")
		if err != nil {
			return nil, generic.Errorf(err, "error while reading float at %s", s.DebugLocation())
		}
		if ok {
			continue
		}

		//scan for integer (no decimal)
		ok, err = s.MatchCharsetMulti(generic.CHARSET_INT, "int")
		if err != nil {
			return nil, generic.Errorf(err, "error while reading int at %s", s.DebugLocation())
		}
		if ok {
			continue
		}

		//scan for single char symbols
		ok, err = s.MatchCharset(".,();*", TT_SYMBOL)
		if err != nil {
			return nil, generic.Errorf(err, "error while reading symbol at %s", s.DebugLocation())
		}
		if ok {
			continue
		}

		//scan for comment lines
		ok, err = s.PeakMatchUntilCharset("@", generic.CHARSET_NEWLINE, "include")
		if err != nil {
			return nil, generic.Errorf(err, "error while reading comment at %s", s.DebugLocation())
		}
		if ok {
			continue
		}

		//scan for keywords
		ok, err = s.MatchAnyString(KEYWORDS, TT_KEYWORD)
		if err != nil {
			return nil, generic.Errorf(err, "error while reading keyword at %s", s.DebugLocation())
		}
		if ok {
			continue
		}

		err = fmt.Errorf("unhandled data at line %s", s.DebugLocation())
		log.Println(s.Output())
		return nil, err
	}

	return s.Output(), nil
}
