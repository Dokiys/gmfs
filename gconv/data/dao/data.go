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
	Item      []*Item
	CreatedAt time.Time
}
