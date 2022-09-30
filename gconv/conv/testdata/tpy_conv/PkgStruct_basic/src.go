package tpyconv

import (
	"gconv/conv/testdata/tpy_conv/PkgStruct_basic/data"
)

type Aa struct {
	PP1 data.Data
	PP2 *data.Data
}

type Bb struct {
	PP1 *data.Data
	PP2 data.Data
}

type A struct {
	P1 *data.Data
	P2 *data.Data
	P3 data.Data
	P4 Aa
	P5 *Aa
}

type B struct {
	P1 data.Data
	P2 *data.Data
	P3 *data.Data
	P4 *Bb
	P5 Bb
}

var X A
var Y B
