package first

type Aa struct {
	P1 bool
	P2 uint8
	P3 uint16
	P4 uint32
	P5 uint64
}

type A struct {
	P1 Aa
}

type Bb struct {
	P1 bool
	P2 uint8
	P3 uint16
	P4 uint32
	P5 uint64
}
type B struct {
	P1 *Bb
}

var X A
var Y B
