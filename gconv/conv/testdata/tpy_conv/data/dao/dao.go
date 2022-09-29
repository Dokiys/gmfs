package dao

import "time"

type Item struct {
	Id        int
	Name      string
	CreatedAt time.Time
}

type Data struct {
	Id        int
	Name      string
	Item      Item
	Itemp     *Item
	Items     []Item
	Itemsp    []*Item
	CreatedAt time.Time
}
