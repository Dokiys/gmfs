package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"syscall"

	"github.com/Dokiys/gmfs"
)

const VERSION = "v0.3.0"

var r string
var intType int
var filename string

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(VERSION)
		return
	}

	flag.Usage = usage
	flag.IntVar(&intType, "i", 64, "Set int convert type, only allow [32,64].")
	flag.StringVar(&r, "r", ".*", "Regexp match struct name.")
	flag.StringVar(&filename, "f", "", "Specifies files(separate multiple values with spaces).")
	flag.Parse()

	checkArgs()

	exp, _ := regexp.Compile(r)
	gmfs.IntType = fmt.Sprintf("int%d", intType)

	// 从标准输入执行
	if strings.TrimSpace(filename) == "" {
		reader := io.MultiReader(bytes.NewReader([]byte("package main\n")), os.Stdin)
		if err := gmfs.GenMsg(reader, os.Stdout, *exp); err != nil {
			errExit(err)
		}

		return
	}

	// 从指定文件执行
	for _, src := range strings.Fields(filename) {
		if strings.TrimSpace(src) == "" {
			continue
		}
		f, err := os.Open(src)
		if err != nil {
			if errors.Is(err, syscall.ENOENT) {
				continue
			}

			errExit(err)
		}

		if err := gmfs.GenMsg(f, os.Stdout, *exp); err != nil {
			errExit(err)
		}
	}
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, `usage: gmfs [OPTION] [GO_FILES]
  -i int
	  Set int convert type, only allow [32,64]. (default 64)
  -r string
	  Regexp match struct name. (default ".*")
  -f string
	  Specifies files(separate multiple values with spaces).
  -v,--version		Show version info and exit.
`)
	os.Exit(2)
}

func checkArgs() {
	_, err := regexp.Compile(r)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "-r invalid: %r\n", err)
		usage()
	}

	// int32 int64
	if intType != 32 && intType != 64 {
		usage()
	}
}

func errExit(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
