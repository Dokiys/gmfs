package gmfs_test

import (
	"bytes"
	"os"
	"regexp"

	"github.com/Dokiys/gmfs"
)

func ExampleGenMsg() {
	src := `
package example

import (
	"strings"
	"time"
)

func (i *Item) P() {}

// Item Comment 1
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
}

type TemplateData struct {
	Arr   []string
	Items []*Item
	Map1  map[string]*Item

	// Unsupported
	function func() bool
	strings.Reader
}
`
	r := bytes.NewReader([]byte(src))

	exp, _ := regexp.Compile(".*")
	_ = gmfs.GenMsg(r, os.Stdout, *exp)
	// Output:
	// // Item Comment 1
	// /*
	//	Item Comment 1
	// */
	// // Item Comment 1
	// message Item {
	//	// Item ItemId Comment 1
	//	int64 item_id = 1;
	//
	//	string name = 2;
	//
	//	Duration duration = 3;
	//
	//	google.protobuf.Timestamp created_at = 4;
	// }
	//
	// message TemplateData {
	//
	//	repeated string arr = 1;
	//
	//	repeated Item items = 2;
	//
	//	map<string,Item> map1 = 3;
	// 	// Unsupported
	//	// Unsupported field: function
	//
	//	// Unsupported field: strings.Reader
	// }
}
