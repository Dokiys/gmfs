package conv

import (
	"bytes"
	"fmt"
)

type gener struct {
	buf bytes.Buffer
}

func newGener() *gener {
	var buf bytes.Buffer
	return &gener{
		buf: buf,
	}
}

func (g *gener) p(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

func (g *gener) string() string {
	return g.buf.String()
}
