package main

import (
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"tsqlgrl/oracle"
)

var wsRegex regexp.Regexp = *regexp.MustCompile(`\s+`)

func IndexOfStartsWith(parts []string, start string) int {
	for i, part := range parts {
		if strings.HasPrefix(part, start) {
			return i
		}
	}
	return -1
}
func StrArrTrimAfter(parts []string, after string) []string {
	i := IndexOfStartsWith(parts, after)
	if i == -1 {
		return parts
	}
	return parts[:i]
}

func StrArrToLower(parts []string) {
	for i, part := range parts {
		parts[i] = strings.ToLower(part)
	}
}

func HandleFile(fpath string) error {
	log.Println(fpath)
	ext := strings.ToLower(path.Ext(fpath))
	if ext != ".sql" {
		return nil
	}
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	tokens, err := oracle.TokenizeFile(f)
	if err != nil {
		return err
	}

	// log.Println(tokens)
	p := oracle.NewParser()
	err = p.Parse(tokens)

	if err != nil {
		return err
	}

	return nil
}

func HandlePath(p string) error {
	fi, err := os.Stat(p)
	if err != nil {
		return err
	}

	if fi.IsDir() {

		return filepath.WalkDir(p, func(filePath string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			return HandleFile(filePath)
		})
	} else {
		return HandleFile(p)
	}
}

func main() {
	argsLen := len(os.Args)
	if argsLen < 2 {
		panic("not enough args, expected first arg to be a file or directory")
	}

	openPath := os.Args[1]

	err := HandlePath(openPath)
	if err != nil {
		panic(err)
	}

	log.Println("done")
}
