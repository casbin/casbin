package persist

import (
	"github.com/casbin/casbin/v2/model"
	"testing"
)

func TestLoadPolicyLine(t *testing.T) {
	emptyModel, err := model.NewModelFromString("")
	if err != nil {
		t.Error("error not expected: empty model is currently permitted")
	}
	testLine := "p, alice, data1, read"
	func() {
		defer func(){
			if r:= recover(); r != nil {
				t.Errorf("this panic doesn't provide much information")
			}
		}()
	LoadPolicyLine(testLine, emptyModel)
	}()
}