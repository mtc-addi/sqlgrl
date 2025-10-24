package main

import (
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"tsqlgrl/oracle"
)

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

	res, err := oracle.ParseReader(fpath, f)
	if err != nil {
		return err
	}
	bs, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		return err
	}
	str := string(bs)
	log.Println(str)
	// log.Println(res)

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
