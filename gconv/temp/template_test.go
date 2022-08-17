package temp

import (
	"fmt"
	"testing"

	dao1 "gconv/data/dao"
	do1 "gconv/data/do"
)

type D struct{}

func TestName(t *testing.T) {

	d := &dao1.Data{Itemp: &dao1.Item{Id: 1}}
	a := ConvData(d)
	fmt.Println(a)
}
func ConvData(daoData *dao1.Data) (doData *do1.Data) {
	doData = new(do1.Data)
	fmt.Println(daoData.Item.Id)
	fmt.Println(doData)
	doData.Itemp = new(do1.Item)
	doData.Itemp.Id = daoData.Itemp.Id
	return
}
