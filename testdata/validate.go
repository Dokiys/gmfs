//go:build gmfs

package testdata

import (
	"time"
)

type DataValidate struct {
	None  bool
	Str   string     `json:"str" binding:"required"`
	Int   int        `json:"int" binding:"required,gte= 1"`
	Int32 int32      `json:"int32" binding:"required"`
	Arr   []int      `json:"arr,omitempty" binding:"required"`
	A     time.Month `json:"a" binding:"required"`
	Time  *time.Time `json:"time" binding:"required"`
	Enum  int        `json:"enum" binding:"required,oneof=1 2 3"`
	Msg   Item       `json:"msg" binding:"required"`
	MsgX  *Item      `json:"msgX" binding:"required"`
}
