package conv

import (
	"bytes"
	"fmt"
)

type gener struct {
	buf    bytes.Buffer
	prefix string
}

const prefix_4Tab = "\t"

func newGener(prefix string) *gener {
	var buf bytes.Buffer
	return &gener{
		buf:    buf,
		prefix: prefix,
	}
}

func (g *gener) p(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, g.prefix+format, args...)
}

func (g *gener) string() string {
	return g.buf.String()
}
