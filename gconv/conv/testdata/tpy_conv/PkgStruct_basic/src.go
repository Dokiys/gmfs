package tpyconv

import (
	data2 "gconv/conv/testdata/tpy_conv/PkgStruct_basic/data"
)

type Aa struct {
	PP1 data2.Data
	PP2 *data2.Data
}

type Bb struct {
	PP1 *data2.Data
	PP2 data2.Data
}

type A struct {
	P1 *data2.Data
	P2 *data2.Data
	P3 data2.Data
	P4 Aa
	P5 *Aa
}

type B struct {
	P1 data2.Data
	P2 *data2.Data
	P3 *data2.Data
	P4 *Bb
	P5 Bb
}

var X A
var Y B
