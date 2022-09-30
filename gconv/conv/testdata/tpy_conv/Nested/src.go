package tpyconv

type Common struct {
	C1 bool
}

type Aa struct {
	P1  bool
	P2  float32
	P3  float64
	P4  complex64
	P5  complex128
	P6  string
	P7  int
	P8  uint
	P9  uintptr
	P10 byte
	P11 rune
}

type Bb struct {
	P1  bool
	P2  float32
	P3  float64
	P4  complex64
	P5  complex128
	P6  string
	P7  int
	P8  uint
	P9  uintptr
	P10 byte
	P11 rune
}

type A struct {
	PC1 Common
	PC2 *Common
	PC3 Common
	PC4 *Common
	P1  Aa
	P2  *Aa
	P3  bool
	P4  float32
	P5  float64
	P6  complex64
	P7  complex128
	P8  string
	P9  int
	P10 uint
	P11 uintptr
	P12 byte
	P13 rune
}

type B struct {
	PC1 Common
	PC2 *Common
	PC3 *Common
	PC4 Common
	P1  *Bb
	P2  Bb
	P3  bool
	P4  float32
	P5  float64
	P6  complex64
	P7  complex128
	P8  string
	P9  int
	P10 uint
	P11 uintptr
	P12 byte
	P13 rune
}

var X A
var Y B
