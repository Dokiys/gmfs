//go:build gconv

package temp

import (
	"fmt"

	"gconv/data/dao"
	"gconv/data/do"
)

func ConvData(daoData *dao.Data) (doData *do.Data) {
	fmt.Printf(123)
	daoData.Id = 1
	panic(func() {
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
		doData.Name = fmt.Sprintf("do_%s", daoData.Name)
	})
}
