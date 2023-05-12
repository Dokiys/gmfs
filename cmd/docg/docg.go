package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"syscall"

	"github.com/Dokiys/gmfs"
)

const VERSION = "v0.1.0"

var r string
var prefix string
var intType int

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(VERSION)
		return
	}

	flag.Usage = usage
	flag.IntVar(&intType, "i", 64, "Set int convert type, only allow [32,64].")
	flag.StringVar(&r, "r", ".*", "Regexp match struct name.")
	flag.StringVar(&prefix, "p", "", "Prefix string add to ech filed.")
	flag.Parse()

	checkArgs()

	exp, _ := regexp.Compile(r)
	gmfs.IntType = fmt.Sprintf("int%d", intType)

	for _, src := range flag.Args() {
		f, err := os.Open(src)
		if err != nil {
			if errors.Is(err, syscall.ENOENT) {
				continue
			}

			errExit(err)
		}

		if err := gmfs.GenMsg(f, os.Stdout, *exp, prefix); err != nil {
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

	if len(flag.Args()) <= 0 {
		usage()
	}
}

func errExit(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
