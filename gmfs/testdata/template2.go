//go:build gmfs

package testdata

import (
	"time"
)

type AA *time.Time

// req
type Request struct {
	Name      string    `json:"name" binding:"required"`                                         // 名称
	Type      string    `json:"type" binding:"required"`                                         // 类型
	StartTime time.Time `json:"start_time" binding:"required" time_format:"2006-01-02 15:04:05"` // 开始时间
	EndTime   time.Time `json:"end_time" binding:"required" time_format:"2006-01-02 15:04:05"`   // 结束时间
	Ids       []int     `json:"ids,omitempty"`
	StrArr    [][]string
	MapArr    map[string][]Item
	aaa       func() bool
	Condition
	*a.Condition
}
