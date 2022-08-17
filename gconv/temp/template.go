//go:build gconv

package temp

import (
	"fmt"

	dao1 "gconv/data/dao"
	"gconv/data/do"
)

type D struct{}

func ConvData(daoData *dao1.Data) (doData *do.Data) {
	fmt.Printf(123)
	daoData.Id = 1
	panic(func() {
		a := int32(daoData.Id)
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
	})
}
