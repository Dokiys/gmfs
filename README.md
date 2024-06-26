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
  -f string
	  Specifies files(separate multiple values with spaces).
  -v,--version		Show version info and exit.
```

```bash
echo '// Item Comment 1
/*
        Item Comment 1
*/
// Item Comment 1
type Item struct {
        // Item ItemId Comment 3

        // Item ItemId Comment 2
        ItemId    int // Item ItemId Comment 1
        Name      string
        Duration  time.Duration
        CreatedAt time.Time
}' | gmfs

// Item Comment 1
/*
	Item Comment 1
*/
// Item Comment 1
message Item {
	// Item ItemId Comment 1
	int64 item_id = 1;

	string name = 2;

	Duration duration = 3;

	google.protobuf.Timestamp created_at = 4;
}
```

If under macOS copy struct name and run `gmfs -r="UserInfo" -f="user.go" | pbcopy` at go file path will copy the result to clipboard.
