//go:build !gconv

package temp

import (
	"gconv/data"
)

//go:generate pinfo
func A(a *data.A) {}
