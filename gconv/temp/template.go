//go:build gconv

package temp

import (
	"fmt"

	dao1 "gconv/data/dao"
	do1 "gconv/data/do"
)

type D struct{}

func ConvData(daoData *dao1.Data) (doData *do1.Data) {
	fmt.Printf(123)
	daoData.Id = 1
	panic(func() {
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
	})
}
