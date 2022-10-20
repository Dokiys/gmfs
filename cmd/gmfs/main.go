package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"syscall"

	"github.com/Dokiys/codemates/gmfs"
)

var s = flag.String("s", ".*", "Regexp match struct name.")
var typInt = flag.Int("i", 64, "Set int convert type, just allow [8,16,32,64].")

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, "usage: gmfs [OPTION] GO_FILES\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	checkArgs()

	exp, _ := regexp.Compile(*s)
	gmfs.TypInt = fmt.Sprintf("int%d", *typInt)

	for _, src := range flag.Args() {
		f, err := os.Open(src)
		if err != nil {
			if errors.Is(err, syscall.ENOENT) {
				continue
			}

			errExit(err)
		}

		w := os.Stdout
		if err := gmfs.GenMsg(f, w, *exp); err != nil {
			errExit(err)
		}
	}

	return
}

func checkArgs() {
	_, err := regexp.Compile(*s)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "-s invalid: %s\n", err)
		usage()
	}

	// int8 int16 int32 int64
	if *typInt%8 != 0 || *typInt/8 < 1 || *typInt/8 > 4 {
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
