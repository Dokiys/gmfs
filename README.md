# gmfs

Generate protobuf message from go struct.

## Installing

```bash
go install github.com/Dokiys/gmfs/cmd/gmfs@latest
```

## Usage

```bash
$ gmfs -h
usage: gmfs [OPTION] [GO_FILES]
  -i int
	  Set int convert type, only allow [32,64]. (default 64)
  -r string
	  Regexp match struct name. (default ".*")
  -v,--version		Show version info and exit.
```
If under macOS copy struct name and run `gmfs -s=$(pbpaste) $(ls | grep ".go") | pbcopy` at go file path will copy the result to clipboard.

![aa](gmfs.gif)
