package gmfs

import "testing"

func TestRegisterValidateFunc(t *testing.T) {
	RegisterValidateFunc(nil)
	if validate != nil {
		t.Fatal("validate should be nil")
	}
	RegisterValidateFunc(defValidateFunc)
}
