# gmfs

Generate protobuf message from go struct.

## Installing

```bash
go install github.com/Dokiys/codemates/cmd/gmfs@latest
```

## Usage

```bash
$ gmfs -h
usage: gstm [-s] GO_FILES
  -s string
    	Regexp match struct name. (default ".*")
```

```bash
$ cat example.go 
package testdata

import "time"

func P() {}

// Item Comment
type Item struct {
        ItemId    int // ItemId Comment
        Name      string
        Duration  time.Duration
        CreatedAt time.Time
}

type TemplateData struct {
        Arr   []string
        Items []*Item
        Map   map[string]*Item
}

$ gmfs example.go
// Item Comment
message Item {
        // ItemId Comment
        int64 item_id = 1;

        string name = 2;

        Duration duration = 3;

        google.protobuf.Timestamp created_at = 4;
}

message TemplateData {

        repeated string arr = 1;

        repeated Item items = 2;

        map<string,Item> map = 3;
}%
```
