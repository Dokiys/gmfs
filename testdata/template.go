//go:build gmfs

package testdata

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
